package wire

// XXX Add JSON again.
// XXX Check for custom marshal/unmarshal functions.
// XXX Scan the codebase for unwraps and double check that they implement above.

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"reflect"
	"time"
)

/*

Wire is an encoding library that can handle interfaces (like protobuf
"oneof") well.  This is achieved by prefixing bytes before each "concrete
type".

A concrete type is some non-interface value (generally a struct) which
implements the interface to be (de)serialized. Not all structures need to
be registered as concrete types -- only when they will be stored in
interface type fields (or interface type slices) do they need to be
registered.

//----------------------------------------
// Registering types

All interfaces and the concrete types that implement them must be registered.

> wire.RegisterInterface((*MyInterface1)(nil), nil)
> wire.RegisterInterface((*MyInterface2)(nil), nil)
> wire.RegisterConcrete(MyStruct1{}, "com.tendermint/MyStruct1", nil)
> wire.RegisterConcrete(&MyStruct2{}, "com.tendermint/MyStruct2", nil)

Notice that an interface is represented by a nil pointer.

Structures that must be deserialized as pointer values must be registered
with a pointer value as well.  It's OK to (de)serialize such structures in
non-pointer (value) form, but when deserializing such structures into an
interface field, they will always be deserialized as pointers.

//----------------------------------------
// How it works

All registered concrete types are encoded with leading 4 bytes (called
"prefix bytes"), even when it's not held in an interface field/element.  In
this way, Wire ensures that concrete types (almost) always have the same
canonical representation.  The first byte of the prefix bytes must not be a
zero byte, so there are 2**(8*4)-2**(8*3) possible values.

When there are 4096 types registered at once, the probability of there
being a conflict is ~ 0.2%. See https://instacalc.com/51189 for estimation.
This is assuming that all registered concrete types have unique natural
names (e.g. prefixed by a unique entity name such as "com.tendermint/", and
not "mined/grinded" to produce a particular sequence of "prefix bytes").

TODO Update instacalc.com link with 255/256 since 0x00 is an escape.

Do not mine/grind to produce a particular sequence of prefix bytes, and avoid
using dependencies that do so.

Since 4 bytes are not sufficient to ensure no conflicts, sometimes it is
necessary to prepend more than the 4 prefix bytes for disambiguation.  Like the
prefix bytes, the disambiguation bytes are also computed from the registered
name of the concrete type.  There are 3 disambiguation bytes, and in binary
form they always precede the prefix bytes.  The first byte of the
disambiguation bytes must not be a zero byte, so there are 2**(8*3)-2**(8*2)
possible values.

// Sample Wire encoded binary bytes with 4 prefix bytes.
> [0xBB 0x9C 0x83 0xDD] [...]

// Sample Wire encoded binary bytes with 3 disambiguation bytes and 4
// prefix bytes.
> 0x00 <0xA8 0xFC 0x54> [0xBB 0x9C 0x83 0xDD] [...]

The prefix bytes never start with a zero byte, so the disambiguation bytes
are escaped with 0x00.

Notice that the 4 prefix bytes always immediately precede the binary
encoding of the concrete type.

//----------------------------------------
// Computing prefix bytes

To compute the prefix bytes, we take `hash := sha256(concreteTypeName)`,
and drop the leading 0x00 bytes.

> hash := sha256("com.tendermint.consensus/MyConcreteName")
> hex.EncodeBytes(hash) // 0x{00 00 BB 9C 83 DD 00 A8 FC 54 4C 03 ...} (example)

In the example above, hash has two leading 0x00 bytes, so we drop them.

> rest = dropLeadingZeroBytes(hash) // 0x{BB 9C 83 DD 00 A8 FC 54 4C 03 ...}
> prefix = rest[0:4]
> rest = dropLeadingZeroBytes(rest[4:])
> disamb = rest[0:3]

The first 4 bytes are called the "name bytes" (in square brackets).
The next 3 bytes are called the "disambiguation bytes" (in angle brackets).

> [0xBB 0x9C 9x83 9xDD] <0xA8 0xFC 0x54> ...

*/

