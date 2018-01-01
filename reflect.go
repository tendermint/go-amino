package wire

// XXX Add JSON again.
// XXX Check for custom marshal/unmarshal functions.
// XXX Scan the codebase for unwraps and double check that they implement above.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"strings"
	"sync"
	"time"

	cmn "github.com/tendermint/tmlibs/common"
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
	ZeroValue    interface{}  // Prototype zero value object.
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
func RegisterInterface(ptr interface{}, opts *InterfaceOptions) {

	// Get reflect.Type from ptr.
	rt := getTypeFromPointer(ptr)
	if rt.Kind() != reflect.Interface {
		panic(fmt.Sprintf("RegisterInterface expects an interface, got %v", rt))
	}

	// Construct InterfaceInfo
	var info InterfaceInfo
	info.Type = rt
	info.InterfaceOptions = *opts

	// Finally, register.
	setInterfaceInfo(rt, &info)
}

// This function should be used to register concrete types that will appear in
// interface fields/elements to be encoded/decoded by go-wire.
// Usage:
// `wire.RegisterConcrete(MyStruct1{}, "com.tendermint/MyStruct1", nil)`
func RegisterConcrete(o interface{}, name string, opts *ConcreteOptions) {

	var pointerPreferred bool

	// Get reflect.Type.
	rt := reflect.TypeOf(o)
	if rt.Kind() == reflect.Ptr {
		pointerPreferred = true
		rt = rt.Elem()
	}
	if rt.Kind() == reflect.Interface {
		panic(fmt.Sprintf("RegisterConcrete expects a non-interface, got %v", rt))
	}

	// Construct ConcreteInfo
	var info ConcreteInfo
	info.Type = rt
	info.PointerPreferred = pointerPreferred
	info.Registered = true
	info.Name = name
	info.Prefix, info.Disamb = nameToPrefix(name)
	info.Fields = parseFieldInfos(rt)
	info.ConcreteOptions = *opts

	// Actually register the interface.
	setConcreteInfo(rt, info)
}

//----------------------------------------
// constants

var timeType = reflect.TypeOf(time.Time{})

const RFC3339Millis = "2006-01-02T15:04:05.000Z" // forced microseconds

//----------------------------------------
// decodeReflectBinary

