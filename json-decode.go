package wire

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/davecgh/go-spew/spew"
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

	if printLog {
		spew.Printf("(d) decodeReflectJSON(bz: %X, info: %v, rv: %#v (%v), opts: %v)\n",
			bz, info, rv.Interface(), rv.Type(), opts)
		defer func() {
			fmt.Printf("(d) -> err: %v\n", err)
		}()
	}

	if !info.Registered {
		// No need for disambiguation, decode as is.
		err = cdc._decodeReflectJSON(bz, info, rv, opts)
		return
	}

	// It's a registered concrete type.
	// Implies that info holds the info we need.
	// Just strip the disfix bytes after checking it.
	disfix, bz, err := decodeDisfixJSON(bz)
	if err != nil {
		return
	}
	if !info.Prefix.EqualBytes(disfix[:]) {
		panic("should not happen")
	}

	err = cdc._decodeReflectJSON(bz, info, rv, opts)
	return
}

// CONTRACT: rv.CanAddr() is true.
func (cdc *Codec) _decodeReflectJSON(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) error {

	// Special case for nil for either interface, pointer, slice
	// NOTE: This doesn't match the binary implementation completely.
	if nilBytes(bz) {
		switch rv.Kind() {
		case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Array:
			rv.Set(info.ZeroValue)
			return nil
		}
	}

	// Dereference-and-construct pointers all the way.
	// This works for pointer-pointers.
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			newPtr := reflect.New(rv.Type().Elem())
			rv.Set(newPtr)
		}
		rv = rv.Elem()
	}

	// If a pointer to the dereferenced type implements json.Unmarshaller...
	if rv.Addr().Type().Implements(unmarshalerType) {
		return rv.Addr().Interface().(json.Unmarshaler).UnmarshalJSON(bz)
	}

	switch ikind := info.Type.Kind(); ikind {

	//----------------------------------------
	// Complex

	case reflect.Interface:
		return cdc.decodeReflectJSONInterface(bz, info, rv, opts)

	case reflect.Array:
		return cdc.decodeReflectJSONArray(bz, info, rv, opts)

	case reflect.Slice:
		return cdc.decodeReflectJSONSlice(bz, info, rv, opts)

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
func (cdc *Codec) decodeReflectJSONInterface(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}
	if !rv.IsNil() {
		// JAE: Heed this note, this is very tricky.
		// I forget why.
		err = errors.New("Decoding to a non-nil interface is not supported yet")
		return
	}

	// Consume disambiguation / prefix info.
	disfix, bz, err := decodeDisfixJSON(bz)
	if err != nil {
		return
	}

	// NOTE: Unlike decodeReflectBinaryInterface, we already dealt with nil in _decodeReflectJSON.
	// NOTE: We also "consumed" the disfix wrapper by replacing `bz` above.

	// Get concrete type info.
	// NOTE: Unlike decodeReflectBinaryInterface, always disfix.
	var cinfo *TypeInfo
	cinfo, err = cdc.getTypeInfoFromDisfix_rlock(disfix)
	if err != nil {
		return
	}

	// Construct the concrete type.
	var crv, irvSet = constructConcreteType(cinfo)

	// Decode into the concrete type.
	err = cdc._decodeReflectJSON(bz, cinfo, crv, opts)
	if err != nil {
		rv.Set(irvSet) // Helps with debugging
		return
	}

	// We need to set here, for when !PointerPreferred and the type
	// is say, an array of bytes (e.g. [32]byte), then we must call
	// rv.Set() *after* the value was acquired.
	rv.Set(irvSet)
	return
}

// CONTRACT: rv.CanAddr() is true.
func (cdc *Codec) decodeReflectJSONArray(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}
	ert := info.Type.Elem()
	length := info.Type.Len()

	switch ert.Kind() {

	case reflect.Uint8: // Special case: byte array
		err = json.Unmarshal(bz, rv)
		return

	default: // General case.
		var einfo *TypeInfo
		einfo, err = cdc.getTypeInfo_wlock(ert)
		if err != nil {
			return
		}

		// Read into rawSlice.
		var rawSlice []json.RawMessage
		if err = json.Unmarshal(bz, &rawSlice); err != nil {
			return
		}
		if len(rawSlice) != length {
			err = fmt.Errorf("decodeReflectJSONArray: length mismatch, got %v want %v", len(rawSlice), length)
			return
		}

		// Decode each item in rawSlice.
		for i := 0; i < length; i++ {
			erv := rv.Index(i)
			ebz := rawSlice[i]
			err = cdc.decodeReflectJSON(ebz, einfo, erv, opts)
			if err != nil {
				return
			}
		}
		return
	}
}

