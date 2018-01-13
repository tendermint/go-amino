package wire

// XXX Add JSON again.
// XXX Check for custom marshal/unmarshal functions.
// XXX Scan the codebase for unwraps and double check that they implement above.

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"io"
	"reflect"
	"time"
)

//----------------------------------------
// constants

var timeType = reflect.TypeOf(time.Time{})

const RFC3339Millis = "2006-01-02T15:04:05.000Z" // forced microseconds
const printLog = false

//----------------------------------------
// cdc.decodeReflectBinary

// CONTRACT: rv.CanAddr() is true.
func (cdc *Codec) decodeReflectBinary(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) (n int, err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}

	if printLog {
		spew.Printf("(d) decodeReflectBinary(bz: %X, info: %v, rv: %#v (%v), opts: %v)\n",
			bz, info, rv.Interface(), rv.Type(), opts)
		defer func() {
			fmt.Printf("(d) -> n: %v, err: %v\n", n, err)
		}()
	}

	var _n int

	if info.Registered {
		if len(bz) < PrefixBytesLen {
			err = errors.New("EOF skipping prefix bytes.")
			return
		}
		bz = bz[PrefixBytesLen:]
		n += PrefixBytesLen
	}

	// SANITY CHECK
	if info.Type.Kind() == reflect.Interface && rv.Kind() == reflect.Ptr {
		panic("should not happen")
	}

	// Handle pointer types.
	if rv.Kind() == reflect.Ptr {
		if len(bz) == 0 {
			err = errors.New("EOF reading pointer type")
			return
		}
		switch bz[0] {
		case 0x00:
			n += 1
			rv.Set(reflect.Zero(rv.Type()))
			return
		case 0x01:
			n += 1
			bz = bz[1:]
			// so continue...
		default:
			err = fmt.Errorf("unexpected pointer byte %X", bz[0])
			return
		}

		// Dereference-and-construct pointers all the way.
		// This works for pointer-pointers.
		for c := true; c; c = rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				newPtr := reflect.New(rv.Type().Elem())
				rv.Set(newPtr)
			}
			rv = rv.Elem()
		}
	}

	switch info.Type.Kind() {

	//----------------------------------------
	// Complex

	case reflect.Interface:
		_n, err = cdc.decodeReflectBinaryInterface(bz, info, rv, opts)
		n += _n
		return

	case reflect.Array:
		_n, err = cdc.decodeReflectBinaryArray(bz, info, rv, opts)
		n += _n
		return

	case reflect.Slice:
		_n, err = cdc.decodeReflectBinarySlice(bz, info, rv, opts)
		n += _n
		return

	case reflect.Struct:
		_n, err = cdc.decodeReflectBinaryStruct(bz, info, rv, opts)
		n += _n
		return

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
			err = errors.New("Float support requires `wire:\"unsafe\"`.")
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
			err = errors.New("Float support requires `wire:\"unsafe\"`.")
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
func (cdc *Codec) decodeReflectBinaryInterface(bz []byte, iinfo *TypeInfo, rv reflect.Value, opts FieldOptions) (n int, err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}
	if !rv.IsNil() {
		// This is very tricky.
		err = errors.New("Decoding to a non-nil interface is not supported yet")
		return
	}

	// Read disambiguation / prefix bytes but do not consume the prefix bytes.
	disfix, hasDisamb, prefix, hasPrefix, isNil, _n, err := decodeDisambPrefixBytes(bz)
	if hasDisamb {
		n += DisfixBytesLen
	}
	if err != nil {
		return
	}

	// Special case for nil
	if isNil {
		rv.Set(iinfo.ZeroValue)
		return
	}

	// Get concrete type info.
	var cinfo *TypeInfo
	if hasDisamb {
		cinfo, err = cdc.getTypeInfoFromDisfix_rlock(disfix)
	} else if hasPrefix {
		cinfo, err = cdc.getTypeInfoFromPrefix_rlock(iinfo, prefix)
	} else {
		err = errors.New("Expected disambiguation or prefix bytes.")
	}
	if err != nil {
		return
	}

	// Construct new concrete type.
	// NOTE: rv.Set() should succeed because it was validated
	// already during Register[Interface/Concrete].
	var crv reflect.Value
	if cinfo.PointerPreferred {
		cPtrRv := reflect.New(cinfo.Type)
		crv = cPtrRv.Elem()
		rv.Set(cPtrRv)
	} else {
		crv = reflect.New(cinfo.Type).Elem()
		rv.Set(crv)
	}

	// Read into crv.
	_n, err = cdc.decodeReflectBinary(bz, cinfo, crv, opts)
	slide(bz, &bz, &n, _n)
	return
}