type PrefixBytes [4]byte
type DisambBytes [3]byte
type DisfixBytes [7]byte // Disamb+Prefix

type TypeInfo struct {
	Type reflect.Type // Interface type.
	InterfaceInfo
	ConcreteInfo
}

type InterfaceInfo struct {
	NilValue reflect.Value
	InterfaceOptions
}

type InterfaceOptions struct {
	Priority           []string // Disamb priority.
	AlwaysDisambiguate bool     // If true, include disamb for all types.
}

type ConcreteInfo struct {
	PointerPreferred bool        // Deserialize to pointer type if possible.
	Registered       bool        // Manually regsitered.
	Name             string      // Ignored if !Registered.
	Prefix           PrefixBytes // Ignored if !Registered.
	Disamb           DisambBytes // Ignored if !Registered.
	Fields           []FieldInfo // If a struct.
	ConcreteOptions
}

type ConcreteOptions struct {
}

type FieldInfo struct {
	Type         reflect.Type // Struct field type
	Index        int          // Struct field index
	ZeroProto    interface{}  // Prototype zero value object.
	FieldOptions              // Encoding options
}

type FieldOptions struct {
	JSONName      string // (JSON) field name
	JSONOmitEmpty bool   // (JSON) omitempty
	BinVarint     bool   // (Binary) Use length-prefixed encoding for (u)int64.
	Unsafe        bool   // e.g. if this field is a float.
}

// This function should be used to register all interfaces that will be
// encoded/decoded by go-wire.
// Usage:
// `wire.RegisterInterface((*MyInterface1)(nil), nil)`
func (cdc *Codec) RegisterInterface(ptr interface{}, opts *InterfaceOptions) {

	// Get reflect.Type from ptr.
	rt := getTypeFromPointer(ptr)
	if rt.Kind() != reflect.Interface {
		panic(fmt.Sprintf("RegisterInterface expects an interface, got %v", rt))
	}

	// Construct InterfaceInfo
	var info = new(TypeInfo)
	info.Type = rt
	info.NilValue = reflect.ValueOf(reflect.Zero(rt))
	if opts != nil {
		info.InterfaceOptions = *opts
	}

	// XXX
	// For each registered concrete type crt:
	//   If crt (pointer if pointer-preferred) implements interface:
	//     If crt doesn't exist in prio list:
	//       If crt prefix bytes conflicts with any InterfaceType.Impls:
	//         return error.
	//     Add crt to InterfaceType.Impls

	// Finally, register.
	cdc.setTypeInfo_wlock(info)
}

// This function should be used to register concrete types that will appear in
// interface fields/elements to be encoded/decoded by go-wire.
// Usage:
// `wire.RegisterConcrete(MyStruct1{}, "com.tendermint/MyStruct1", nil)`
func (cdc *Codec) RegisterConcrete(o interface{}, name string, opts *ConcreteOptions) {

	var pointerPreferred bool

	// Get reflect.Type.
	rt := reflect.TypeOf(o)
	if rt.Kind() == reflect.Interface {
		panic(fmt.Sprintf("expected a non-interface: %v", rt))
	}
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		if rt.Kind() == reflect.Ptr {
			// We can encode/decode pointer-pointers, but not register them.
			panic(fmt.Sprintf("registering pointer-pointers not yet supported: *%v", rt))
		}
		if rt.Kind() == reflect.Interface {
			panic(fmt.Sprintf("registering interface-pointers not yet supported: *%v", rt))
		}
		pointerPreferred = true
	}

	// Construct ConcreteInfo
	var info = new(TypeInfo)
	info.Type = rt
	info.PointerPreferred = pointerPreferred
	info.Registered = true
	info.Name = name
	info.Prefix, info.Disamb = nameToPrefix(name)
	info.Fields = cdc.parseFieldInfos(rt)
	if opts != nil {
		info.ConcreteOptions = *opts
	}

	// XXX
	// For each registered interface:
	//   If crt (pointer if pointer-preferred) implements interface:
	//     If crt doesn't exist in prio list:
	//       If crt prefix bytes conflicts with any InterfaceType.Impls:
	//         return error.
	//     Add crt to InterfaceType.Impls

	// Actually register the interface.
	cdc.setTypeInfo_wlock(info)
}

