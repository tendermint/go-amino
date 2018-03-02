package wire

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

//----------------------------------------
// Typ3 and Typ4

type Typ3 uint8
type Typ4 uint8 // Typ3 | 0x80 (pointer bit)

const (
	// Typ3 types
	Typ3_Varint     = Typ3(0)
	Typ3_8Byte      = Typ3(1)
	Typ3_ByteLength = Typ3(2)
	Typ3_Struct     = Typ3(3)
	Typ3_StructTerm = Typ3(4)
	Typ3_4Byte      = Typ3(5)
	Typ3_List       = Typ3(6)
	Typ3_Interface  = Typ3(7)

	// Typ4 bit
	Typ4_Pointer = Typ4(0x08)
)

func (typ Typ3) String() string {
	switch typ {
	case Typ3_Varint:
		return "Varint"
	case Typ3_8Byte:
		return "8Byte"
	case Typ3_ByteLength:
		return "ByteLength"
	case Typ3_Struct:
		return "Struct"
	case Typ3_StructTerm:
		return "StructTerm"
	case Typ3_4Byte:
		return "4Byte"
	case Typ3_List:
		return "List"
	case Typ3_Interface:
		return "Interface"
	default:
		return fmt.Sprintf("<Invalid Typ3 %X>", byte(typ))
	}
}

func (typ Typ4) Typ3() Typ3 { return Typ3(typ & 0x07) }
func (typ Typ4) String() string {
	if typ&0xF0 != 0 {
		return fmt.Sprintf("<Invalid Typ4 %X>", byte(typ))
	}
	if typ&0x80 != 0 {
		return "*" + Typ3(typ).String()
	} else {
		return Typ3(typ).String()
	}
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

func (cdc *Codec) MarshalBinaryLengthPrefied(o interface{}) ([]byte, error) {
	panic("not implemented yet") // XXX
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
