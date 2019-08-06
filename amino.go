package amino

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"time"

	"encoding/binary"
	"encoding/json"

	"github.com/pkg/errors"
)

var (
	// Global methods for global sealed codec.
	gcdc *Codec

	// we use this time to init. a zero value (opposed to reflect.Zero which gives time.Time{} / 01-01-01 00:00:00)
	zeroTime time.Time

	// ErrNoPointer is thrown when you call a method that expects a pointer, e.g. Unmarshal
	ErrNoPointer = errors.New("expected a pointer")
)

const (
	unixEpochStr = "1970-01-01 00:00:00 +0000 UTC"
	epochFmt     = "2006-01-02 15:04:05 +0000 UTC"
)

func init() {
	gcdc = NewCodec().Seal()
	var err error
	zeroTime, err = time.Parse(epochFmt, unixEpochStr)
	if err != nil {
		panic("couldn't parse Zero value for time")
	}
}

func MarshalBinaryLengthPrefixed(o interface{}) ([]byte, error) {
	return gcdc.MarshalBinaryLengthPrefixed(o)
}

func MarshalBinaryLengthPrefixedWriter(w io.Writer, o interface{}) (n int64, err error) {
	return gcdc.MarshalBinaryLengthPrefixedWriter(w, o)
}

func MustMarshalBinaryLengthPrefixed(o interface{}) []byte {
	return gcdc.MustMarshalBinaryLengthPrefixed(o)
}

func MarshalBinaryBare(o interface{}) ([]byte, error) {
	return gcdc.MarshalBinaryBare(o)
}

func MustMarshalBinaryBare(o interface{}) []byte {
	return gcdc.MustMarshalBinaryBare(o)
}

func UnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{}) error {
	return gcdc.UnmarshalBinaryLengthPrefixed(bz, ptr)
}

func UnmarshalBinaryLengthPrefixedReader(r io.Reader, ptr interface{}, maxSize int64) (n int64, err error) {
	return gcdc.UnmarshalBinaryLengthPrefixedReader(r, ptr, maxSize)
}

func MustUnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{}) {
	gcdc.MustUnmarshalBinaryLengthPrefixed(bz, ptr)
}

func UnmarshalBinaryBare(bz []byte, ptr interface{}) error {
	return gcdc.UnmarshalBinaryBare(bz, ptr)
}

func MustUnmarshalBinaryBare(bz []byte, ptr interface{}) {
	gcdc.MustUnmarshalBinaryBare(bz, ptr)
}

func MarshalJSON(o interface{}) ([]byte, error) {
	return gcdc.MarshalJSON(o)
}

func UnmarshalJSON(bz []byte, ptr interface{}) error {
	return gcdc.UnmarshalJSON(bz, ptr)
}

func MarshalJSONIndent(o interface{}, prefix, indent string) ([]byte, error) {
	return gcdc.MarshalJSONIndent(o, prefix, indent)
}

//----------------------------------------
// Typ3

type Typ3 uint8

const (
	// Typ3 types
	Typ3Varint     = Typ3(0)
	Typ38Byte      = Typ3(1)
	Typ3ByteLength = Typ3(2)
	//Typ3_Struct     = Typ3(3)
	//Typ3_StructTerm = Typ3(4)
	Typ3_4Byte = Typ3(5)
	//Typ3_List       = Typ3(6)
	//Typ3_Interface  = Typ3(7)
)

func (typ Typ3) String() string {
	switch typ {
	case Typ3Varint:
		return "(U)Varint"
	case Typ38Byte:
		return "8Byte"
	case Typ3ByteLength:
		return "ByteLength"
	//case Typ3_Struct:
	//	return "Struct"
	//case Typ3_StructTerm:
	//	return "StructTerm"
	case Typ3_4Byte:
		return "4Byte"
	//case Typ3_List:
	//	return "List"
	//case Typ3_Interface:
	//	return "Interface"
	default:
		return fmt.Sprintf("<Invalid Typ3 %X>", byte(typ))
	}
}

//----------------------------------------
// *Codec methods