func decodeReflectBinary(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) (n int, err error) {
	var _n int

	switch info.Type.Kind() {

	//----------------------------------------
	// Complex

	case reflect.Array:
		return decodeReflectBinaryArray(bz, info, rv, opts)

	case reflect.Interface:
		return decodeReflectBinaryInterface(bz, info, rv, opts)

	case reflect.Slice:
		return decodeReflectBinarySlice(bz, info, rv, opts)

	case reflect.Struct:
		return decodeReflectBinaryStruct(bz, info, rv, opts)

	//----------------------------------------
	// Signed

	case reflect.Int64:
		var num int64
		if opts.Varint {
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
		num, _n, err = DecodeInt32(bz)
		if slide(bz, &bz, &n, _n) && err != nil {
			return
		}
		rv.SetInt(int64(num))
		return

	case reflect.Int16:
		num, _n, err = DecodeInt16(bz)
		if slide(bz, &bz, &n, _n) && err != nil {
			return
		}
		rv.SetInt(int64(num))
		return

	case reflect.Int8:
		num, _n, err = DecodeInt8(bz)
		if slide(bz, &bz, &n, _n) && err != nil {
			return
		}
		rv.SetInt(int64(num))
		return

	case reflect.Int:
		num, _n, err = DecodeVarint(bz)
		if slide(bz, &bz, &n, _n) && err != nil {
			return
		}
		rv.SetInt(int64(num))
		return

	//----------------------------------------
	// Unsigned

	case reflect.Uint64:
		var num uint64
		if opts.Varint {
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
		num, _n, err = DecodeUint32(bz)
		if slide(bz, &bz, &n, _n) && err != nil {
			return
		}
		rv.SetUint(uint64(num))
		return

	case reflect.Uint16:
		num, _n, err = DecodeUint16(bz)
		if slide(bz, &bz, &n, _n) && err != nil {
			return
		}
		rv.SetUint(uint64(num))
		return

	case reflect.Uint8:
		num, _n, err = DecodeUint8(bz)
		if slide(bz, &bz, &n, _n) && err != nil {
			return
		}
		rv.SetUint(uint64(num))
		return

	case reflect.Uint:
		num, _n, err = DecodeUvarint(bz)
		if slide(bz, &bz, &n, _n) && err != nil {
			return
		}
		rv.SetUint(uint64(num))
		return

	//----------------------------------------
	// Misc.

	case reflect.Bool:
		b, _n, err = DecodeBool(bz)
		if slide(bz, &bz, &n, _n) && err != nil {
			return
		}
		rv.SetBool(b)
		return

	case reflect.Float64:
		if !opts.Unsafe {
			err = fmt.Errorf("float support requires `wire:\"unsafe\"`")
			return
		}
		f, _n, err = DecodeFloat64(bz)
		if slide(bz, &bz, &n, _n) && err != nil {
			return
		}
		rv.SetBool(f)
		return

	case reflect.Float32:
		if !opts.Unsafe {
			err = fmt.Errorf("float support requires `wire:\"unsafe\"`")
			return
		}
		f, _n, err = DecodeFloat32(bz)
		if slide(bz, &bz, &n, _n) && err != nil {
			return
		}
		rv.SetBool(f)
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

func decodeReflectBinaryInterface(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) (n int, err error) {

	// Read disambiguation / prefix bytes.
	disfix, hasDisamb, prefix, hasPrefix, _n := decodeDisambPrefixBytes(bz)
	slide(bz, &bz, &n, _n)

	// Get concrete type info.
	var cinfo *TypeInfo
	if hasDisamb {
		cinfo, err = getTypeInfoFromDisfix(disfix)
	} else if hasPrefix {
		cinfo, err = getTypeInfoFromPrefix(prefix)
	} else {
		err = fmt.Errorf("Expected disambiguation or prefix bytes")
	}
	if err != nil {
		return
	}

	// Construct new concrete type.
	var crv = reflect.New(cinfo.Type).Elem()

	// Read into crv.
	err = decodeReflectBinary(bz, cinfo, crv, opts)
	return
}

func decodeReflectBinaryArray(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) (n int, err error) {
	ert := info.Type.Elem()
	length := info.Type.Len()
	_n := 0

	switch ert.Kind() {

	case reflect.Uint8: // Special case: byte array
		if len(bz) < length {
			return 0, fmt.Errorf("insufficient bytes to decode [%v]byte", length)
		}
		reflect.Copy(rv, reflect.ValueOf(bz[0:length]))
		return

	default: // General case.
		einfo := getTypeInfo(ert)
		for i := 0; i < length; i++ {
			erv := rv.Index(i)
			_n, err = decodeReflectBinary(bz, einfo, erv, opts)
			if slide(bz, &bz, &n, _n) && err != nil {
				return
			}
		}
		return
	}
}

func decodeReflectBinarySlice(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) (n int, err error) {
	ert := info.Type.Elem()
	_n := 0

	switch ert.Kind() {

	case reflect.Uint8: // Special case: byte slice
		var byteslice []byte
		byteslice, _n, err = DecodeByteSlice(r, lmt, n, err)
		if slide(bz, &bz, &n, _n) && err != nil {
			return
		}
		rv.Set(reflect.ValueOf(byteslice))
		return

	default: // General case.

		// Read length.
		var length int
		length, _n, err = DecodeVarint(bz)
		if slide(bz, &bz, &n, _n) && err != nil {
			return
		}

		// Read into a new slice.
		var srv = reflect.MakeSlice(rt, 0, length)
		einfo := getTypeInfo(ert)
		for i := 0; i < length; i++ {
			erv := srv.Index(j)
			_n, err = decodeReflectBinary(bz, einfo, erv, opts)
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

func decodeReflectBinaryStruct(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) (n int, err error) {
	_n := 0

	switch info.Type {

	case timeType: // Special case: time.Time
		t, _n, err := DecodeTime(r)
		if slide(bz, &bz, &n, _n) && err != nil {
			return
		}
		rv.Set(reflect.ValueOf(t))
		return

	default:
		for _, finfo := range typeInfo.Fields {
			frv := rv.Field(finfo.Index)
			_n, err = decodeReflectBinary(bz, finfo, frv, finfo.Options)
			if slide(bz, &bz, &n, _n) && err != nil {
				return err
			}
		}
		return
	}
}

//----------------------------------------
// encodeReflectBinary

func encodeReflectBinary(w io.Writer, info *TypeInfo, rv reflect.Value, opts Options) (err error) {

	// Dereference pointer transparently.
	if info.Type.Kind() == reflect.Ptr {
		var rt reflect.Type
		rv, rt = rv.Elem(), info.Type.Elem()
		info = getTypeInfo(rt)
	}

	switch info.Type.Kind() {

	//----------------------------------------
	// Complex

	case reflect.Array:
		err = encodeReflectBinaryArray(w, info, rv, opts)

	case reflect.Interface:
		err = encodeReflectBinaryInterface(w, info, rv, opts)

	case reflect.Slice:
		err = encodeReflectBinarySlice(w, info, rv, opts)

	case reflect.Struct:
		err = encodeReflectBinaryStruct(w, info, rv, opts)

	//----------------------------------------
	// Signed

	case reflect.Int64:
		if opts.Varint {
			err = EncodeVarint(w, int(rv.Int()))
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
		err = EncodeVarint(w, int(rv.Int()))

	//----------------------------------------
	// Unsigned

	case reflect.Uint64:
		if opts.Varint {
			err = EncodeUvarint(w, uint(rv.Uint()))
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
		err = EncodeUvarint(w, uint(rv.Uint()))

	//----------------------------------------
	// Misc

	case reflect.Bool:
		err = EncodeBool(w, rv.Bool())

	case reflect.Float64:
		if !opts.Unsafe {
			err = fmt.Errorf("Wire float* support requires `wire:\"unsafe\"`")
			return
		}
		err = EncodeFloat64(w, rv.Float())

	case reflect.Float32:
		if !opts.Unsafe {
			err = fmt.Errorf("Wire float* support requires `wire:\"unsafe\"`")
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

func encodeReflectBinaryInterface(w io.Writer, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {

	if rv.IsNil() {
		err = w.Write([]byte{0x00, 0x00, 0x00, 0x00})
		return
	}

	crv := rv.Elem()  // concrete reflection value
	crt := crv.Type() // concrete reflection type

	// Dereference pointer transparently.
	if crt.Kind() == reflect.Ptr {
		crv = crt.Elem()
		crt = crt.Elem()
		if !crv.IsValid() {
			err = fmt.Errorf("unexpected nil-pointer of type %v for registered interface %v. "+
				"For compatibility with other languages, nil-pointer interface values are forbidden.", crt, rt.Name())
			return
		}
	}

	// Get *TypeInfo for concrete type.
	cinfo := getTypeInfo(crt)
	if !cinfo.Registered {
		err = fmt.Errorf("Cannot encode unknown type %v", crt)
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

	err = encodeReflectBinary(w, cinfo, crv, opts)
	return
}

func encodeReflectBinaryArray(w io.Writer, info *TypeInfo, rv reflect.Value, opts Options) (err error) {
	ert := info.Type.Elem()
	length := rt.Len()

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
		einfo := getTypeInfo(ert)
		for i := 0; i < length; i++ {
			erv := rv.Index(i)
			err = encodeReflectBinary(w, einfo, erv, opts)
			if err != nil {
				return err
			}
		}
		return
	}
}

func encodeReflectBinarySlice(w io.Writer, info *TypeInfo, rv reflect.Value, opts Options) (err error) {
	ert := info.Type.Elem()

	switch ert.Kind() {

	case reflect.Uint8: // Special case: byte slice
		byteslice := rv.Bytes()
		_, err = EncodeByteSlice(w, byteslice)
		return

	default:
		// Write length
		length := rv.Len()
		_, err = EncodeVarint(w, length)
		if err != nil {
			return err
		}

		// Write elems
		einfo := getTypeInfo(ert)
		for i := 0; i < length; i++ {
			erv := rv.Index(i)
			_, err = encodeReflectBinary(w, einfo, erv, opts)
			if err != nil {
				return
			}
		}
		return
	}
}

func encodeReflectBinaryStruct(w io.Writer, info *TypeInfo, rv reflect.Value, opts Options) (err error) {

	switch rt.Kind() {

	case timeType: // Special case: time.Time
		_, err = EncodeTime(w, rv.Interface().(time.Time))
		return

	default:
		for _, finfo := range info.Fields {
			// fieldIdx, fieldType, opts := fieldInfo.unpack()
			frv := rv.Field(finvi.Index)
			err = encodeReflectBinary(w, finfo, frv, finfo.FieldOptions)
			if err != nil {
				return
			}
		}
		return
	}

}

//----------------------------------------
// TypeInfo

var mtx sync.RWMutex
var typeInfos = make(map[reflect.Type]*TypeInfo)
var interfaceInfos []*TypeInfo
var prefixToTypeInfos = make(map[PrefixBytes]*TypeInfo)
var disfixToTypeInfos = make(map[DisfixBytes]*TypeInfo)

func setTypeInfo(info *TypeInfo) {
	mtx.Lock()
	defer mtx.Unlock()

	typeInfos[info.Type] = info
	if info.Type.Kind() == reflect.Interface {
		interfaceInfos = append(interfaceInfos, info)
	} else if info.Registered {
		prefix := info.Prefix
		disamb := info.Disamb
		disfix := toDisfix(prefix, disamb)
		prefixToTypeInfos[prefix] = info
		disfixToTypeInfos[disfix] = info
	}
}

func getTypeInfo(rt reflect.Type) (info *TypeInfo, err error) {
	mtx.RLock()
	defer mtx.RUnlock()

	info, ok := typeInfos[rt]
	if !ok {
		err = fmt.Errorf("unregistered interface type %v", rt)
	}
	return
}

func getTypeInfoFromPrefix(pb PrefixBytes) (info *TypeInfo, err error) {
	mtx.RLock()
	defer mtx.RUnlock()

	info, ok := prefixToTypeInfos[pb]
	if !ok {
		err = fmt.Errorf("unrecognized prefix bytes %X", pb)
	}
	return
}

func getTypeInfoFromDisfix(df DisfixBytes) (info *TypeInfo, err error) {
	mtx.RLock()
	defer mtx.RUnlock()

	info, ok := disfixToTypeInfos[df]
	if !ok {
		err = fmt.Errorf("unrecognized disambiguation+prefix bytes %X", df)
	}
	return
}

func parseFieldInfos(rt reflect.Type) (infos []FieldInfo) {
	if rt.Kind() != reflect.Struct {
		return nil
	}

	infos = make([]FieldInfo, 0, rt.NumField())
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if field.PkgPath != "" {
			continue // field is private
		}
		skip, opts := parseFieldOptions(field)
		if skip {
			continue // e.g. json:"-"
		}
		fieldInfo := FieldInfo{
			Index:        i,
			Type:         field.Type,
			ZeroValue:    reflect.Zero(field.Type).Interface(),
			FieldOptions: opts,
		}
		checkUnsafe(fieldInfo)
		infos = append(infos, fieldInfo)
	}
	return infos
}

func parseFieldOptions(field reflect.StructField) (skip bool, opts FieldOptions) {
	binTag := field.Tag.Get("binary")
	wireTag := field.Tag.Get("wire")
	jsonTag := field.Tag.Get("json")

	// If `json:"-"`, don't encode.
	// NOTE: This skips binary as well.
	if jsonTag == "-" {
		skip = true
		return
	}

	// Get JSON field name.
	jsonTagParts := strings.Split(jsonTag, ",")
	if jsonTagParts[0] == "" {
		opts.JSONName = field.Name
	} else {
		opts.JSONName = jsonTagParts[0]
	}

	// Get JSON omitempty.
	if len(jsonTagParts) > 1 {
		if jsonTagParts[1] == "omitempty" {
			opts.JSONOmitEmpty = true
		}
	}

	// Parse binary tags.
	if binTag == "varint" { // TODO: extend
		opts.Varint = true
	}

	// Parse wire tags.
	if wireTag == "unsafe" {
		opts.Unsafe = true
	}

	return
}

//----------------------------------------
// Misc.

func getTypeFromPointer(ptr interface{}) reflect.Type {
	rt := reflect.TypeOf(ptr)
	if rt.Kind() == reflect.Ptr {
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
	bz := hasher.Sum()
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
		err = fmt.Errorf("eof while reading prefix bytes")
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
			err = fmt.Errorf("eof while reading disamb bytes")
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
