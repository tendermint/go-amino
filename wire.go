package wire

import (
	"bytes"
	"fmt"
	"reflect"
)

func MarshalBinary(o interface{}) ([]byte, error) {
	w := new(bytes.Buffer)
	rv := reflect.ValueOf(o)
	rt := reflect.TypeOf(o)
	info, err := getTypeInfo(rt)
	if err != nil {
		return nil, err
	}
	err = encodeReflectBinary(w, info, rv, FieldOptions{})
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func UnmarshalBinary(bz []byte, ptr interface{}) error {
	rv, rt := reflect.ValueOf(ptr), reflect.TypeOf(ptr)
	if rv.Kind() != reflect.Ptr {
		panic("Unmarshal expects a pointer")
	}
	rv, rt = rv.Elem(), rt.Elem()
	info, err := getTypeInfo(rt)
	if err != nil {
		return err
	}
	n, err := decodeReflectBinary(bz, info, rv, FieldOptions{})
	if err != nil {
		return err
	}
	if n != len(bz) {
		return fmt.Errorf("Unmarshal didn't read all bytes")
	}
	return nil
}

func UnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{}) error {
	panic("not implemented yet") // XXX
}

func MarshalJSON(o interface{}) ([]byte, error) {
	panic("not implemented yet") // XXX
}

func UnmarshalJSON(bz []byte, ptr interface{}) error {
	panic("not implemented yet") // XXX
}

func UnmarshalJSONLengthPrefixed(bz []byte, ptr interface{}) error {
	panic("not implemented yet") // XXX
}