// MarshalBinaryLengthPrefixed encodes the object o according to the Amino spec,
// but prefixed by a uvarint encoding of the object to encode.
// Use MarshalBinaryBare if you don't want byte-length prefixing.
//
// For consistency, MarshalBinaryLengthPrefixed will first dereference pointers
// before encoding.  MarshalBinaryLengthPrefixed will panic if o is a nil-pointer,
// or if o is invalid.
func (cdc *Codec) MarshalBinaryLengthPrefixed(o interface{}) ([]byte, error) {

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

// MarshalBinaryLengthPrefixedWriter writes the bytes as would be returned from
// MarshalBinaryLengthPrefixed to the writer w.
func (cdc *Codec) MarshalBinaryLengthPrefixedWriter(w io.Writer, o interface{}) (n int64, err error) {
	var bz, _n = []byte(nil), int(0)
	bz, err = cdc.MarshalBinaryLengthPrefixed(o)
	if err != nil {
		return 0, err
	}
	_n, err = w.Write(bz) // TODO: handle overflow in 32-bit systems.
	n = int64(_n)
	return
}

// Panics if error.
func (cdc *Codec) MustMarshalBinaryLengthPrefixed(o interface{}) []byte {
	bz, err := cdc.MarshalBinaryLengthPrefixed(o)
	if err != nil {
		panic(err)
	}
	return bz
}

// MarshalBinaryBare encodes the object o according to the Amino spec.
// MarshalBinaryBare doesn't prefix the byte-length of the encoding,
// so the caller must handle framing.
func (cdc *Codec) MarshalBinaryBare(o interface{}) ([]byte, error) {

	// Dereference value if pointer.
	var rv, _, isNilPtr = derefPointers(reflect.ValueOf(o))
	if isNilPtr {
		// NOTE: You can still do so by calling
		// `.MarshalBinaryLengthPrefixed(struct{ *SomeType })` or so on.
		panic("MarshalBinaryBare cannot marshal a nil pointer directly. Try wrapping in a struct?")
	}

	// Encode Amino:binary bytes.
	var bz []byte
	buf := new(bytes.Buffer)
	rt := rv.Type()
	info, err := cdc.getTypeInfoWlock(rt)
	if err != nil {
		return nil, err
	}
	// in the case of of a repeated struct (e.g. type Alias []SomeStruct),
	// we do not need to prepend with `(field_number << 3) | wire_type` as this
	// would need to be done for each struct and not only for the first.
	if rv.Kind() != reflect.Struct && !isStructOrRepeatedStruct(info) {
		writeEmpty := false
		typ3 := typeToTyp3(info.Type, FieldOptions{})
		bare := typ3 != Typ3ByteLength
		if err := cdc.writeFieldIfNotEmpty(buf, 1, info, FieldOptions{}, FieldOptions{}, rv, writeEmpty, bare); err != nil {
			return nil, err
		}
		bz = buf.Bytes()
	} else {
		err = cdc.encodeReflectBinary(buf, info, rv, FieldOptions{BinFieldNum: 1}, true)
		if err != nil {
			return nil, err
		}
		bz = buf.Bytes()
	}
	// If registered concrete, prepend prefix bytes.
	if info.Registered {
		// TODO: https://github.com/tendermint/go-amino/issues/267
		//return MarshalBinaryBare(RegisteredAny{
		//	AminoPreOrDisfix: info.Prefix.Bytes(),
		//	Value: bz,
		//})
		pb := info.Prefix.Bytes()
		bz = append(pb, bz...)
	}

	return bz, nil
}

//type RegisteredAny struct {
//	AminoPreOrDisfix []byte
//	Value []byte
//}

// Panics if error.
func (cdc *Codec) MustMarshalBinaryBare(o interface{}) []byte {
	bz, err := cdc.MarshalBinaryBare(o)
	if err != nil {
		panic(err)
	}
	return bz
}

// Like UnmarshalBinaryBare, but will first decode the byte-length prefix.
// UnmarshalBinaryLengthPrefixed will panic if ptr is a nil-pointer.
// Returns an error if not all of bz is consumed.
func (cdc *Codec) UnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{}) error {
	if len(bz) == 0 {
		return errors.New("UnmarshalBinaryLengthPrefixed cannot decode empty bytes")
	}

	// Read byte-length prefix.
	u64, n := binary.Uvarint(bz)
	if n < 0 {
		return errors.Errorf("Error reading msg byte-length prefix: got code %v", n)
	}
	if u64 > uint64(len(bz)-n) {
		return errors.Errorf("Not enough bytes to read in UnmarshalBinaryLengthPrefixed, want %v more bytes but only have %v",
			u64, len(bz)-n)
	} else if u64 < uint64(len(bz)-n) {
		return errors.Errorf("Bytes left over in UnmarshalBinaryLengthPrefixed, should read %v more bytes but have %v",
			u64, len(bz)-n)
	}
	bz = bz[n:]

	// Decode.
	return cdc.UnmarshalBinaryBare(bz, ptr)
}

