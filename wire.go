package wire

import (
	"bytes"
	"errors"
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

func (cdc *Codec) MarshalBinary(o interface{}) ([]byte, error) {
	w := new(bytes.Buffer)
	rv := reflect.ValueOf(o)
	rt := reflect.TypeOf(o)
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
		return errors.New("Unmarshal didn't read all bytes")
	}
	return nil
}

func (cdc *Codec) UnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{}) error {
	panic("not implemented yet") // XXX
}

func (cdc *Codec) MarshalJSON(o interface{}) ([]byte, error) {
	panic("not implemented yet") // XXX
}

func (cdc *Codec) UnmarshalJSON(bz []byte, ptr interface{}) error {
	panic("not implemented yet") // XXX
}

func (cdc *Codec) UnmarshalJSONLengthPrefixed(bz []byte, ptr interface{}) error {
	panic("not implemented yet") // XXX
}
