package wire

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

//----------------------------------------
// cdc.decodeReflectJSON

// CONTRACT: rv.CanAddr() is true.
func (cdc *Codec) decodeReflectJSON(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}
	if info.Type.Kind() == reflect.Interface && rv.Kind() == reflect.Ptr {
		panic("should not happen")
	}

	// No need for disambiguation, decode as is.
	if !info.Registered {
		err = cdc._decodeReflectJSON(bz, info, rv, opts)
		return
	}

	// Otherwise, disambiguation time.
	disfix, bz, err := decodeDisfixJSON(bz)
	if err != nil {
		return
	}
	cinfo, err := cdc.getTypeInfoFromDisfix_rlock(disfix)
	if err != nil {
		return
	}

	// And we need to construct the concrete type
	// that'll then be set into the interface field.
	var crv, _ = constructConcreteType(cinfo)
	if err = cdc._decodeReflectJSON(bz, info, crv, opts); err != nil {
		rv.Set(crv) // Helps with debugging
		return
	}

	// We need to set here, for when !PointerPreferred and the type
	// is say, an array of bytes (e.g. [32]byte), then we must call
	// rv.Set() *after* the value was acquired.
	rv.Set(crv)
	return
}

func (cdc *Codec) _decodeReflectJSON(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) error {

	// If the type implements json.Unmarshaler, just
	// automatically respect that and skip to it.
	if rv.Addr().Type().Implements(unmarshalerType) {
		return rv.Addr().Interface().(json.Unmarshaler).UnmarshalJSON(bz)
	}

	// Special case for nil for either interface, pointer, slice
	if nilBytes(bz) {
		switch rv.Kind() {
		case reflect.Interface, reflect.Ptr, reflect.Slice, reflect.Array:
			rv.Set(info.ZeroValue)
			return nil
		}
	}

	// Ensure that any pointer field that's
	// nil is constructed, but also dereference
	// until the non-pointer type.
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			newPtr := reflect.New(rv.Type().Elem())
			rv.Set(newPtr)
		}
		rv = rv.Elem()
	}

	switch ikind := info.Type.Kind(); ikind {

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

	case reflect.Map, reflect.Func, reflect.Chan: // We explicitly don't support maps, funcs or channels
		return fmt.Errorf("unsupported kind: %s", ikind)

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

// CONTRACT: rv.CanAddr() is true.
func (cdc *Codec) decodeReflectJSONInterface(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) error {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}

	// Always drill down and grab its concrete
	// type information through disambiguation.
	disfix, bz, err := decodeDisfixJSON(bz)
	if err != nil {
		return err
	}

	info, err = cdc.getTypeInfoFromDisfix_rlock(disfix)
	if err != nil {
		return err
	}

	// Create the concrete type since we are dealing with an
	// interface that we have just disambiguated from above.
	cPtr := reflect.New(info.Type)
	crv := cPtr.Elem()
	if err := cdc._decodeReflectJSON(bz, info, crv, opts); err != nil {
		return err
	}

	// Now the interface has a concrete type set to it!
	rv.Set(crv)
	return nil
}

// CONTRACT: rv.CanAddr() is true.
func (cdc *Codec) decodeReflectJSONArrayOrSlice(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) error {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}

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
	nf := len(info.Fields)
	if nf == 0 {
		return nil
	}

	// Map all the fields(keys) to their blobs/bytes.
	fieldsToByteValuesMap := make(map[string]*blobSaver)
	if err := json.Unmarshal(bz, &fieldsToByteValuesMap); err != nil {
		return err
	}

	typ := rv.Type()
	for _, field := range info.Fields {
		typField := typ.Field(field.Index)
		if !isExported(typField.Name) {
			continue
		}

		fieldKey := field.JSONName
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
		finfo, err := cdc.getTypeInfo_wlock(field.Type)
		if err != nil {
			return err
		}
		valField := rv.Field(field.Index)
		if err := cdc.decodeReflectJSON(blobSave.blob, finfo, valField, opts); err != nil {
			return err
		}
	}

	return nil
}

func itsPtrImplements(v reflect.Value, check reflect.Type) bool {
	return v.Kind() != reflect.Ptr && reflect.PtrTo(v.Type()).Implements(check)
}

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

var _ json.Unmarshaler = (*blobSaver)(nil)

type disfixRepr parseableDisfixRepr

type parseableDisfixRepr struct {
	Disfix string     `json:"_df"`
	Data   *blobSaver `json:"_v"`
}

func (dfr *disfixRepr) UnmarshalJSON(b []byte) error {
	// Some content might not be parseable
	recv := new(parseableDisfixRepr)
	if err := json.Unmarshal(b, recv); err != nil {
		// Perhaps the type doesn't conform to `{"_df":<disfix>, "_v":<data>}`
		// so in this case just save the data as it was sent in.
		recv.Data = &blobSaver{blob: b}
	}
	*dfr = (disfixRepr)(*recv)
	return nil
}

var blankDisfix DisfixBytes

// decodeDisfixJSON helps unravel the disfix and
// the stored data, which are expected in the form:
// {
//    "_df": "XXXXXXXXXXXXXXXXX",
//    "_v":  {}
// }
func decodeDisfixJSON(bz []byte) (disfix DisfixBytes, dataBytes []byte, err error) {
	bz = bytes.TrimSpace(bz)
	if len(bz) < DisfixBytesLen {
		return disfix, bz, errors.New("decodeDisfixJSON: EOF skipping prefix bytes.")
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
		return disfix, bz, fmt.Errorf("Disfix length got=%d want=%d data=%s", g, w, bz)
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