//----------------------------------------
// constants

var timeType = reflect.TypeOf(time.Time{})

const RFC3339Millis = "2006-01-02T15:04:05.000Z" // forced microseconds

//----------------------------------------
// cdc.decodeReflectBinary

// CONTRACT: rv.CanAddr() is true.
// CONTRACT: caller holds cdc.mtx.
func (cdc *Codec) decodeReflectBinary(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) (n int, err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}

	var _n int

	// Transparently deal with pointer.
	// This works for pointer-pointers.
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			newPtr := reflect.New(rv.Type()).Elem()
			rv.Set(newPtr)
		}
		rv = rv.Elem()
	}

	switch info.Type.Kind() {

	//----------------------------------------
	// Complex

	case reflect.Array:
		return cdc.decodeReflectBinaryArray(bz, info, rv, opts)

	case reflect.Interface:
		return cdc.decodeReflectBinaryInterface(bz, info, rv, opts)

	case reflect.Slice:
		return cdc.decodeReflectBinarySlice(bz, info, rv, opts)

	case reflect.Struct:
		return cdc.decodeReflectBinaryStruct(bz, info, rv, opts)

	//----------------------------------------
	// Signed

	case reflect.Int64:
		var num int64
		if opts.BinVarint {
			num, _n, err = DecodeVarint(bz)
			if slide(bz, &bz, &n, _n) && err != nil {
				return
			}
			rv.SetInt(num)
		} else {
			num, _n, err = DecodeInt64(bz)
			if slide(bz, &bz, &n, _n) && err != nil {
				return
			}
			rv.SetInt(num)
		}
		return

	case reflect.Int32:
		var num int32
		num, _n, err = DecodeInt32(bz)
		if slide(bz, &bz, &n, _n) && err != nil {
			return
		}
		rv.SetInt(int64(num))
		return

	case reflect.Int16:
		var num int16
		num, _n, err = DecodeInt16(bz)
		if slide(bz, &bz, &n, _n) && err != nil {
			return
		}
		rv.SetInt(int64(num))
		return

	case reflect.Int8:
		var num int8
		num, _n, err = DecodeInt8(bz)
		if slide(bz, &bz, &n, _n) && err != nil {
			return
		}
		rv.SetInt(int64(num))
		return

	case reflect.Int:
		var num int64
		num, _n, err = DecodeVarint(bz)
		if slide(bz, &bz, &n, _n) && err != nil {
			return
		}
		rv.SetInt(num)
		return

	//----------------------------------------
	// Unsigned

	case reflect.Uint64:
		var num uint64
		if opts.BinVarint {
			num, _n, err = DecodeUvarint(bz)
			if slide(bz, &bz, &n, _n) && err != nil {
				return
			}
			rv.SetUint(num)
		} else {
			num, _n, err = DecodeUint64(bz)
			if slide(bz, &bz, &n, _n) && err != nil {
				return
			}
			rv.SetUint(num)
		}
		return

	case reflect.Uint32:
		var num uint32
		num, _n, err = DecodeUint32(bz)
		if slide(bz, &bz, &n, _n) && err != nil {
			return
		}
		rv.SetUint(uint64(num))
		return

	case reflect.Uint16:
		var num uint16
		num, _n, err = DecodeUint16(bz)
		if slide(bz, &bz, &n, _n) && err != nil {
			return
		}
		rv.SetUint(uint64(num))
		return

	case reflect.Uint8:
		var num uint8
		num, _n, err = DecodeUint8(bz)
		if slide(bz, &bz, &n, _n) && err != nil {
			return
		}
		rv.SetUint(uint64(num))
		return

	case reflect.Uint:
		var num uint64
		num, _n, err = DecodeUvarint(bz)
		if slide(bz, &bz, &n, _n) && err != nil {
			return
		}
		rv.SetUint(num)
		return

	//----------------------------------------
	// Misc.

	case reflect.Bool:
		var b bool
		b, _n, err = DecodeBool(bz)
		if slide(bz, &bz, &n, _n) && err != nil {
			return
		}
		rv.SetBool(b)
		return

	case reflect.Float64:
		var f float64
		if !opts.Unsafe {
			err = fmt.Errorf("Float support requires `wire:\"unsafe\"`.")
			return
		}
		f, _n, err = DecodeFloat64(bz)
		if slide(bz, &bz, &n, _n) && err != nil {
			return
		}
		rv.SetFloat(f)
		return

	case reflect.Float32:
		var f float32
		if !opts.Unsafe {
			err = fmt.Errorf("Float support requires `wire:\"unsafe\"`.")
			return
		}
		f, _n, err = DecodeFloat32(bz)
		if slide(bz, &bz, &n, _n) && err != nil {
			return
		}
		rv.SetFloat(float64(f))
		return

	case reflect.String:
		var str string
		str, _n, err = DecodeString(bz)
		if slide(bz, &bz, &n, _n) && err != nil {
			return
		}
		rv.SetString(str)
		return

	default:
		panic(fmt.Sprintf("unknown field type %v", info.Type.Kind()))
	}

}

