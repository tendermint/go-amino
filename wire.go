package wire

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
)

//----------------------------------------
// *Codec methods

// For consistency, MarshalBinary will first dereference pointers
// before encoding.  MarshalBinary will panic if o is a nil-pointer,
// or if o is invalid.
func (cdc *Codec) MarshalBinary(o interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := cdc._marshalBinaryStream(buf, o); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (cdc *Codec) MarshalBinaryStream(w io.Writer, o interface{}) error {
	return cdc._marshalBinaryStream(w, o)
}

func (cdc *Codec) _marshalBinaryStream(w io.Writer, o interface{}) error {

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

	rt := rv.Type()
	info, err := cdc.getTypeInfo_wlock(rt)
	if err != nil {
		return err
	}
	return cdc.encodeReflectBinary(w, info, rv, FieldOptions{})
}

// UnmarshalBinary will panic if ptr is a nil-pointer.
func (cdc *Codec) UnmarshalBinary(bz []byte, ptr interface{}) error {
	return cdc._unmarshalBinary(nil, bz, ptr)
}

func (cdc *Codec) UnmarshalBinaryStream(r io.Reader, ptr interface{}) error {
	return cdc._unmarshalBinary(r, nil, ptr)
}

func (cdc *Codec) _unmarshalBinary(r io.Reader, bzz []byte, ptr interface{}) error {
	rv, rt := reflect.ValueOf(ptr), reflect.TypeOf(ptr)
	if rv.Kind() != reflect.Ptr {
		panic("Unmarshal expects a pointer")
	}
	rv, rt = rv.Elem(), rt.Elem()
	info, err := cdc.getTypeInfo_wlock(rt)
	if err != nil {
		return err
	}

	readFromByteSlice := r == nil || len(bzz) > 0

	// Creating the *bufio.Reader
	var bz *bufio.Reader
	if readFromByteSlice {
		bz = bufio.NewReader(bytes.NewReader(bzz))
	} else {
		if bzr, ok := r.(*bufio.Reader); ok {
			bz = bzr
		} else {
			bz = bufio.NewReader(r)
		}
	}

	n, err := cdc.decodeReflectBinary(bz, info, rv, FieldOptions{})
	if err != nil {
		return err
	}
	if readFromByteSlice && n != len(bzz) {
		return fmt.Errorf("Unmarshal didn't read all bytes. Expected to read %v, only read %v", len(bzz), n)
	}
	return nil
}

func (cdc *Codec) UnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{}) error {
	panic("not implemented yet") // XXX
}

var (
	marshalerType   = reflect.TypeOf(new(json.Marshaler)).Elem()
	unmarshalerType = reflect.TypeOf(new(json.Unmarshaler)).Elem()
)

func (cdc *Codec) MarshalJSON(o interface{}) ([]byte, error) {
	rv := reflect.ValueOf(o)
	if rv.Kind() == reflect.Invalid {
		return []byte("null"), nil
	}
	rt := rv.Type()

	// Note that we can't yet skip directly
	// to checking if a type implements
	// json.Marshaler because in some cases
	// var s GenericInterface = t1(v1)
	// var t GenericInterface = t2(v1)
	// but we need to be able to encode
	// both s and t disambiguated, so:
	//    {"_df":<disfix>, "_v":<data>}
	// for the above case.

	w := new(bytes.Buffer)
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
	rv := reflect.ValueOf(ptr)
	if rv.Kind() != reflect.Ptr {
		return errors.New("UnmarshalJSON expects a pointer")
	}

	// If the type implements json.Unmarshaler, just
	// automatically respect that and skip to it.
	// if rv.Type().Implements(unmarshalerType) {
	// 	return rv.Interface().(json.Unmarshaler).UnmarshalJSON(bz)
	// }

	// 1. Dereference until we find the first addressable type.
	rv = rv.Elem()
	rt := rv.Type()
	info, err := cdc.getTypeInfo_wlock(rt)
	if err != nil {
		return err
	}
	return cdc.decodeReflectJSON(bz, info, rv, FieldOptions{})
}
