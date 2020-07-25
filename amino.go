package amino

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"reflect"
	"runtime"
	"time"

	"github.com/pkg/errors"
	"github.com/tendermint/go-amino/pkg"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Package "pkg" exists So dependencies can create Packages.
// We export it here so this amino package can use it natively.
type Package = pkg.Package

var (
	// Global methods for global auto-sealing codec.
	gcdc *Codec

	// we use this time to init. an empty value (opposed to reflect.Zero which gives time.Time{} / 01-01-01 00:00:00)
	emptyTime time.Time

	// ErrNoPointer is thrown when you call a method that expects a pointer, e.g. Unmarshal
	ErrNoPointer = errors.New("expected a pointer")
)

const (
	unixEpochStr = "1970-01-01 00:00:00 +0000 UTC"
	epochFmt     = "2006-01-02 15:04:05 +0000 UTC"
)

func init() {
	gcdc = NewCodec().Autoseal()
	var err error
	emptyTime, err = time.Parse(epochFmt, unixEpochStr)
	if err != nil {
		panic("couldn't parse empty value for time")
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

func MarshalBinaryInterfaceLengthPrefixed(o interface{}) ([]byte, error) {
	return gcdc.MarshalBinaryInterfaceLengthPrefixed(o)
}

func MustMarshalBinaryInterfaceLengthPrefixed(o interface{}) []byte {
	return gcdc.MustMarshalBinaryInterfaceLengthPrefixed(o)
}

func MarshalBinaryBare(o interface{}) ([]byte, error) {
	return gcdc.MarshalBinaryBare(o)
}

func MustMarshalBinaryBare(o interface{}) []byte {
	return gcdc.MustMarshalBinaryBare(o)
}

func MarshalBinaryInterfaceBare(o interface{}) ([]byte, error) {
	return gcdc.MarshalBinaryInterfaceBare(o)
}

func MustMarshalBinaryInterfaceBare(o interface{}) []byte {
	return gcdc.MustMarshalBinaryInterfaceBare(o)
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

func UnmarshalBinaryAny(typeURL string, value []byte, ptr interface{}) error {
	return gcdc.UnmarshalBinaryAny(typeURL, value, ptr)
}

func MustUnmarshalBinaryAny(typeURL string, value []byte, ptr interface{}) {
	gcdc.MustUnmarshalBinaryAny(typeURL, value, ptr)
}

func MarshalJSON(o interface{}) ([]byte, error) {
	return gcdc.MarshalJSON(o)
}

func MarshalJSONInterface(o interface{}) ([]byte, error) {
	return gcdc.MarshalJSONInterface(o)
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
	Typ34Byte = Typ3(5)
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
	case Typ34Byte:
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

//----------------------------------------
// Marshal* methods

// MarshalBinaryLengthPrefixed encodes the object o according to the Amino spec,
// but prefixed by a uvarint encoding of the object to encode.
// Use MarshalBinaryBare if you don't want byte-length prefixing.
//
// For consistency, MarshalBinaryLengthPrefixed will first dereference pointers
// before encoding.  MarshalBinaryLengthPrefixed will panic if o is a nil-pointer,
// or if o is invalid.
func (cdc *Codec) MarshalBinaryLengthPrefixed(o interface{}) ([]byte, error) {
	cdc.doAutoseal()

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
	var (
		bz []byte
		_n int
	)
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

func (cdc *Codec) MarshalBinaryInterfaceLengthPrefixed(o interface{}) ([]byte, error) {
	cdc.doAutoseal()

	// Write the bytes here.
	var buf = new(bytes.Buffer)

	// Write the bz without length-prefixing.
	bz, err := cdc.MarshalBinaryInterfaceBare(o)
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

func (cdc *Codec) MustMarshalBinaryInterfaceLengthPrefixed(o interface{}) []byte {
	bz, err := cdc.MarshalBinaryInterfaceLengthPrefixed(o)
	if err != nil {
		panic(err)
	}
	return bz
}

// MarshalBinaryBare encodes the object o according to the Amino spec.
// MarshalBinaryBare doesn't prefix the byte-length of the encoding,
// so the caller must handle framing.
// Type information as in google.protobuf.Any isn't included, so manually wrap
// before calling if you need to decode into an interface.
// NOTE: nil-struct-pointers have no encoding. In the context of a struct,
// the absence of a field does denote a nil-struct-pointer, but in general
// this is not the case, so unlike MarshalJSON.
func (cdc *Codec) MarshalBinaryBare(o interface{}) ([]byte, error) {
	cdc.doAutoseal()

	if cdc.usePBBindings {
		pbm, ok := o.(PBMessager)
		if ok {
			return cdc.marshalBinaryBarePBBindings(pbm)
		} else {
			// Fall back to using relfection for native primitive types.
		}
	}

	return cdc.marshalBinaryBareReflect(o)
}

// Use reflection.
func (cdc *Codec) marshalBinaryBareReflect(o interface{}) ([]byte, error) {

	// Dereference value if pointer.
	var rv = reflect.ValueOf(o)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			panic("MarshalBinaryBare cannot marshal a nil pointer directly. Try wrapping in a struct?")
			// NOTE: You can still do so by calling
			// `.MarshalBinaryLengthPrefixed(struct{ *SomeType })` or so on.
		}
		rv = rv.Elem()
		if rv.Kind() == reflect.Ptr {
			panic("nested pointers not allowed")
		}
	}

	// Encode Amino:binary bytes.
	var bz []byte
	buf := new(bytes.Buffer)
	rt := rv.Type()
	info, err := cdc.getTypeInfoWLock(rt)
	if err != nil {
		return nil, err
	}
	// Implicit struct or not?
	// NOTE: similar to binary interface encoding.
	fopts := FieldOptions{}
	if !info.IsStructOrUnpacked(fopts) {
		writeEmpty := false
		// Encode with an implicit struct, with a single field with number 1.
		// The type of this implicit field determines whether any
		// length-prefixing happens after the typ3 byte.
		// The second FieldOptions is empty, because this isn't a list of
		// Typ3_ByteLength things, so however it is encoded, that option is no
		// longer needed.
		if err = cdc.writeFieldIfNotEmpty(buf, 1, info, FieldOptions{}, FieldOptions{}, rv, writeEmpty); err != nil {
			return nil, err
		}
		bz = buf.Bytes()
	} else {
		// The passed in BinFieldNum is only relevant for when the type is to
		// be encoded unpacked (elements are Typ3_ByteLength).  In that case,
		// encodeReflectBinary will repeat the field number as set here, as if
		// encoded with an implicit struct.
		err = cdc.encodeReflectBinary(buf, info, rv, FieldOptions{BinFieldNum: 1}, true, 0)
		if err != nil {
			return nil, err
		}
		bz = buf.Bytes()
	}

	return bz, nil
}

// Use pbbindings.
func (cdc *Codec) marshalBinaryBarePBBindings(pbm PBMessager) ([]byte, error) {
	pbo, err := pbm.ToPBMessage(cdc)
	if err != nil {
		return nil, err
	}
	bz, err := proto.Marshal(pbo)
	return bz, err
}

// Panics if error.
func (cdc *Codec) MustMarshalBinaryBare(o interface{}) []byte {
	bz, err := cdc.MarshalBinaryBare(o)
	if err != nil {
		panic(err)
	}
	return bz
}

// MarshalBinaryInterfaceBare encodes the registered object
// wrapped with google.protobuf.Any.
func (cdc *Codec) MarshalBinaryInterfaceBare(o interface{}) ([]byte, error) {
	cdc.doAutoseal()

	// o cannot be nil, otherwise we don't know what type it is.
	if o == nil {
		return nil, errors.New("MarshalBinaryInterfaceBare() requires non-nil argument")
	}

	// Dereference value if pointer.
	var rv, _, _ = maybeDerefValue(reflect.ValueOf(o))
	var rt = rv.Type()

	// rv cannot be an interface.
	if rv.Kind() == reflect.Interface {
		return nil, errors.New("MarshalBinaryInterfaceBare() requires registered concrete type")
	}

	// Make a temporary interface var, to contain the value of o.
	var ivar interface{} = rv.Interface()
	var iinfo *TypeInfo
	iinfo, err := cdc.getTypeInfoWLock(rt)
	if err != nil {
		return nil, err
	}

	// Encode as interface.
	buf := new(bytes.Buffer)
	err = cdc.encodeReflectBinaryInterface(buf, iinfo, reflect.ValueOf(&ivar).Elem(), FieldOptions{}, true)
	if err != nil {
		return nil, err
	}
	bz := buf.Bytes()

	return bz, nil
}

// Panics if error.
func (cdc *Codec) MustMarshalBinaryInterfaceBare(o interface{}) []byte {
	bz, err := cdc.MarshalBinaryInterfaceBare(o)
	if err != nil {
		panic(err)
	}
	return bz
}

//----------------------------------------
// Unmarshal* methods

// Like UnmarshalBinaryBare, but will first decode the byte-length prefix.
// UnmarshalBinaryLengthPrefixed will panic if ptr is a nil-pointer.
// Returns an error if not all of bz is consumed.
func (cdc *Codec) UnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{}) error {
	if len(bz) == 0 {
		return errors.New("unmarshalBinaryLengthPrefixed cannot decode empty bytes")
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
		_ = errors.Errorf( //nolint:errcheck
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
	return n, err
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
	cdc.doAutoseal()

	if cdc.usePBBindings {
		pbm, ok := ptr.(PBMessager)
		if ok {
			return cdc.unmarshalBinaryBarePBBindings(bz, pbm)
		} else {
			// Fall back to using reflection for native primitive types.
		}
	}

	return cdc.unmarshalBinaryBareReflect(bz, ptr)
}

// Use reflection.
func (cdc *Codec) unmarshalBinaryBareReflect(bz []byte, ptr interface{}) error {
	rv := reflect.ValueOf(ptr)
	if rv.Kind() != reflect.Ptr {
		return ErrNoPointer
	}
	rv = rv.Elem()
	rt := rv.Type()
	info, err := cdc.getTypeInfoWLock(rt)
	if err != nil {
		return err
	}

	// See if we need to read the typ3 encoding of an implicit struct.
	//
	// If the dest ptr is an interface, it is assumed that the object is
	// wrapped in a google.protobuf.Any object, so skip this step.
	//
	// See corresponding encoding message in this file, and also
	// binary-decode.
	var bare = true
	var nWrap int
	if !info.IsStructOrUnpacked(FieldOptions{}) &&
		len(bz) > 0 &&
		(rv.Kind() != reflect.Interface) {
		var (
			fnum      uint32
			typ       Typ3
			nFnumTyp3 int
		)
		fnum, typ, nFnumTyp3, err = decodeFieldNumberAndTyp3(bz)
		if err != nil {
			return errors.Wrap(err, "could not decode field number and type")
		}
		if fnum != 1 {
			return fmt.Errorf("expected field number: 1; got: %v", fnum)
		}
		typWanted := info.GetTyp3(FieldOptions{})
		if typ != typWanted {
			return fmt.Errorf("expected field type %v for # %v of %v, got %v",
				typWanted, fnum, info.Type, typ)
		}

		slide(&bz, &nWrap, nFnumTyp3)
		// "bare" is ignored when primitive, byteslice, bytearray.
		// When typ3 != ByteLength, then typ3 is one of Typ3Varint, Typ38Byte,
		// Typ34Byte; and they are all primitive.
		bare = false
	}

	// Decode contents into rv.
	n, err := cdc.decodeReflectBinary(bz, info, rv, FieldOptions{BinFieldNum: 1}, bare, 0)
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

// Use pbbindings.
func (cdc *Codec) unmarshalBinaryBarePBBindings(bz []byte, pbm PBMessager) error {
	pbo := pbm.EmptyPBMessage(cdc)
	err := proto.Unmarshal(bz, pbo)
	if err != nil {
		return err
	}
	err = pbm.FromPBMessage(cdc, pbo)
	if err != nil {
		return err
	}
	return nil
}

// Panics if error.
func (cdc *Codec) MustUnmarshalBinaryBare(bz []byte, ptr interface{}) {
	err := cdc.UnmarshalBinaryBare(bz, ptr)
	if err != nil {
		panic(err)
	}
}

// UnmarshalBinaryAny decodes the registered object
// from the Any fields.
func (cdc *Codec) UnmarshalBinaryAny(typeURL string, value []byte, ptr interface{}) (err error) {
	cdc.doAutoseal()

	rv := reflect.ValueOf(ptr)
	if rv.Kind() != reflect.Ptr {
		return ErrNoPointer
	}
	rv = rv.Elem()
	_, err = cdc.decodeReflectBinaryAny(typeURL, value, rv, FieldOptions{})
	return
}

func (cdc *Codec) MustUnmarshalBinaryAny(typeURL string, value []byte, ptr interface{}) {
	err := cdc.UnmarshalBinaryAny(typeURL, value, ptr)
	if err != nil {
		panic(err)
	}
	return
}

func (cdc *Codec) MarshalJSON(o interface{}) ([]byte, error) {
	cdc.doAutoseal()

	rv := reflect.ValueOf(o)
	if rv.Kind() == reflect.Invalid {
		return []byte("null"), nil
	}
	rt := rv.Type()
	w := new(bytes.Buffer)
	info, err := cdc.getTypeInfoWLock(rt)
	if err != nil {
		return nil, err
	}
	if err = cdc.encodeReflectJSON(w, info, rv, FieldOptions{}); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (cdc *Codec) MarshalJSONInterface(o interface{}) ([]byte, error) {
	// o cannot be nil, otherwise we don't know what type it is.
	if o == nil {
		return nil, errors.New("MarshalJSONInterface() requires non-nil argument")
	}

	// Dereference value if pointer.
	var rv = reflect.ValueOf(o)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	var rt = rv.Type()

	// rv cannot be an interface.
	if rv.Kind() == reflect.Interface {
		return nil, errors.New("MarshalJSONInterface() requires registered concrete type")
	}

	// Make a temporary interface var, to contain the value of o.
	var ivar interface{} = rv.Interface()
	var iinfo *TypeInfo
	iinfo, err := cdc.getTypeInfoWLock(rt)
	if err != nil {
		return nil, err
	}

	// Encode as interface.
	buf := new(bytes.Buffer)
	err = cdc.encodeReflectJSONInterface(buf, iinfo, reflect.ValueOf(&ivar).Elem(), FieldOptions{})
	if err != nil {
		return nil, err
	}
	bz := buf.Bytes()

	return bz, nil
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
	cdc.doAutoseal()
	if len(bz) == 0 {
		return errors.New("cannot decode empty bytes")
	}

	rv := reflect.ValueOf(ptr)
	if rv.Kind() != reflect.Ptr {
		return errors.New("expected a pointer")
	}
	rv = rv.Elem()
	rt := rv.Type()
	info, err := cdc.getTypeInfoWLock(rt)
	if err != nil {
		return err
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

//----------------------------------------
// Other

// NOTE: do not modify the result.
func RegisterPackage(pi *pkg.Package) *Package {
	gcdc.RegisterPackage(pi)
	return pi
}

func NewPackage(gopkg string, p3pkg string, dirname string) *Package {
	return pkg.NewPackage(gopkg, p3pkg, dirname)
}

// NOTE: duplicated in pkg/pkg.go
func GetCallersDirname() string {
	var dirname = "" // derive from caller.
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		panic("could not get caller to derive caller's package directory")
	}
	dirname = path.Dir(filename)
	if filename == "" || dirname == "" {
		panic("could not derive caller's package directory")
	}
	return dirname
}

//----------------------------------------
// Object

// All concrete types must implement the Object interface for genproto
// bindings.  They are generated automatically by genproto/bindings.go
type Object interface {
	GetTypeURL() string
}

// TODO: this does need the cdc receiver,
// as it should also work for non-pbbindings-optimized types.
// Returns the default type url for the given concrete type.
// XXX Unstable API.
func (cdc *Codec) GetTypeURL(o interface{}) string {
	if obj, ok := o.(Object); ok {
		return obj.GetTypeURL()
	}
	switch o.(type) {
	case time.Time, *time.Time, *timestamppb.Timestamp:
		return "/google.protobuf.Timestamp"
	case time.Duration, *time.Duration, *durationpb.Duration:
		return "/google.protobuf.Duration"
	}
	rv := reflect.ValueOf(o)
	switch rv.Kind() {
	case reflect.String:
		return "/google.protobuf.StringValue"
	case reflect.Int64, reflect.Int:
		return "/google.protobuf.Int64Value"
	case reflect.Int32:
		return "/google.protobuf.Int32Value"
	case reflect.Int16:
		return "/google.protobuf.Int32Value"
	case reflect.Int8:
		return "/google.protobuf.Int32Value"
	case reflect.Uint64, reflect.Uint:
		return "/google.protobuf.UInt64Value"
	case reflect.Uint32:
		return "/google.protobuf.UInt32Value"
	case reflect.Uint16:
		return "/google.protobuf.UInt32Value"
	case reflect.Uint8:
		return "/google.protobuf.UInt32Value"
	case reflect.Bool:
		return "/google.protobuf.BoolValue"
	case reflect.Array:
		if rv.Elem().Kind() == reflect.Uint8 {
			return "/google.protobuf.BytesValue"
		} else {
			panic("not yet supported")
		}
	case reflect.Slice:
		if rv.Elem().Kind() == reflect.Uint8 {
			return "/google.protobuf.BytesValue"
		} else {
			panic("not yet supported")
		}
	default:
		panic("not yet implemented")
	}
}

//----------------------------------------

// Methods generated by genproto/bindings.go for faster encoding.
type PBMessager interface {
	ToPBMessage(*Codec) (proto.Message, error)
	EmptyPBMessage(*Codec) proto.Message
	FromPBMessage(*Codec, proto.Message) error
}