// CONTRACT: rv.CanAddr() is true.
// CONTRACT: caller holds cdc.mtx.
func (cdc *Codec) decodeReflectBinaryInterface(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) (n int, err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}
	if !rv.IsNil() {
		// This is very tricky.
		err = fmt.Errorf("Decoding to a non-nil interface is not supported yet")
		return
	}

	// Read disambiguation / prefix bytes.
	disfix, hasDisamb, prefix, hasPrefix, isNil, _n, err := decodeDisambPrefixBytes(bz)
	if slide(bz, &bz, &n, _n) && err != nil {
		return
	}

	// Special case for nil
	if isNil {
		rv.Set(info.NilValue)
		return
	}

	// Get concrete type info.
	var cinfo *TypeInfo
	if hasDisamb {
		cinfo, err = cdc.getTypeInfoFromDisfix_rlock(disfix)
	} else if hasPrefix {
		cinfo, err = cdc.getTypeInfoFromPrefix_rlock(prefix)
	} else {
		err = fmt.Errorf("Expected disambiguation or prefix bytes.")
	}
	if err != nil {
		return
	}

	// Construct new concrete type.
	var crv reflect.Value
	if cinfo.PointerPreferred {
		crv = reflect.New(cinfo.Type)
	} else {
		crv = reflect.New(cinfo.Type).Elem()
	}
	// NOTE: this should succeed because it must be true
	// for both Interface and Concrete to be registered.
	rv.Set(crv)

	// Read into crv.
	_n, err = cdc.decodeReflectBinary(bz, cinfo, crv, opts)
	slide(bz, &bz, &n, _n)
	return
}

// CONTRACT: rv.CanAddr() is true.
// CONTRACT: caller holds cdc.mtx.
func (cdc *Codec) decodeReflectBinaryArray(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) (n int, err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}
	ert := info.Type.Elem()
	length := info.Type.Len()
	_n := 0

	switch ert.Kind() {

	case reflect.Uint8: // Special case: byte array
		if len(bz) < length {
			return 0, fmt.Errorf("Insufficient bytes to decode [%v]byte.", length)
		}
		reflect.Copy(rv, reflect.ValueOf(bz[0:length]))
		return

	default: // General case.
		var einfo *TypeInfo
		einfo, err = cdc.getTypeInfo_wlock(ert)
		if err != nil {
			return
		}
		for i := 0; i < length; i++ {
			erv := rv.Index(i)
			_n, err = cdc.decodeReflectBinary(bz, einfo, erv, opts)
			if slide(bz, &bz, &n, _n) && err != nil {
				return
			}
		}
		return
	}
}

