/*
The purpose of the JSON marshalling and unmarshalling code is to
ensure that any interface's concrete types are encoded with disambiguation
in order to extract out the concrete type when decoding
For example given
type Transport struct {
    Vehicle
    Capacity int
}

type Vehicle interface {
    Move() error
}

type Car   string
type Boat  int
type Plane int

func (c Car) Move() error { return nil }
func (b Boat) Move() error { return nil }
func (p Plane) Move() error { return nil }

wire.RegisterConcrete(&Transport{}, "our/transport", nil)
wire.RegisterConcrete(&Car("Acura"), "car", nil)
wire.MarshalJSON(&Transport{Car{}, 10})
should give
{
  "_df": "XXXXXXXXXXXXXX",
  "Vehicle":  "Acura",
  "Capacity": 10
}

which will strictly only unmarshal to:
Transport{Car("Acura"), Capacity: 10}
given:

tr := new(Transport)
wire.UnmarshalJSON(blob, tr)
*/
package wire

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"unicode"
)

func (cdc *Codec) MarshalJSON(o interface{}) ([]byte, error) {
	// reflect.ValueOf retrieves the concrete type
	// that the interface contains.
	val := reflect.ValueOf(o)
	origVal := val

	switch val.Kind() {
	case reflect.Invalid: // They passed in: an untyped nil, marshal it to "null"
		return []byte("null"), nil
	case reflect.Ptr:
		if val.IsNil() {
			return []byte("null"), nil
		}

		// Perfect we've got a pointer
		// Dereference it to the non-pointer type
		val = deref(val)
	}

	// If the type implements json.MarshalJSON, then defer to its MarshalJSON method.
	if anyImplementsJSONMarshaler(val, origVal) {
		return json.Marshal(o)
	}

	// Except for:
	//  * maps
	//  * structs
	//  * arrays and slices
	// all other types should be encoded as they are.
	switch val.Kind() {
	case reflect.Map:
		buf := new(bytes.Buffer)
		if err := cdc.marshalMap(val, buf); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil

	case reflect.Slice, reflect.Array:
		buf := new(bytes.Buffer)
		if err := cdc.marshalSliceOrArray(val, buf); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil

	case reflect.Struct:
		// For each field, invoke MarshalJSON on it
		buf := new(bytes.Buffer)
		if err := cdc.marshalStruct(val, buf); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil

	default:
		return json.Marshal(val.Interface())
	}
}

var jsonMarshaler = reflect.TypeOf(new(json.Marshaler)).Elem()

func anyImplementsJSONMarshaler(vals ...reflect.Value) bool {
	for _, val := range vals {
		if typ := val.Type(); typ.Implements(jsonMarshaler) {
			return true
		}
	}
	return false
}

var (
	bytesColon            = []byte(":")
	bytesComma            = []byte(",")
	bytesOpenBrace        = []byte("{")
	bytesCloseBrace       = []byte("}")
	bytesOpenLBrace       = []byte("[")
	bytesCloseLBrace      = []byte("]")
	bytesQuote            = []byte("\"")
	bytesBlankString      = []byte("")
	bytesOpenCloseBraces  = []byte("{}")
	bytesOpenCloseLBraces = []byte("[]")
	bytesZero             = []byte("0")
	bytesFalse            = []byte("false")
)

func (cdc *Codec) marshalSliceOrArray(m reflect.Value, out io.Writer) error {
	nl := m.Len()
	if nl == 0 {
		_, err := out.Write(bytesOpenCloseLBraces)
		return err
	}

	bytesKVPairs := make([][]byte, 0, nl)
	for i := 0; i < nl; i++ {
		val := m.Index(i)
		valBlob, err := cdc.MarshalJSON(val.Interface())
		if err != nil {
			return err
		}
		bytesKVPairs = append(bytesKVPairs, valBlob)
	}
	if _, err := out.Write(bytesOpenLBrace); err != nil {
		return err
	}
	if _, err := out.Write(bytes.Join(bytesKVPairs, bytesComma)); err != nil {
		return err
	}
	_, err := out.Write(bytesCloseLBrace)
	return err
}

func (cdc *Codec) marshalMap(m reflect.Value, out io.Writer) error {
	if m.Len() == 0 {
		_, err := out.Write(bytesOpenCloseBraces)
		return err
	}
	keys := m.MapKeys()
	bytesKVPairs := make([][]byte, 0, len(keys))
	for _, key := range keys {
		keyBlob, err := cdc.MarshalJSON(key.Interface())
		if err != nil {
			return err
		}
		value := m.MapIndex(key)
		valueBlob, err := cdc.MarshalJSON(value.Interface())
		if err != nil {
			return err
		}
		kvBlob := bytes.Join([][]byte{keyBlob, valueBlob}, bytesColon)
		bytesKVPairs = append(bytesKVPairs, kvBlob)
	}
	return writeOutBlobPairs(out, bytesKVPairs)
}

