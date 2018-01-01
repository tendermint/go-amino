package wire

import (
	"bytes"
	"fmt"
	"reflect"
)

func MarshalBinary(o interface{}) ([]byte, error) {
	w, n, err := new(bytes.Buffer), new(int), new(error)
	WriteBinary(o, w, n, err)
	if *err != nil {
		return nil, *err
	}

	return w.Bytes(), nil

	rv := reflect.ValueOf(o)
	rt := reflect.TypeOf(o)
	writeReflectBinary(rv, rt, Options{}, w, n, err)
}

func UnmarshalBinary(bz []byte, ptr interface{}) error {
	rv, rt := reflect.ValueOf(ptr), reflect.TypeOf(ptr)
	if rv.Kind() != reflect.Ptr {
		panic("Unmarshal expects a pointer")
	}
	n, err := decodeReflectBinary(bz, rv.Elem(), rt.Elem(), FieldOptions{})
	if err != nil {
		return err
	}
	if n != len(bz) {
		err = fmt.Errorf("Unmarshal didn't read all bytes")
	}
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
