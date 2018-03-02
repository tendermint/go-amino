package wire

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

// MarshalBinary encodes the object o according to the Wire spec,
// but prefixed by a uvarint encoding of the object to encode.
// Use MarshalBinaryBare if you don't want byte-length prefixing.
//
// For consistency, MarshalBinary will first dereference pointers
// before encoding.  MarshalBinary will panic if o is a nil-pointer,
// or if o is invalid.
func (cdc *Codec) MarshalBinary(o interface{}) ([]byte, error) {

	// Write the bytes here.
	var buf = new(bytes.Buffer)

	// Write the bz without length-prefixing.
	bz, err := cdc.MarshalBinaryBare(o)
	if err != nil {
		return nil, err
	}

	// Write uvarint(len(bz)).
	err = EncodeUvarint(buf, uint64(len(bz)))
	if err != nil {
		return nil, err
	}

	// Write bz.
	_, err = buf.Write(bz)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// MarshalBinaryBare encodes the object o according to the Wire spec.
// MarshalBinaryBare doesn't prefix the byte-length of the encoding,
// so the caller must handle framing.
func (cdc *Codec) MarshalBinaryBare(o interface{}) ([]byte, error) {

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

	// Encode Wire:binary bytes.
	var bz []byte
	buf := new(bytes.Buffer)
	rt := rv.Type()
	info, err := cdc.getTypeInfo_wlock(rt)
	if err != nil {
		return nil, err
	}
	err = cdc.encodeReflectBinary(buf, info, rv, FieldOptions{})
	if err != nil {
		return nil, err
	}
	bz = buf.Bytes()

	return bz, nil
}

// Like UnmarshalBinaryBare, but will first decode the byte-length prefix.
// UnmarshalBinary will panic if ptr is a nil-pointer.
// Returns an error if not all of bz is consumed.
func (cdc *Codec) UnmarshalBinary(bz []byte, ptr interface{}) error {

	// Read byte-length prefix.
	u64, n := binary.Uvarint(bz)
	if n < 0 {
		return fmt.Errorf("Error reading msg byte-length prefix: got code %v", n)
	}
	if u64 > uint64(len(bz)-n) {
		return fmt.Errorf("Not enough bytes to read in UnmarshalBinary, want %v more bytes but only have %v",
			u64, len(bz)-n)
	} else if u64 < uint64(len(bz)-n) {
		return fmt.Errorf("Bytes left over in UnmarshalBinary, should read %v more bytes but have %v",
			u64, len(bz)-n)
	}
	bz = bz[n:]

	// Decode.
	return cdc.UnmarshalBinaryBare(bz, ptr)
}

// Like UnmarshalBinaryBare, but will first read the byte-length prefix.
// UnmarshalBinaryReader will panic if ptr is a nil-pointer.
// If maxSize is 0, there is no limit (not recommended).
func (cdc *Codec) UnmarshalBinaryReader(r io.Reader, ptr interface{}, maxSize int64) error {
	if maxSize < 0 {
		panic("maxSize cannot be negative.")
	}
	var br = bufio.NewReader(r)

	// Read byte-length prefix.
	var l int64
	u64, err := binary.ReadUvarint(br)
	if err != nil {
		return err
	}
	if maxSize > 0 && u64 > uint64(maxSize) {
		return fmt.Errorf("Read overflow, maxSize is %v but next message is %v bytes", maxSize, u64)
	}
	l = int64(u64)

	// Read that many bytes.
	var bz = make([]byte, l, l)
	_, err = io.ReadFull(br, bz)
	if err != nil {
		return err
	}

	// Decode.
	return cdc.UnmarshalBinaryBare(bz, ptr)
}

// UnmarshalBinaryBare will panic if ptr is a nil-pointer.
func (cdc *Codec) UnmarshalBinaryBare(bz []byte, ptr interface{}) error {
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

//----------------------------------------
// Misc.

var (
	marshalerType   = reflect.TypeOf(new(json.Marshaler)).Elem()
	unmarshalerType = reflect.TypeOf(new(json.Unmarshaler)).Elem()
)
