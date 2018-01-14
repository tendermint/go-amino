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

func (cdc *Codec) encodeReflectJSON(w io.Writer, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {
	// Dereference pointers all the way if any.
	// This works for pointer-pointers.
	var foundPointer = false
	for rv.Kind() == reflect.Ptr {
		foundPointer = true
		rv = rv.Elem()
	}

	if foundPointer {
		if !rv.IsValid() {
			_, err = w.Write(bytesNull)
			return err
		}
	}

	switch info.Type.Kind() {
	//----------------------------------------
	// Complex types

	case reflect.Interface:
		return cdc.encodeReflectJSONInterface(w, info, rv, opts)

	case reflect.Array, reflect.Slice:
		return cdc.encodeReflectJSONSliceOrArray(w, info, rv, opts)

	case reflect.Struct:
		return cdc.encodeReflectJSONStruct(w, info, rv, opts)

	case reflect.Map: // We explicitly don't support maps
		return errJSONMarshalMap

	default: // All others
		// Note: Please don't stream out the output because that adds a newline
		// using json.NewEncoder(w).Encode(data)
		// as per https://golang.org/pkg/encoding/json/#Encoder.Encode
		blob, err := json.Marshal(rv.Interface())
		if err != nil {
			return err
		}
		_, err = w.Write(blob)
		return err
	}
}

func (cdc *Codec) encodeReflectJSONSliceOrArray(w io.Writer, info *TypeInfo, rv reflect.Value, opts FieldOptions) error {
	length := rv.Len()
	if length == 0 {
		_, err := w.Write(bytesOpenCloseLBraces)
		return err
	}

	if _, err := w.Write(bytesOpenLBrace); err != nil {
		return err
	}
	kvBytes := make([][]byte, 0, length)
	for i := 0; i < length; i++ {
		erv := rv.Index(i)
		ecrt := erv.Type() // non-pointer non-interface concrete type

		// Get *TypeInfo for concrete type.
		// However, we don't really care for unregistered types
		// while performing JSON encoding.
		einfo, _ := cdc.getTypeInfo_wlock(ecrt)

		elemsBuf := new(bytes.Buffer)
		if err := cdc.encodeReflectJSON(elemsBuf, einfo, erv, opts); err != nil {
			return err
		}

		eBlob := elemsBuf.Bytes()
		// Re-use the valuesBuf
		elemsBuf.Reset()

		kvBytes = append(kvBytes, eBlob)
	}
	if _, err := w.Write(bytes.Join(kvBytes, bytesComma)); err != nil {
		return err
	}
	_, err := w.Write(bytesCloseLBrace)
	return err
}

func (cdc *Codec) encodeReflectJSONInterface(w io.Writer, info *TypeInfo, rv reflect.Value, opts FieldOptions) error {
	crv, err := deref(rv, info)
	if err != nil {
		return err
	}
	crt := crv.Type() // non-pointer non-interface concrete type

	// Get *TypeInfo for concrete type.
	cinfo, err := cdc.getTypeInfo_wlock(crt)
	if err != nil {
		return err
	}

	// Write the disambiguation bytes if needed.
	var needDisamb bool = false
	if info.AlwaysDisambiguate {
		needDisamb = true
	} else if len(info.Implementers[cinfo.Prefix]) > 1 {
		needDisamb = true
	}
	if !needDisamb {
		return cdc.encodeReflectJSON(w, cinfo, crv, opts)
	}

	// Otherwise, let's encode the disambiguation bytes
	fmt.Fprintf(w, `{"_df":"%x","_v":`, cinfo.Disamb)
	if err := cdc.encodeReflectJSON(w, cinfo, crv, opts); err != nil {
		return err
	}
	_, err = w.Write(bytesCloseBrace)
	return err
}

func (cdc *Codec) encodeReflectJSONStruct(w io.Writer, info *TypeInfo, rv reflect.Value, opts FieldOptions) error {
	typ := rv.Type()
	nf := typ.NumField()
	if nf == 0 {
		_, err := w.Write(bytesOpenCloseBraces)
		return err
	}
	valuesBuf := new(bytes.Buffer)
	bytesKVPairs := make([][]byte, 0, nf)
	for i := 0; i < nf; i++ {
		typField := typ.Field(i)
		if !isExported(typField) {
			continue
		}
		valField := rv.Field(i)
		fieldName, omitEmpty := jsonFieldName(typField)
		if omitEmpty && isBlank(valField) {
			continue
		}
		info, err := cdc.getTypeInfo_wlock(typField.Type)
		if err != nil {
			return err
		}
		if err := cdc.encodeReflectJSONInterface(valuesBuf, info, valField, opts); err != nil {
			return err
		}

		valueBlob := valuesBuf.Bytes()
		// Re-use the valuesBuf
		valuesBuf.Reset()
		if omitEmpty && allBlankBytes(valueBlob) {
			continue
		}

		keyBlob := bytes.Join([][]byte{bytesQuote, []byte(fieldName), bytesQuote}, bytesBlankString)
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

func isExported(st reflect.StructField) bool {
	nm := st.Name
	return len(nm) > 0 && unicode.IsUpper(rune(nm[0]))
}

var errJSONMarshalMap = errors.New("maps are unsupported")

var (
	bytesNull             = []byte("null")
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