// CONTRACT: rv.CanAddr() is true.
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
		n += length
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
		if len(byteslice) == 0 {
			rv.Set(reflect.ValueOf([]byte(nil)))
		} else {
			rv.Set(reflect.ValueOf(byteslice))
		}
		return

	default: // General case.

		// Read length.
		var length int64
		length, _n, err = DecodeVarint(bz)
		if slide(bz, &bz, &n, _n) && err != nil {
			return
		}

		// Special case when length is 0.
		if length == 0 {
			rv.Set(info.ZeroValue)
			return
		}

		// Read into a new slice.
		var esrt = reflect.SliceOf(ert) // TODO could be optimized.
		var srv = reflect.MakeSlice(esrt, int(length), int(length))
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
		}

		// TODO do we need this extra step?
		rv.Set(srv)
		return
	}
}

// CONTRACT: rv.CanAddr() is true.
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

// This is the main entrypoint for encoding all types.  This function calls
// encodeReflectBinary*, but those functions should only call this one.
// (This is necessary because the prefix bytes are only written here).
func (cdc *Codec) encodeReflectBinary(w io.Writer, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {

	if printLog {
		spew.Printf("(e) encodeReflectBinary(info: %v, rv: %#v (%v), opts: %v)\n",
			info, rv.Interface(), rv.Type(), opts)
		defer func() {
			fmt.Printf("(e) -> err: %v\n", err)
		}()
	}

	// Write the prefix bytes if it is a registered concrete type.
	if info.Registered {
		_, err = w.Write(info.Prefix[:])
		if err != nil {
			return
		}
	}

	// Dereference pointers all the way if any.
	// This works for pointer-pointers.
	var foundPointer = false
	for rv.Kind() == reflect.Ptr {
		foundPointer = true
		rv = rv.Elem()
	}

	// Write pointer byte if necessary.
	if foundPointer {
		if rv.IsValid() {
			_, err = w.Write([]byte{0x01})
			// and continue...
		} else {
			_, err = w.Write([]byte{0x00})
			return
		}
	}

	// Sanity check
	if info.Registered && foundPointer {
		panic("should not happen")
	}

	switch info.Type.Kind() {

	//----------------------------------------
	// Complex

	case reflect.Interface:
		err = cdc.encodeReflectBinaryInterface(w, info, rv, opts)

	case reflect.Array:
		err = cdc.encodeReflectBinaryArray(w, info, rv, opts)

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
			err = errors.New("Wire float* support requires `wire:\"unsafe\"`.")
			return
		}
		err = EncodeFloat64(w, rv.Float())

	case reflect.Float32:
		if !opts.Unsafe {
			err = errors.New("Wire float* support requires `wire:\"unsafe\"`.")
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
	var needDisamb bool = false
	if info.AlwaysDisambiguate {
		needDisamb = true
	} else if len(info.Implementers[cinfo.Prefix]) > 1 {
		needDisamb = true
	}
	if needDisamb {
		_, err = w.Write(append([]byte{0x00}, cinfo.Disamb[:]...))
		if err != nil {
			return
		}
	}

	err = cdc.encodeReflectBinary(w, cinfo, crv, opts)
	return
}

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

func nameToDisamb(name string) (db DisambBytes) {
	db, _ = nameToDisfix(name)
	return
}

func nameToPrefix(name string) (pb PrefixBytes) {
	_, pb = nameToDisfix(name)
	return
}

func nameToDisfix(name string) (db DisambBytes, pb PrefixBytes) {
	hasher := sha256.New()
	hasher.Write([]byte(name))
	bz := hasher.Sum(nil)
	for bz[0] == 0x00 {
		bz = bz[1:]
	}
	copy(db[:], bz[0:3])
	bz = bz[3:]
	for bz[0] == 0x00 {
		bz = bz[1:]
	}
	copy(pb[:], bz[0:4])
	return
}

func toDisfix(db DisambBytes, pb PrefixBytes) (df DisfixBytes) {
	copy(df[0:3], db[0:3])
	copy(df[3:7], pb[0:4])
	return
}

func decodeDisambPrefixBytes(bz []byte) (df DisfixBytes, hasDb bool, pb PrefixBytes, hasPb bool, isNil bool, n int, err error) {
	// Validate
	if len(bz) < 4 {
		err = errors.New("EOF reading prefix bytes.")
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
			err = errors.New("EOF reading disamb bytes.")
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