// CONTRACT: rv.CanAddr() is true.
func (cdc *Codec) decodeReflectJSONSlice(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}
	ert := info.Type.Elem()

	switch ert.Kind() {

	case reflect.Uint8: // Special case: byte slice
		err = json.Unmarshal(bz, rv)
		if err != nil {
			return
		}
		if rv.Len() == 0 {
			// Special case when length is 0.
			// NOTE: We prefer nil slices.
			rv.Set(info.ZeroValue)
		} else {
			// NOTE: Already set via json.Unmarshal() above.
		}
		return

	default: // General case.
		var einfo *TypeInfo
		einfo, err = cdc.getTypeInfo_wlock(ert)
		if err != nil {
			return
		}

		// Read into rawSlice.
		var rawSlice []json.RawMessage
		if err = json.Unmarshal(bz, &rawSlice); err != nil {
			return
		}

		// Special case when length is 0.
		// NOTE: We prefer nil slices.
		var length = len(rawSlice)
		if length == 0 {
			rv.Set(info.ZeroValue)
			return
		}

		// Read into a new slice.
		var esrt = reflect.SliceOf(ert) // TODO could be optimized.
		var srv = reflect.MakeSlice(esrt, length, length)
		for i := 0; i < length; i++ {
			erv := srv.Index(i)
			err = cdc.decodeReflectJSON(bz, einfo, erv, opts)
			if err != nil {
				return
			}
		}

		// TODO do we need this extra step?
		rv.Set(srv)
		return
	}
}

// CONTRACT: rv.CanAddr() is true.
func (cdc *Codec) decodeReflectJSONStruct(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}

	// Map all the fields(keys) to their blobs/bytes.
	// NOTE: In decodeReflectBinaryStruct, we don't need to do this,
	// since fields are encoded in order.
	var rawMap = make(map[string]json.RawMessage)
	err = json.Unmarshal(bz, &rawMap)
	if err != nil {
		return
	}

	for _, field := range info.Fields {

		// Get value from rawMap.
		var valueBytes, ok = rawMap[field.JSONName]
		if !ok {
			// TODO: Since the Go stdlib's JSON codec allows case-insensitive
			// keys perhaps we need to also do case-insensitive lookups here.
			// So "Vanilla" and "vanilla" would both match to the same field.
			// It is actually a security flaw with encoding/json library
			//  See https://github.com/golang/go/issues/14750
			// but perhaps we are aiming for as much compatibility here.
			continue
		}
		if valueBytes == nil {
			continue
		}

		// Get field rv and info.
		var frv = rv.Field(field.Index)
		var finfo *TypeInfo
		finfo, err = cdc.getTypeInfo_wlock(field.Type)
		if err != nil {
			return
		}

		// Decode into field rv.
		err = cdc.decodeReflectJSON(valueBytes, finfo, frv, opts)
		if err != nil {
			return
		}
	}

	return nil
}

//----------------------------------------
// Misc.

type disfixWrapper struct {
	Disfix string          `json:"_df"`
	Data   json.RawMessage `json:"_v"`
}

// decodeDisfixJSON helps unravel the disfix and
// the stored data, which are expected in the form:
// {
//    "_df": "XXXXXXXXXXXXXXXXX",
//    "_v":  {}
// }
func decodeDisfixJSON(bz []byte) (df DisfixBytes, data []byte, err error) {
	dfw := new(disfixWrapper)
	err = json.Unmarshal(bz, dfw)
	if err != nil {
		err = fmt.Errorf("Cannot parse disfix JSON wrapper: %v", err)
		return
	}
	dfBytes, err := hex.DecodeString(dfw.Disfix)
	if err != nil {
		return
	}

	// Get disfix.
	if g, w := len(dfBytes), DisfixBytesLen; g != w {
		err = fmt.Errorf("Disfix length got=%d want=%d data=%s", g, w, bz)
		return
	}
	copy(df[:], dfBytes)
	if (DisfixBytes{}).EqualBytes(df[:]) {
		err = errors.New("Unexpected zero disfix in JSON")
		return
	}

	// Get data.
	data = dfw.Data
	return
}