// CONTRACT: rv.CanAddr() is true.
// CONTRACT: caller holds cdc.mtx.
func (cdc *Codec) decodeReflectBinarySlice(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) (n int, err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}
	ert := info.Type.Elem()
	_n := 0

	switch ert.Kind() {

	case reflect.Uint8: // Special case: byte slice
		var byteslice []byte
		byteslice, _n, err = DecodeByteSlice(bz)
		if slide(bz, &bz, &n, _n) && err != nil {
			return
		}
		rv.Set(reflect.ValueOf(byteslice))
		return

	default: // General case.

		// Read length.
		var length int64
		length, _n, err = DecodeVarint(bz)
		if slide(bz, &bz, &n, _n) && err != nil {
			return
		}

		// Read into a new slice.
		var srv = reflect.MakeSlice(ert, 0, int(length))
		var einfo *TypeInfo
		einfo, err = cdc.getTypeInfo_wlock(ert)
		if err != nil {
			return
		}
		for i := 0; i < int(length); i++ {
			erv := srv.Index(i)
			_n, err = cdc.decodeReflectBinary(bz, einfo, erv, opts)
			if slide(bz, &bz, &n, _n) && err != nil {
				return
			}
			srv = reflect.AppendSlice(srv, erv)
		}

		// TODO do we need this extra step?
		rv.Set(srv)
		return
	}
}

// CONTRACT: rv.CanAddr() is true.
// CONTRACT: caller holds cdc.mtx.
func (cdc *Codec) decodeReflectBinaryStruct(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) (n int, err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}
	_n := 0

	switch info.Type {

	case timeType: // Special case: time.Time
		var t time.Time
		t, _n, err = DecodeTime(bz)
		if slide(bz, &bz, &n, _n) && err != nil {
			return
		}
		rv.Set(reflect.ValueOf(t))
		return

	default:
		for _, field := range info.Fields {
			var finfo *TypeInfo
			finfo, err = cdc.getTypeInfo_wlock(field.Type)
			if err != nil {
				return
			}
			frv := rv.Field(field.Index)
			_n, err = cdc.decodeReflectBinary(bz, finfo, frv, field.FieldOptions)
			if slide(bz, &bz, &n, _n) && err != nil {
				return
			}
		}
		return
	}
}

//----------------------------------------
// cdc.encodeReflectBinary