func allBlankBytes(b []byte) bool {
	return len(b) == 0 ||
		bytes.Equal(b, nil) ||
		bytes.Equal(b, bytesBlankString) ||
		bytes.Equal(b, bytesZero) ||
		bytes.Equal(b, bytesFalse)
}

func isBlank(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Chan, reflect.Map, reflect.Slice, reflect.Array, reflect.Ptr, reflect.Func:
		return v.IsNil()
	default:
		return isBlankInterface(v.Interface())
	}
}

func isBlankInterface(v interface{}) bool {
	switch v {
	case 0, "", false, nil:
		// Obviously these work for untyped values but nonetheless an
		// attempt at finding zero values
		return true
	default:
		// Not much we can do
		return false
	}
}

func (cdc *Codec) marshalStruct(sv reflect.Value, out io.Writer) error {
	typ := sv.Type()
	nf := typ.NumField()
	if nf == 0 {
		_, err := out.Write(bytesOpenCloseBraces)
		return err
	}
	bytesKVPairs := make([][]byte, 0, nf)
	for i := 0; i < nf; i++ {
		typField := typ.Field(i)
		if !isExported(typField) {
			continue
		}
		valField := sv.Field(i)
		fieldName, omitEmpty := jsonFieldName(typField)
		if omitEmpty && isBlank(valField) {
			continue
		}
		valueBlob, err := cdc.MarshalJSON(valField.Interface())
		if err != nil {
			return err
		}
		if omitEmpty && allBlankBytes(valueBlob) {
			continue
		}
		keyBlob := bytes.Join([][]byte{bytesQuote, []byte(fieldName), bytesQuote}, bytesBlankString)

		// If it is an interface type, we need to set the disambiguation bytes
		// so that decoding later on will be able to map to the right concrete type.
		if typField.Type.Kind() == reflect.Interface {
			info, err := cdc.getTypeInfo_wlock(typField.Type)
			if err == nil {
				disfix := info.disfix()
				// We'll need to transform the normal blob to {"_dis":0xFF,"data":...}
				valueBlob = bytes.Join([][]byte{
					bytesOpenBrace,
					[]byte(fmt.Sprintf(`"_df":"%x",`, disfix)),
					[]byte(`"_data":`), valueBlob,
					bytesCloseBrace,
				}, bytesBlankString)
			}
		}
		kvBlob := bytes.Join([][]byte{keyBlob, valueBlob}, bytesColon)
		bytesKVPairs = append(bytesKVPairs, kvBlob)
	}
	return writeOutBlobPairs(out, bytesKVPairs)
}

func isExported(st reflect.StructField) bool {
	nm := st.Name
	return len(nm) > 0 && unicode.IsUpper(rune(nm[0]))
}

func writeOutBlobPairs(out io.Writer, bytesKVPairs [][]byte) error {
	if len(bytesKVPairs) == 0 {
		_, err := out.Write(bytesOpenCloseBraces)
		return err
	}
	// Now we need to create "{" <kvPairs...> "}"
	if _, err := out.Write(bytesOpenBrace); err != nil {
		return err
	}
	if _, err := out.Write(bytes.Join(bytesKVPairs, bytesComma)); err != nil {
		return err
	}
	_, err := out.Write(bytesCloseBrace)
	return err
}

func jsonFieldName(f reflect.StructField) (string, bool) {
	jsonTagName, ok := f.Tag.Lookup("json")
	if !ok {
		return f.Name, false
	}
	// Otherwise we need to figure out which name to use.
	splits := strings.Split(jsonTagName, ",")
	omitEmpty := strings.Contains(jsonTagName, ",omitempty")
	head := splits[0]
	if head == "" {
		head = f.Name
	}
	return head, omitEmpty
}

func deref(val reflect.Value) reflect.Value {
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return val
	}
	elem := val.Elem()
	for elem.Kind() == reflect.Ptr {
		elem = elem.Elem()
	}
	return elem
}

func (cdc *Codec) UnmarshalJSON(bz []byte, ptr interface{}) error {
	panic("not implemented yet") // XXX
}

func (cdc *Codec) UnmarshalJSONLengthPrefixed(bz []byte, ptr interface{}) error {
	panic("not implemented yet") // XXX
}