// Like UnmarshalBinaryBare, but will first read the byte-length prefix.
// UnmarshalBinaryLengthPrefixedReader will panic if ptr is a nil-pointer.
// If maxSize is 0, there is no limit (not recommended).
func (cdc *Codec) UnmarshalBinaryLengthPrefixedReader(r io.Reader, ptr interface{},
	maxSize int64) (n int64, err error) {
	if maxSize < 0 {
		panic("maxSize cannot be negative.")
	}

	// Read byte-length prefix.
	var l int64
	var buf [binary.MaxVarintLen64]byte
	for i := 0; i < len(buf); i++ {
		_, err = r.Read(buf[i : i+1])
		if err != nil {
			return
		}
		n++
		if buf[i]&0x80 == 0 {
			break
		}
		if n >= maxSize {
			err = errors.Errorf(
				"read overflow, maxSize is %v but uvarint(length-prefix) is itself greater than maxSize",
				maxSize,
			)
		}
	}
	u64, _ := binary.Uvarint(buf[:])
	if err != nil {
		return
	}
	if maxSize > 0 {
		if uint64(maxSize) < u64 {
			err = errors.Errorf("read overflow, maxSize is %v but this amino binary object is %v bytes", maxSize, u64)
			return
		}
		if (maxSize - n) < int64(u64) {
			err = errors.Errorf(
				"read overflow, maxSize is %v but this length-prefixed amino binary object is %v+%v bytes",
				maxSize, n, u64,
			)
			return
		}
	}
	l = int64(u64)
	if l < 0 {
		_ = errors.Errorf(
			"read overflow, this implementation can't read this because, why would anyone have this much data? Hello from 2018",
		)
	}

	// Read that many bytes.
	var bz = make([]byte, l)
	_, err = io.ReadFull(r, bz)
	if err != nil {
		return
	}
	n += l

	// Decode.
	err = cdc.UnmarshalBinaryBare(bz, ptr)
	return
}

// Panics if error.
func (cdc *Codec) MustUnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{}) {
	err := cdc.UnmarshalBinaryLengthPrefixed(bz, ptr)
	if err != nil {
		panic(err)
	}
}

// UnmarshalBinaryBare will panic if ptr is a nil-pointer.
func (cdc *Codec) UnmarshalBinaryBare(bz []byte, ptr interface{}) error {

	rv := reflect.ValueOf(ptr)
	if rv.Kind() != reflect.Ptr {
		return ErrNoPointer
	}
	rv = rv.Elem()
	rt := rv.Type()
	info, err := cdc.getTypeInfoWlock(rt)
	if err != nil {
		return err
	}

	// If registered concrete, consume and verify prefix bytes.
	if info.Registered {
		// TODO: https://github.com/tendermint/go-amino/issues/267
		pb := info.Prefix.Bytes()
		if len(bz) < 4 {
			return fmt.Errorf(
				"unmarshalBinaryBare expected to read prefix bytes %X (since it is registered concrete) but got %X",
				pb, bz,
			)
		} else if !bytes.Equal(bz[:4], pb) {
			return fmt.Errorf(
				"unmarshalBinaryBare expected to read prefix bytes %X (since it is registered concrete) but got %X",
				pb, bz[:4],
			)
		}
		bz = bz[4:]
	}
	// Only add length prefix if we have another typ3 then Typ3ByteLength.
	// Default is non-length prefixed:
	bare := true
	var nWrap int
	isKnownType := (info.Type.Kind() != reflect.Map) && (info.Type.Kind() != reflect.Func)
	if !isStructOrRepeatedStruct(info) &&
		!isPointerToStructOrToRepeatedStruct(rv, rt) &&
		len(bz) > 0 &&
		(rv.Kind() != reflect.Interface) &&
		isKnownType {
		fnum, typ, nFnumTyp3, err := decodeFieldNumberAndTyp3(bz)
		if err != nil {
			return errors.Wrap(err, "could not decode field number and type")
		}
		if fnum != 1 {
			return fmt.Errorf("expected field number: 1; got: %v", fnum)
		}
		typWanted := typeToTyp3(info.Type, FieldOptions{})
		if typ != typWanted {
			return fmt.Errorf("expected field type %v for # %v of %v, got %v",
				typWanted, fnum, info.Type, typ)
		}

		slide(&bz, &nWrap, nFnumTyp3)
		bare = typeToTyp3(info.Type, FieldOptions{}) != Typ3ByteLength
	}

	// Decode contents into rv.
	n, err := cdc.decodeReflectBinary(bz, info, rv, FieldOptions{BinFieldNum: 1}, bare)
	if err != nil {
		return fmt.Errorf(
			"unmarshal to %v failed after %d bytes (%v): %X",
			info.Type,
			n+nWrap,
			err,
			bz,
		)
	}
	if n != len(bz) {
		return fmt.Errorf(
			"unmarshal to %v didn't read all bytes. Expected to read %v, only read %v: %X",
			info.Type,
			len(bz),
			n+nWrap,
			bz,
		)
	}

	return nil
}