// CONTRACT: caller holds cdc.mtx.
func (cdc *Codec) encodeReflectBinary(w io.Writer, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {

	// Dereference pointer transparently.
	// This works for pointer-pointers.
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	switch info.Type.Kind() {

	//----------------------------------------
	// Complex

	case reflect.Array:
		err = cdc.encodeReflectBinaryArray(w, info, rv, opts)

	case reflect.Interface:
		err = cdc.encodeReflectBinaryInterface(w, info, rv, opts)

	case reflect.Slice:
		err = cdc.encodeReflectBinarySlice(w, info, rv, opts)

	case reflect.Struct:
		err = cdc.encodeReflectBinaryStruct(w, info, rv, opts)

	//----------------------------------------
	// Signed

	case reflect.Int64:
		if opts.BinVarint {
			err = EncodeVarint(w, rv.Int())
		} else {
			err = EncodeInt64(w, rv.Int())
		}

	case reflect.Int32:
		err = EncodeInt32(w, int32(rv.Int()))

	case reflect.Int16:
		err = EncodeInt16(w, int16(rv.Int()))

	case reflect.Int8:
		err = EncodeInt8(w, int8(rv.Int()))

	case reflect.Int:
		err = EncodeVarint(w, rv.Int())

	//----------------------------------------
	// Unsigned

	case reflect.Uint64:
		if opts.BinVarint {
			err = EncodeUvarint(w, rv.Uint())
		} else {
			err = EncodeUint64(w, rv.Uint())
		}

	case reflect.Uint32:
		err = EncodeUint32(w, uint32(rv.Uint()))

	case reflect.Uint16:
		err = EncodeUint16(w, uint16(rv.Uint()))

	case reflect.Uint8:
		err = EncodeUint8(w, uint8(rv.Uint()))

	case reflect.Uint:
		err = EncodeUvarint(w, rv.Uint())

	//----------------------------------------
	// Misc

	case reflect.Bool:
		err = EncodeBool(w, rv.Bool())

	case reflect.Float64:
		if !opts.Unsafe {
			err = fmt.Errorf("Wire float* support requires `wire:\"unsafe\"`.")
			return
		}
		err = EncodeFloat64(w, rv.Float())

	case reflect.Float32:
		if !opts.Unsafe {
			err = fmt.Errorf("Wire float* support requires `wire:\"unsafe\"`.")
			return
		}
		err = EncodeFloat32(w, float32(rv.Float()))

	case reflect.String:
		err = EncodeString(w, rv.String())

	default:
		panic(fmt.Sprintf("unknown field type %v", info.Type.Kind()))
	}

	return
}

// CONTRACT: caller holds cdc.mtx.
func (cdc *Codec) encodeReflectBinaryInterface(w io.Writer, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {

	if rv.IsNil() {
		_, err = w.Write([]byte{0x00, 0x00, 0x00, 0x00})
		return
	}

	crv := rv.Elem() // concrete reflection value

	// Dereference pointer transparently.
	// This also works for pointer-pointers.
	// NOTE: Encoding pointer-pointers only work for no-method interfaces like
	// `interface{}`.
	for crv.Kind() == reflect.Ptr {
		crv = crv.Elem()
		if crv.Kind() == reflect.Interface {
			err = fmt.Errorf("Unexpected interface-pointer of type *%v for registered interface %v. Not supported yet.", crv.Type(), info.Type)
			return
		}
		if !crv.IsValid() {
			err = fmt.Errorf("Illegal nil-pointer of type %v for registered interface %v. "+
				"For compatibility with other languages, nil-pointer interface values are forbidden.", crv.Type(), info.Type)
			return
		}
	}

	crt := crv.Type() // non-pointer non-interface concrete type

	// Get *TypeInfo for concrete type.
	var cinfo *TypeInfo
	cinfo, err = cdc.getTypeInfo_wlock(crt)
	if err != nil {
		return
	}
	if !cinfo.Registered {
		err = fmt.Errorf("Cannot encode unregistered concrete type %v.", crt)
		return
	}

	// Write the disambiguation bytes if needed.
	if info.AlwaysDisambiguate {
		_, err = w.Write(cinfo.Disamb[:])
		if err != nil {
			return
		}
	}

	// Write the prefix bytes.
	_, err = w.Write(cinfo.Prefix[:])
	if err != nil {
		return
	}

	err = cdc.encodeReflectBinary(w, cinfo, crv, opts)
	return
}

// CONTRACT: caller holds cdc.mtx.
func (cdc *Codec) encodeReflectBinaryArray(w io.Writer, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {
	ert := info.Type.Elem()
	length := info.Type.Len()

	switch ert.Kind() {

	case reflect.Uint8: // Special case: byte array
		if rv.CanAddr() {
			bz := rv.Slice(0, length).Bytes()
			_, err = w.Write(bz)
			return
		} else {
			buf := make([]byte, length)
			reflect.Copy(reflect.ValueOf(buf), rv) // XXX: looks expensive!
			_, err = w.Write(buf)
			return
		}

	default:
		var einfo *TypeInfo
		einfo, err = cdc.getTypeInfo_wlock(ert)
		if err != nil {
			return
		}
		for i := 0; i < length; i++ {
			erv := rv.Index(i)
			err = cdc.encodeReflectBinary(w, einfo, erv, opts)
			if err != nil {
				return err
			}
		}
		return
	}
}

// CONTRACT: caller holds cdc.mtx.
func (cdc *Codec) encodeReflectBinarySlice(w io.Writer, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {
	ert := info.Type.Elem()

	switch ert.Kind() {

	case reflect.Uint8: // Special case: byte slice
		byteslice := rv.Bytes()
		err = EncodeByteSlice(w, byteslice)
		return

	default:
		// Write length
		length := rv.Len()
		err = EncodeVarint(w, int64(length))
		if err != nil {
			return err
		}

		// Write elems
		var einfo *TypeInfo
		einfo, err = cdc.getTypeInfo_wlock(ert)
		if err != nil {
			return
		}
		for i := 0; i < length; i++ {
			erv := rv.Index(i)
			err = cdc.encodeReflectBinary(w, einfo, erv, opts)
			if err != nil {
				return
			}
		}
		return
	}
}

// CONTRACT: caller holds cdc.mtx.
func (cdc *Codec) encodeReflectBinaryStruct(w io.Writer, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {

	switch info.Type {

	case timeType: // Special case: time.Time
		err = EncodeTime(w, rv.Interface().(time.Time))
		return

	default:
		for _, field := range info.Fields {
			var finfo *TypeInfo
			finfo, err = cdc.getTypeInfo_wlock(field.Type)
			if err != nil {
				return
			}
			frv := rv.Field(field.Index)
			err = cdc.encodeReflectBinary(w, finfo, frv, field.FieldOptions)
			if err != nil {
				return
			}
		}
		return
	}

}

//----------------------------------------
// Misc.

func getTypeFromPointer(ptr interface{}) reflect.Type {
	rt := reflect.TypeOf(ptr)
	if rt.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("expected pointer, got %v", rt))
	}
	return rt.Elem()
}

func checkUnsafe(field FieldInfo) {
	if field.Unsafe {
		return
	}
	switch field.Type.Kind() {
	case reflect.Float32, reflect.Float64:
		panic("floating point types are unsafe for go-wire")
	}
}

func nameToPrefix(name string) (pb PrefixBytes, db DisambBytes) {
	hasher := sha256.New()
	hasher.Write([]byte(name))
	bz := hasher.Sum(nil)
	for bz[0] == 0x00 {
		bz = bz[1:]
	}
	copy(pb[:], bz[0:4])
	bz = bz[4:]
	for bz[0] == 0x00 {
		bz = bz[1:]
	}
	copy(db[:], bz[0:3])
	return
}

func toDisfix(pb PrefixBytes, db DisambBytes) (df DisfixBytes) {
	copy(df[0:3], db[0:3])
	copy(df[3:7], pb[0:4])
	return
}

func decodeDisambPrefixBytes(bz []byte) (df DisfixBytes, hasDb bool, pb PrefixBytes, hasPb bool, isNil bool, n int, err error) {
	// Validate
	if len(bz) < 4 {
		err = fmt.Errorf("EOF while reading prefix bytes.")
		return // hasPb = false
	}
	if bz[0] == 0x00 {
		// Special case: nil
		if bytes.Equal(bz[1:3], []byte{0x00, 0x00, 0x00}) {
			isNil = true
			n = 4
			return
		}
		// Validate
		if len(bz) < 8 {
			err = fmt.Errorf("EOF while reading disamb bytes.")
			return // hasPb = false
		}
		copy(df[0:7], bz[1:8])
		copy(pb[0:4], bz[4:8])
		hasDb = true
		hasPb = true
		n = 8
		return
	} else {
		// General case with no disambiguation
		copy(pb[0:4], bz[0:4])
		hasDb = false
		hasPb = true
		n = 4
		return
	}
}

// CONTRACT: by the time this is called, len(bz) >= _n
// Returns true so you can write one-liners.
func slide(bz []byte, bz2 *[]byte, n *int, _n int) bool {
	*bz2 = bz[_n:]
	*n += _n
	return true
}
