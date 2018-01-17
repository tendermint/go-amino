package wire

import (
	"bytes"
	"encoding/hex"
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

const (
	disfixKeyQuoted = `"_df"`
	dataKeyQuoted   = `"_v"`
)

// *** Encoding/MarshalJSON ***
func (cdc *Codec) encodeReflectJSON(w io.Writer, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {
	disambiguate := info.Registered || info.AlwaysDisambiguate
	if disambiguate {
		// Write the disfix
		disfix := toDisfix(info.Disamb, info.Prefix)
		fmt.Fprintf(w, `{%s:"%X",%s:`, disfixKeyQuoted, disfix, dataKeyQuoted)
	}

	if err = cdc._encodeReflectJSON(w, info, rv, opts); err != nil {
		return
	}
	if disambiguate {
		_, err = w.Write(bytesCloseBrace)
	}
	return err
}

func (cdc *Codec) _encodeReflectJSON(w io.Writer, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {
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

func (cdc *Codec) encodeReflectJSONArrayOrSlice(w io.Writer, info *TypeInfo, rv reflect.Value, opts FieldOptions) error {
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

		// Retrieve *TypeInfo for concrete type.
		einfo, err := cdc.getTypeInfo_wlock(ecrt)
		if err != nil {
			// TODO: However, we shouldn't really care for unregistered types
			// while performing JSON encoding, hence no check for error.
			return err
		}

		elemsBuf := new(bytes.Buffer)
		if err := cdc.encodeReflectJSONInterface(elemsBuf, einfo, erv, opts); err != nil {
			return err
		}
		eBlob := elemsBuf.Bytes()
		kvBytes = append(kvBytes, eBlob)
	}

	// Now insert a comma in-between every two key-value pairs.
	if _, err := w.Write(bytes.Join(kvBytes, bytesComma)); err != nil {
		return err
	}
	_, err := w.Write(bytesCloseLBrace)
	return err
}

func safeElem(v reflect.Value) reflect.Value {
	// As per https://golang.org/pkg/reflect/#Value.Elem
	// Elem can only be invoked on an interface or a pointer.
	if v.Kind() == reflect.Interface || v.Kind() == reflect.Ptr {
		return v.Elem()
	}
	return v
}

func (cdc *Codec) encodeReflectJSONInterface(w io.Writer, info *TypeInfo, rv reflect.Value, opts FieldOptions) error {
	if safeIsNil(rv) {
		_, err := w.Write(bytesNull)
		return err
	}

	// Concrete reflect value
	crv := safeElem(rv)

	crv, err := deref(crv, info)
	if err != nil {
		return err
	}
	crt := crv.Type()

	// Retrieve *TypeInfo for concrete type.
	cinfo, err := cdc.getTypeInfo_wlock(crt)
	if err != nil {
		return err
	}
	if !cinfo.Registered && false { // Hmm, primitive types would be a pain to complain about.
		return fmt.Errorf("Cannot encode unregistered concrete type %v.", crt)
	}
	return cdc.encodeReflectJSON(w, cinfo, crv, opts)
}

func (cdc *Codec) encodeReflectJSONStruct(w io.Writer, info *TypeInfo, rv reflect.Value, opts FieldOptions) error {
	// MARK:
	typ := rv.Type()
	nf := typ.NumField()
	if nf == 0 {
		_, err := w.Write(bytesOpenCloseBraces)
		return err
	}
	bytesKVPairs := make([][]byte, 0, nf)
	for i := 0; i < nf; i++ {
		typField := typ.Field(i)
		if !isExported(typField) {
			continue
		}
		valField := rv.Field(i)
		fieldKey, omitEmpty := jsonFieldKey(typField)
		if omitEmpty && isBlank(valField) {
			continue
		}
		info, err := cdc.getTypeInfo_wlock(typField.Type)
		if err != nil {
			return err
		}
		valuesBuf := new(bytes.Buffer)
		if err := cdc.encodeReflectJSONInterface(valuesBuf, info, valField, opts); err != nil {
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

func isExported(st reflect.StructField) bool {
	nm := st.Name
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

func nilBytes(b []byte) bool {
	return len(b) == 0 || bytes.Equal(b, bytesNull)
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

// blobSaver is a workaround to save a blob when parsing
// unknown bytes in mixed types such as if we have
//    `{"c": 0, "d": "foo", "e": {"k": "bar"}}`
// in the above blob, if we want to just check the
// keys but retain the bytes without having to first unmarshal
// to map[string]interface{}, and then marshal back
// in order to get the respective keys' blobs.
type blobSaver struct {
	blob []byte
}

func (ab *blobSaver) UnmarshalJSON(b []byte) error {
	ab.blob = b
	return nil
}

type disfixRepr struct {
	Disfix string     `json:"_df"`
	Data   *blobSaver `json:"_v"`
}

var blankDisfix DisfixBytes

// parseDisfixAndData helps unravel the disfix and
// the stored data, which are expected in the form:
// {
//    "_df": "XXXXXXXXXXXXXXXXX",
//    "_v":  {}
// }
func parseDisfixAndData(bz []byte) (disfix DisfixBytes, dataBytes []byte, err error) {
	bz = bytes.TrimSpace(bz)
	if len(bz) < DisfixBytesLen {
		return disfix, bz, errors.New("parseDisfixAndData: EOF skipping prefix bytes.")
	}
	dfr := new(disfixRepr)
	if err := json.Unmarshal(bz, dfr); err != nil {
		return disfix, bz, fmt.Errorf("Parsing Disfix and data: %v", err)
	}
	hexBytes, err := hex.DecodeString(dfr.Disfix)
	if err != nil {
		return disfix, bz, err
	}
	if g, w := len(hexBytes), DisfixBytesLen; g != w {
		return disfix, bz, fmt.Errorf("Disfix length got=%d want=%d", g, w)
	}
	copy(disfix[:], hexBytes)
	if bytes.Equal(disfix[:], blankDisfix[:]) {
		return disfix, bz, errors.New("expected a non-blank disfix")
	}
	if blobSaver := dfr.Data; blobSaver != nil {
		dataBytes = blobSaver.blob
	}
	return disfix, dataBytes, err
}

//----------------------------------------
// cdc.decodeReflectJSON

// CONTRACT: rv.CanAddr() is true.
func (cdc *Codec) decodeReflectJSON(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) error {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}

	disambiguation := info.Registered || info.AlwaysDisambiguate
	orv := rv
	if disambiguation {
		// Expecting
		//  {"_df":"<DF>","_v":"data"}
		if len(bz) < DisfixBytesLen {
			return errors.New("EOF skipping prefix bytes.")
		}
		gotDisfix, bzRest, err := parseDisfixAndData(bz)
		if err != nil {
			return err
		}
		info, err = cdc.getTypeInfoFromDisfix_rlock(gotDisfix)
		if err != nil {
			return err
		}
		// Ensure that the concrete type matches the disfix
		wantDisfix := toDisfix(info.Disamb, info.Prefix)
		if g, w := gotDisfix[:], wantDisfix[:]; !bytes.Equal(g, w) {
			return fmt.Errorf("decodeReflectJSON: disfix mismatch got=%X want=%X", g, w)
		}
		bz = bzRest
		crv, _ := constructConcreteType(info)
		rv = crv
	}

	if err := cdc._decodeReflectJSON(bz, info, rv, opts); err != nil {
		return err
	}
	if disambiguation {
		orv.Set(rv)
	}
	return nil
}

func (cdc *Codec) _decodeReflectJSON(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) error {
	// SANITY CHECK
	if info.Type.Kind() == reflect.Interface && rv.Kind() == reflect.Ptr {
		panic("should not happen")
	}

	// Special case for nil
	if nilBytes(bz) {
		rv.Set(info.ZeroValue)
		return nil
	}

	// Handle pointer types.
	if rv.Kind() == reflect.Ptr {
		// Dereference-and-construct pointers all the way.
		// This works for pointer-pointers.
		for c := true; c; c = rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				newPtr := reflect.New(rv.Type().Elem())
				rv.Set(newPtr)
			}
			rv = rv.Elem()
		}
	}

	switch info.Type.Kind() {

	//----------------------------------------
	// Complex

	case reflect.Interface:
		return cdc.decodeReflectJSONInterface(bz, info, rv, opts)

	case reflect.Array, reflect.Slice:
		return cdc.decodeReflectJSONArrayOrSlice(bz, info, rv, opts)

	case reflect.Struct:
		return cdc.decodeReflectJSONStruct(bz, info, rv, opts)

	//----------------------------------------

	case reflect.Float32, reflect.Float64:
		if !opts.Unsafe {
			return errors.New("Wire.JSON float* support requires `wire:\"unsafe\"`.")
		}
		return invokeStdlibJSONUnmarshal(bz, info, rv, opts)

	case reflect.Map: // We explicitly don't support maps
		return errJSONMarshalMap

	default: // All others
		return invokeStdlibJSONUnmarshal(bz, info, rv, opts)
	}
}

func invokeStdlibJSONUnmarshal(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) error {
	if !rv.CanAddr() && rv.Kind() != reflect.Ptr {
		panic("rv not addressable nor pointer")
	}

	var rrv reflect.Value = rv
	if rv.Kind() != reflect.Ptr {
		rrv = reflect.New(rv.Type())
	}
	if err := json.Unmarshal(bz, rrv.Interface()); err != nil {
		return err
	}
	rv.Set(rrv.Elem())
	return nil
}

// *** Decoding/UnmarshalJSON ***
func (cdc *Codec) decodeReflectJSONInterface(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) error {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}

	// Special case for nil
	if nilBytes(bz) {
		rv.Set(info.ZeroValue)
		return nil
	}

	// Concrete reflect value
	crv := safeElem(rv)

	crv, err := deref(crv, info)
	if err != nil {
		return err
	}

	// Retrieve concrete type info.
	var cinfo *TypeInfo
	if info.Registered || info.AlwaysDisambiguate {
		disfix, restBlob, err := parseDisfixAndData(bz)
		if err != nil {
			return err
		}
		bz = restBlob
		cinfo, err = cdc.getTypeInfoFromDisfix_rlock(disfix)
	} else {
		cinfo = info
	}

	if err != nil {
		return err
	}

	crv, rvSet := constructConcreteType(cinfo)
	// Read into crv.
	if err := cdc.decodeReflectJSON(bz, cinfo, crv, opts); err != nil {
		return err
	}

	// And finally set rv.
	rv.Set(rvSet)
	return nil
}

// constructConcreteType creates the concrete value as
// well as the corresponding settable value for it.
func constructConcreteType(cinfo *TypeInfo) (crv, rvSet reflect.Value) {
	// Construct new concrete type.
	// NOTE: rv.Set() should succeed because it was validated
	// already during Register[Interface/Concrete].
	if cinfo.PointerPreferred {
		cPtrRv := reflect.New(cinfo.Type)
		crv = cPtrRv.Elem()
	} else {
		crv = reflect.New(cinfo.Type).Elem()
		rvSet = crv
	}
	return crv, rvSet
}

func (cdc *Codec) decodeReflectJSONArrayOrSlice(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) error {
	bz = bytes.TrimSpace(bz)
	if nilBytes(bz) {
		return nil
	}

	innerTyp := rv.Type().Elem()
	info, err := cdc.getTypeInfo_wlock(innerTyp)
	if err != nil {
		return err
	}

	// First things first, basic validation
	if g, w := bz[0], byteOpenLBrace; g != w {
		return fmt.Errorf("decodeReflectJSONArrayOrSlice: got %c want %c bz: %s", g, w, bz)
	}
	if g, w := bz[len(bz)-1], byteCloseLBrace; g != w {
		return fmt.Errorf("decodeReflectJSONArrayOrSlice: got %c want %c bz: %s", g, w, bz)
	}

	var blobHolder []*blobSaver
	if err := json.Unmarshal(bz, &blobHolder); err != nil {
		return err
	}
	outSlice := reflect.MakeSlice(rv.Type(), 0, len(blobHolder))
	for _, bh := range blobHolder {
		ithElemPtr := reflect.New(innerTyp)
		if bh == nil {
			continue
		}
		ithElem := ithElemPtr.Elem()
		if err := cdc.decodeReflectJSON(bh.blob, info, ithElem, opts); err != nil {
			return err
		}
		outSlice = reflect.Append(outSlice, ithElem)
	}
	rv.Set(outSlice)
	return nil
}

func (cdc *Codec) decodeReflectJSONStruct(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) error {
	typ := rv.Type()
	nf := typ.NumField()
	if nf == 0 {
		return nil
	}

	// Map all the fields(keys) to their blobs/bytes.
	fieldsToByteValuesMap := make(map[string]*blobSaver)
	if err := json.Unmarshal(bz, &fieldsToByteValuesMap); err != nil {
		return err
	}

	for i := 0; i < nf; i++ {
		typField := typ.Field(i)
		if !isExported(typField) {
			continue
		}

		valField := rv.Field(i)
		fieldKey, _ := jsonFieldKey(typField)
		blobSave, ok := fieldsToByteValuesMap[fieldKey]
		if !ok {
			// TODO: Since the Go stdlib's JSON codec allows case-insensitive
			// keys perhaps we need to also do case-insensitive lookups here.
			// So "Vanilla" and "vanilla" would both match to the same field.
			// It is actually a security flaw with encoding/json library
			//  See https://github.com/golang/go/issues/14750
			// but perhaps we are aiming for as much compatibility here.
			continue
		}
		if blobSave == nil {
			continue
		}

		// Now let's look up this field's type information.
		finfo, err := cdc.getTypeInfo_wlock(typField.Type)
		if err != nil {
			return err
		}
		if err := cdc.decodeReflectJSON(blobSave.blob, finfo, valField, opts); err != nil {
			return err
		}
	}

	return nil
}