func isStructOrRepeatedStruct(info *TypeInfo) bool {
	if info.Type.Kind() == reflect.Struct {
		return true
	}
	isRepeatedStructAr := info.Type.Kind() == reflect.Array && info.Type.Elem().Kind() == reflect.Struct
	isRepeatedStructSl := info.Type.Kind() == reflect.Slice && info.Type.Elem().Kind() == reflect.Struct
	return isRepeatedStructAr || isRepeatedStructSl
}

func isPointerToStructOrToRepeatedStruct(rv reflect.Value, rt reflect.Type) bool {
	if rv.Kind() == reflect.Struct {
		return true
	}

	drv, isPtr, isNil := derefPointers(rv)
	if isPtr && drv.Kind() == reflect.Struct {
		return true
	}

	if isPtr && isNil {
		rt := derefType(rt)
		if rt.Kind() == reflect.Struct {
			return true
		}
		return rt.Kind() == reflect.Slice && rt.Elem().Kind() == reflect.Struct ||
			rt.Kind() == reflect.Array && rt.Elem().Kind() == reflect.Struct
	}
	isRepeatedStructSl := isPtr && drv.Kind() == reflect.Slice && drv.Elem().Kind() == reflect.Struct
	isRepeatedStructAr := isPtr && drv.Kind() == reflect.Array && drv.Elem().Kind() == reflect.Struct
	return isRepeatedStructAr || isRepeatedStructSl
}

func derefType(rt reflect.Type) (drt reflect.Type) {
	drt = rt
	for drt.Kind() == reflect.Ptr {
		drt = drt.Elem()
	}
	return
}

// Panics if error.
func (cdc *Codec) MustUnmarshalBinaryBare(bz []byte, ptr interface{}) {
	err := cdc.UnmarshalBinaryBare(bz, ptr)
	if err != nil {
		panic(err)
	}
}

func (cdc *Codec) MarshalJSON(o interface{}) ([]byte, error) {
	rv := reflect.ValueOf(o)
	if rv.Kind() == reflect.Invalid {
		return []byte("null"), nil
	}
	rt := rv.Type()
	w := new(bytes.Buffer)
	info, err := cdc.getTypeInfoWlock(rt)
	if err != nil {
		return nil, err
	}

	// Write the disfix wrapper if it is a registered concrete type.
	if info.Registered {
		err = writeStr(w, _fmt(`{"type":"%s","value":`, info.Name))
		if err != nil {
			return nil, err
		}
	}

	// Write the rest from rv.
	if err := cdc.encodeReflectJSON(w, info, rv, FieldOptions{}); err != nil {
		return nil, err
	}

	// disfix wrapper continued...
	if info.Registered {
		err = writeStr(w, `}`)
		if err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}

// MustMarshalJSON panics if an error occurs. Besides tha behaves exactly like MarshalJSON.
func (cdc *Codec) MustMarshalJSON(o interface{}) []byte {
	bz, err := cdc.MarshalJSON(o)
	if err != nil {
		panic(err)
	}
	return bz
}

func (cdc *Codec) UnmarshalJSON(bz []byte, ptr interface{}) error {
	if len(bz) == 0 {
		return errors.New("cannot decode empty bytes")
	}

	rv := reflect.ValueOf(ptr)
	if rv.Kind() != reflect.Ptr {
		return errors.New("expected a pointer")
	}
	rv = rv.Elem()
	rt := rv.Type()
	info, err := cdc.getTypeInfoWlock(rt)
	if err != nil {
		return err
	}
	// If registered concrete, consume and verify type wrapper.
	if info.Registered {
		// Consume type wrapper info.
		name, data, err := decodeInterfaceJSON(bz)
		if err != nil {
			return err
		}
		// Check name against info.
		if name != info.Name {
			return errors.Errorf("wanted to decode %v but found %v", info.Name, name)
		}
		bz = data
	}
	return cdc.decodeReflectJSON(bz, info, rv, FieldOptions{})
}

// MustUnmarshalJSON panics if an error occurs. Besides tha behaves exactly like UnmarshalJSON.
func (cdc *Codec) MustUnmarshalJSON(bz []byte, ptr interface{}) {
	if err := cdc.UnmarshalJSON(bz, ptr); err != nil {
		panic(err)
	}
}

// MarshalJSONIndent calls json.Indent on the output of cdc.MarshalJSON
// using the given prefix and indent string.
func (cdc *Codec) MarshalJSONIndent(o interface{}, prefix, indent string) ([]byte, error) {
	bz, err := cdc.MarshalJSON(o)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	err = json.Indent(&out, bz, prefix, indent)
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
