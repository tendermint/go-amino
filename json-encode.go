package wire

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"unicode"
)

//----------------------------------------
// cdc.encodeReflectJSON

// *** Encoding/MarshalJSON ***
func (cdc *Codec) encodeReflectJSON(w io.Writer, info *TypeInfo, rv reflect.Value, opts FieldOptions) error {

	var disambiguate = info.Registered
	if disambiguate {
		// Write the disfix
		disfix := toDisfix(info.Disamb, info.Prefix)
		fmt.Fprintf(w, `{%s:"%X",%s:`, disfixKeyQuoted, disfix, dataKeyQuoted)
	}

	err := cdc._encodeReflectJSON(w, info, rv, opts)
	if err != nil {
		return err
	}

	// And finally if disambiguating, close the the disambiguation sequence.
	if disambiguate {
		_, err = w.Write(bytesCloseBrace)
	}
	return err
}

func (cdc *Codec) _encodeReflectJSON(w io.Writer, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {
	// 1. If we encounter the nil interface, encode null
	if rv.Kind() == reflect.Invalid {
		_, err = w.Write(bytesNull)
		return err
	}

	// 2a. If an object implements json.Marshaler
	// automatically respect that and skip to it.
	// before any dereferencing.
	if ok, err := processIfMarshaler(w, rv); ok || err != nil {
		return err
	}

	// 2b. Otherwise, dereference pointers
	var foundPointer = false
	// Dereference pointers all the way if any.
	// This works for pointer-pointers.
	for rv.Kind() == reflect.Ptr {
		foundPointer = true
		rv = rv.Elem()
	}

	if foundPointer {
		if !rv.IsValid() {
			_, err = w.Write(bytesNull)
			return
		}
	}

	// 2c. If the dereferenced object implements
	// json.Marshaler automatically respect that and skip to it.
	if ok, err := processIfMarshaler(w, rv); ok || err != nil {
		return err
	}

	switch info.Type.Kind() {
	//----------------------------------------
	// Complex types

	case reflect.Interface:
		return cdc.encodeReflectJSONInterface(w, info, rv, opts)

	case reflect.Array, reflect.Slice:
		return cdc.encodeReflectJSONArrayOrSlice(w, info, rv, opts)

	case reflect.Struct:
		return cdc.encodeReflectJSONStruct(w, info, rv, opts)

	case reflect.Map: // We explicitly don't support maps
		return errJSONMarshalMap

	case reflect.Float32, reflect.Float64:
		if !opts.Unsafe {
			return errors.New("Wire.JSON float* support requires `wire:\"unsafe\"`.")
		}
		return invokeStdlibJSONMarshal(w, rv.Interface())

	default: // All others
		return invokeStdlibJSONMarshal(w, rv.Interface())
	}
}

func (cdc *Codec) encodeReflectJSONArrayOrSlice(w io.Writer, info *TypeInfo, rv reflect.Value, opts FieldOptions) error {
	length := rv.Len()
	if length == 0 {
		_, err := w.Write(bytesOpenCloseLBraces)
		return err
	}

	if _, err := w.Write(bytesOpenLBrace); err != nil {
		return err
	}

	for i := 0; i < length; i++ {
		erv := rv.Index(i)
		ecrt := erv.Type() // non-pointer non-interface concrete type

		// Retrieve *TypeInfo for concrete type.
		einfo, err := cdc.getTypeInfo_wlock(ecrt)
		if err != nil {
			// TODO: However, we shouldn't really care for unregistered types
			// while performing JSON encoding, hence no check for error.
			return err
		}

		if err := cdc.encodeReflectJSON(w, einfo, erv, opts); err != nil {
			return err
		}
		// And then add the comma if it isn't the last item.
		if i != length-1 {
			if _, err := w.Write(bytesComma); err != nil {
				return err
			}
		}
	}
	_, err := w.Write(bytesCloseLBrace)
	return err
}

func (cdc *Codec) encodeReflectJSONInterface(w io.Writer, iinfo *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {
	if safeIsNil(rv) {
		_, err = w.Write(bytesNull)
		return err
	}

	// If the type implements json.Marshaler, just
	// automatically respect that and skip to it.
	if rv.Type().Implements(marshalerType) {
		var blob []byte
		blob, err = rv.Interface().(json.Marshaler).MarshalJSON()
		if err != nil {
			return
		}
		_, err = w.Write(blob)
		return
	}

	// Get concrete non-pointer reflect value & type.
	var crv = rv.Elem()
	crv, err = derefForInterface(crv, iinfo)
	if err != nil {
		return
	}
	var crt = crv.Type()

	// Retrieve *TypeInfo for concrete type.
	var cinfo *TypeInfo
	cinfo, err = cdc.getTypeInfo_wlock(crt)
	if err != nil {
		return
	}
	if !cinfo.Registered && false { // Hmm, primitive types would be a pain to complain about.
		err = fmt.Errorf("Cannot encode unregistered concrete type %v.", crt)
		return
	}
	err = cdc.encodeReflectJSON(w, cinfo, crv, opts)
	return
}

func (cdc *Codec) encodeReflectJSONStruct(w io.Writer, info *TypeInfo, rv reflect.Value, opts FieldOptions) error {
	nf := len(info.Fields)
	if nf == 0 {
		_, err := w.Write(bytesOpenCloseBraces)
		return err
	}
	typ := rv.Type()
	bytesKVPairs := make([][]byte, 0, nf)
	for _, field := range info.Fields {
		typField := typ.Field(field.Index)
		if !isExported(typField.Name) {
			continue
		}
		valField := rv.Field(field.Index)
		fieldKey, omitEmpty := jsonFieldKey(typField)
		if omitEmpty && isBlank(valField) {
			continue
		}
		info, err := cdc.getTypeInfo_wlock(field.Type)
		if err != nil {
			return err
		}
		valuesBuf := new(bytes.Buffer)
		if err := cdc.encodeReflectJSON(valuesBuf, info, valField, opts); err != nil {
			return err
		}

		valueBlob := valuesBuf.Bytes()
		if omitEmpty && allBlankBytes(valueBlob) {
			continue
		}

		keyBlob := bytes.Join([][]byte{bytesQuote, []byte(fieldKey), bytesQuote}, bytesBlankString)
		kvBlob := bytes.Join([][]byte{keyBlob, valueBlob}, bytesColon)
		bytesKVPairs = append(bytesKVPairs, kvBlob)
	}

	if len(bytesKVPairs) == 0 {
		_, err := w.Write(bytesOpenCloseBraces)
		return err
	}
	if _, err := w.Write(bytesOpenBrace); err != nil {
		return err
	}
	joinedKVBytes := bytes.Join(bytesKVPairs, bytesComma)
	if _, err := w.Write(joinedKVBytes); err != nil {
		return err
	}
	_, err := w.Write(bytesCloseBrace)
	return err
}

func isExported(nm string) bool {
	return len(nm) > 0 && unicode.IsUpper(rune(nm[0]))
}

var (
	bytesNull             = []byte("null")
	bytesColon            = []byte(":")
	bytesComma            = []byte(",")
	bytesOpenBrace        = []byte("{")
	bytesCloseBrace       = []byte("}")
	byteOpenLBrace        = byte('[')
	bytesOpenLBrace       = []byte{byteOpenLBrace}
	byteCloseLBrace       = byte(']')
	bytesCloseLBrace      = []byte{byteCloseLBrace}
	bytesQuote            = []byte("\"")
	bytesBlankString      = []byte("")
	bytesOpenCloseBraces  = []byte("{}")
	bytesOpenCloseLBraces = []byte("[]")
	bytesZero             = []byte("0")
	bytesFalse            = []byte("false")
)

func jsonFieldKey(f reflect.StructField) (string, bool) {
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

// XXX remove?
func trimWhitespace(bz []byte) []byte {
	for i, b := range bz {
		if b == ' ' || b == '\n' || b == '\t' || b == '\r' {
			continue
		}
	}
}

func nilBytes(b []byte) bool {
	return bytes.Equal(b, bytesNull)
}

func allBlankBytes(b []byte) bool {
	return len(b) == 0 ||
		bytes.Equal(b, nil) ||
		bytes.Equal(b, bytesBlankString) ||
		bytes.Equal(b, bytesZero) ||
		bytes.Equal(b, bytesFalse)
}

// safeIsNil safely invokes reflect.Value.IsNil only on
// * reflect.Interface, reflect.Ptr, reflect.Map, reflect.Slice,
// * reflect.Array, reflect.Ptr, reflect.Func
// otherwise it returns false
func safeIsNil(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Interface, reflect.Chan, reflect.Map, reflect.Slice,
		reflect.Array, reflect.Ptr, reflect.Func:
		return v.IsNil()
	default:
		return false
	}
}

func isBlank(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Interface, reflect.Chan, reflect.Map, reflect.Slice,
		reflect.Array, reflect.Ptr, reflect.Func:
		return v.IsNil()
	default:
		return isBlankInterface(v.Interface())
	}
}

func isBlankInterface(v interface{}) bool {
	switch v {
	case 0, "", false, nil:
		// Obviously these work for only untyped constants but
		// nonetheless an attempt at finding zero values
		return true
	default:
		// Not much we can do
		return false
	}
}

var (
	errJSONMarshalMap     = errors.New("maps are not supported")
	errExpectingOpenBrace = errors.New("expecting '{'")
)

const (
	disfixKeyQuoted = `"_df"`
	dataKeyQuoted   = `"_v"`
)

func invokeJSONMarshaler(w io.Writer, rv reflect.Value) error {
	blob, err := rv.Interface().(json.Marshaler).MarshalJSON()
	if err != nil {
		return err
	}
	_, err = w.Write(blob)
	return err
}

// processIfMarshaler checks if the type or its pointer
// to implements json.Marshaler and if so, invokes it.
func processIfMarshaler(w io.Writer, rv reflect.Value) (marshalable bool, err error) {
	if rv.Type().Implements(marshalerType) {
		return true, invokeJSONMarshaler(w, rv)
	}

	// Otherwise if its pointer implements
	// json.Marshaler, try that too.
	if itsPtrImplements(rv, marshalerType) {
		return true, invokeJSONMarshaler(w, rv.Addr())
	}

	return false, nil
}

func invokeStdlibJSONMarshal(w io.Writer, v interface{}) error {
	// Note: Please don't stream out the output because that adds a newline
	// using json.NewEncoder(w).Encode(data)
	// as per https://golang.org/pkg/encoding/json/#Encoder.Encode
	blob, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = w.Write(blob)
	return err
}
