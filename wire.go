package wire

import (
	"bytes"
	"fmt"
	"reflect"
)

//----------------------------------------
// Global entrypoint

var gCodec = NewCodec()

func MarshalBinary(o interface{}) ([]byte, error) {
	return gCodec.MarshalBinary(o)
}
func UnmarshalBinary(bz []byte, ptr interface{}) error {
	return gCodec.UnmarshalBinary(bz, ptr)
}
func UnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{}) error {
	return gCodec.UnmarshalBinaryLengthPrefixed(bz, ptr)
}
func MarshalJSON(o interface{}) ([]byte, error) {
	return gCodec.MarshalJSON(o)
}
func UnmarshalJSON(bz []byte, ptr interface{}) error {
	return gCodec.UnmarshalJSON(bz, ptr)
}
func UnmarshalJSONLengthPrefixed(bz []byte, ptr interface{}) error {
	return gCodec.UnmarshalJSONLengthPrefixed(bz, ptr)
}
func RegisterInterface(ptr interface{}, opts *InterfaceOptions) {
	gCodec.RegisterInterface(ptr, opts)
}
func RegisterConcrete(o interface{}, name string, opts *ConcreteOptions) {
	gCodec.RegisterConcrete(o, name, opts)
}

//----------------------------------------
// *Codec methods

// For consistency, MarshalBinary will first dereference pointers
// before encoding.  MarshalBinary will panic if o is a nil-pointer,
// or if o is invalid.
func (cdc *Codec) MarshalBinary(o interface{}) ([]byte, error) {

	// Dereference pointer.
	var rv = reflect.ValueOf(o)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
		if !rv.IsValid() {
			// NOTE: You can still do so by calling
			// `.MarshalBinary(struct{ *SomeType })` or so on.
			panic("MarshalBinary cannot marshal a nil pointer.")
		}
	}

	w := new(bytes.Buffer)
	rt := rv.Type()
	info, err := cdc.getTypeInfo_wlock(rt)
	if err != nil {
		return nil, err
	}
	err = cdc.encodeReflectBinary(w, info, rv, FieldOptions{})
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// UnmarshalBinary will panic if ptr is a nil-pointer.
func (cdc *Codec) UnmarshalBinary(bz []byte, ptr interface{}) error {
	rv, rt := reflect.ValueOf(ptr), reflect.TypeOf(ptr)
	if rv.Kind() != reflect.Ptr {
		panic("Unmarshal expects a pointer")
	}
	rv, rt = rv.Elem(), rt.Elem()
	info, err := cdc.getTypeInfo_wlock(rt)
	if err != nil {
		return err
	}
	n, err := cdc.decodeReflectBinary(bz, info, rv, FieldOptions{})
	if err != nil {
		return err
	}
	if n != len(bz) {
		return fmt.Errorf("Unmarshal didn't read all bytes. Expected to read %v, only read %v", len(bz), n)
	}
	return nil
}

func (cdc *Codec) UnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{}) error {
	panic("not implemented yet") // XXX
}

func (cdc *Codec) MarshalJSON(o interface{}) ([]byte, error) {
	w := new(bytes.Buffer)
	rv := reflect.ValueOf(o)
	if rv.Kind() == reflect.Invalid {
		return []byte("null"), nil
	}
	rt := rv.Type()
	info, err := cdc.getTypeInfo_wlock(rt)
	if err != nil {
		return nil, err
	}
	if err := cdc.encodeReflectJSON(w, info, rv, FieldOptions{}); err != nil {
		return nil, err
	}
	return w.Bytes(), nil

}

func (cdc *Codec) UnmarshalJSON(bz []byte, ptr interface{}) error {
	panic("not implemented yet") // XXX
}

func (cdc *Codec) UnmarshalJSONLengthPrefixed(bz []byte, ptr interface{}) error {
	panic("not implemented yet") // XXX
}
