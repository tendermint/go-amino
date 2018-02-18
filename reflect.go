package wire

// XXX Add JSON again.
// XXX Check for custom marshal/unmarshal functions.
// XXX Scan the codebase for unwraps and double check that they implement above.

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/davecgh/go-spew/spew"
)

//----------------------------------------
// constants

var timeType = reflect.TypeOf(time.Time{})

const RFC3339Millis = "2006-01-02T15:04:05.000Z" // forced microseconds
const printLog = false

//----------------------------------------
// cdc.decodeReflectBinary

// This is the main entrypoint for decoding all types from binary form.  This
// function calls decodeReflectBinary*, and generally those functions should
// only call this one, for the prefix bytes are consumed here when present.
// CONTRACT: rv.CanAddr() is true.
func (cdc *Codec) decodeReflectBinary(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) (n int, err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}
	if info.Type.Kind() == reflect.Interface && rv.Kind() == reflect.Ptr {
		panic("should not happen")
	}

	if printLog {
		spew.Printf("(d) decodeReflectBinary(bz: %X, info: %v, rv: %#v (%v), opts: %v)\n",
			bz, info, rv.Interface(), rv.Type(), opts)
		defer func() {
			fmt.Printf("(d) -> n: %v, err: %v\n", n, err)
		}()
	}

	// TODO Read the disamb bytes here if necessary.
	// e.g. rv isn't an interface, and
	// info.ConcreteType.AlwaysDisambiguate.  But we don't support
	// this yet.

	if !info.Registered {
		// No need for disambiguation, decode as is.
		n, err = cdc._decodeReflectBinary(bz, info, rv, opts)
		return
	}

	// It's a registered concrete type.
	// Implies that info holds the info we need.
	// Just strip the prefix bytes after checking it.
	if len(bz) < PrefixBytesLen {
		err = errors.New("EOF skipping prefix bytes.")
		return
	}
	if !info.Prefix.EqualBytes(bz) {
		panic("should not happen")
	}
	bz = bz[PrefixBytesLen:]
	n += PrefixBytesLen

	_n := 0
	_n, err = cdc._decodeReflectBinary(bz, info, rv, opts)
	slide(&bz, &n, _n)
	return
}

// CONTRACT: any immediate disamb/prefix bytes have been consumed/stripped.
// CONTRACT: rv.CanAddr() is true.
func (cdc *Codec) _decodeReflectBinary(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) (n int, err error) {

	// TODO consider the binary equivalent of json.Unmarshaller.

	// If a pointer, handle pointer byte.
	// 0x00 means nil, 0x01 means not nil.
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
	}

	// Dereference-and-construct pointers all the way.
	// This works for pointer-pointers.
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			newPtr := reflect.New(rv.Type().Elem())
			rv.Set(newPtr)
		}
		rv = rv.Elem()
	}

	var _n int
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
			if slide(&bz, &n, _n) && err != nil {
				return
			}
			rv.SetInt(num)
		} else {
			num, _n, err = DecodeInt64(bz)
			if slide(&bz, &n, _n) && err != nil {
				return
			}
			rv.SetInt(num)
		}
		return

	case reflect.Int32:
		var num int32
		num, _n, err = DecodeInt32(bz)
		if slide(&bz, &n, _n) && err != nil {
			return
		}
		rv.SetInt(int64(num))
		return

	case reflect.Int16:
		var num int16
		num, _n, err = DecodeInt16(bz)
		if slide(&bz, &n, _n) && err != nil {
			return
		}
		rv.SetInt(int64(num))
		return

	case reflect.Int8:
		var num int8
		num, _n, err = DecodeInt8(bz)
		if slide(&bz, &n, _n) && err != nil {
			return
		}
		rv.SetInt(int64(num))
		return

	case reflect.Int:
		var num int64
		num, _n, err = DecodeVarint(bz)
		if slide(&bz, &n, _n) && err != nil {
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
			if slide(&bz, &n, _n) && err != nil {
				return
			}
			rv.SetUint(num)
		} else {
			num, _n, err = DecodeUint64(bz)
			if slide(&bz, &n, _n) && err != nil {
				return
			}
			rv.SetUint(num)
		}
		return

	case reflect.Uint32:
		var num uint32
		num, _n, err = DecodeUint32(bz)
		if slide(&bz, &n, _n) && err != nil {
			return
		}
		rv.SetUint(uint64(num))
		return

	case reflect.Uint16:
		var num uint16
		num, _n, err = DecodeUint16(bz)
		if slide(&bz, &n, _n) && err != nil {
			return
		}
		rv.SetUint(uint64(num))
		return

	case reflect.Uint8:
		var num uint8
		num, _n, err = DecodeUint8(bz)
		if slide(&bz, &n, _n) && err != nil {
			return
		}
		rv.SetUint(uint64(num))
		return

	case reflect.Uint:
		var num uint64
		num, _n, err = DecodeUvarint(bz)
		if slide(&bz, &n, _n) && err != nil {
			return
		}
		rv.SetUint(num)
		return

	//----------------------------------------
	// Misc.

	case reflect.Bool:
		var b bool
		b, _n, err = DecodeBool(bz)
		if slide(&bz, &n, _n) && err != nil {
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
		if slide(&bz, &n, _n) && err != nil {
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
		if slide(&bz, &n, _n) && err != nil {
			return
		}
		rv.SetFloat(float64(f))
		return

	case reflect.String:
		var str string
		str, _n, err = DecodeString(bz)
		if slide(&bz, &n, _n) && err != nil {
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
		// JAE: Heed this note, this is very tricky.
		err = errors.New("Decoding to a non-nil interface is not supported yet")
		return
	}

	// Peek disambiguation / prefix info.
	disfix, hasDisamb, prefix, hasPrefix, isNil, _, err := decodeDisambPrefixBytes(bz)
	if err != nil {
		return
	}

	// Special case for nil.
	if isNil {
		n += 1 + DisambBytesLen // Consume 0x{00 00 00 00}
		rv.Set(iinfo.ZeroValue)
		return
	}

	// Consume disamb (if any) and prefix bytes.
	if hasDisamb {
		slide(&bz, &n, 1+DisfixBytesLen)
	} else {
		slide(&bz, &n, PrefixBytesLen)
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

	// Construct the concrete type.
	var crv, irvSet = constructConcreteType(cinfo)

	// Decode into the concrete type.
	_n := 0
	_n, err = cdc._decodeReflectBinary(bz, cinfo, crv, opts)
	if slide(&bz, &n, _n) && err != nil {
		rv.Set(irvSet) // Helps with debugging
		return
	}

	// We need to set here, for when !PointerPreferred and the type
	// is say, an array of bytes (e.g. [32]byte), then we must call
	// rv.Set() *after* the value was acquired.
	// NOTE: rv.Set() should succeed because it was validated
	// already during Register[Interface/Concrete].
	rv.Set(irvSet)
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
			if slide(&bz, &n, _n) && err != nil {
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
	_n := 0 // nolint: ineffassign

	switch ert.Kind() {

	case reflect.Uint8: // Special case: byte slice
		var byteslice []byte
		byteslice, _n, err = DecodeByteSlice(bz)
		if slide(&bz, &n, _n) && err != nil {
			return
		}
		if len(byteslice) == 0 {
			// Special case when length is 0.
			// NOTE: We prefer nil slices.
			rv.Set(info.ZeroValue)
		} else {
			rv.Set(reflect.ValueOf(byteslice))
		}
		return

	default: // General case.
		var einfo *TypeInfo
		einfo, err = cdc.getTypeInfo_wlock(ert)
		if err != nil {
			return
		}

		// Read length.
		var length = int(0)
		length64 := int64(0)
		length64, _n, err = DecodeVarint(bz)
		if slide(&bz, &n, _n) && err != nil {
			return
		}
		length = int(length64)
		if length < 0 {
			err = errors.New("Invalid negative slice length")
			return
		}

		// Special case when length is 0.
		// NOTE: We prefer nil slices.
		if length == 0 {
			rv.Set(info.ZeroValue)
			return
		}

		// Read into a new slice.
		var esrt = reflect.SliceOf(ert) // TODO could be optimized.
		var srv = reflect.MakeSlice(esrt, length, length)
		for i := 0; i < length; i++ {
			erv := srv.Index(i)
			_n, err = cdc.decodeReflectBinary(bz, einfo, erv, opts)
			if slide(&bz, &n, _n) && err != nil {
				return
			}
		}

		// TODO do we need this extra step?
		rv.Set(srv)
		return
	}
}

// CONTRACT: rv.CanAddr() is true.
func (cdc *Codec) decodeReflectBinaryStruct(bz []byte, info *TypeInfo, rv reflect.Value, _ FieldOptions) (n int, err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}
	_n := 0 // nolint: ineffassign

	switch info.Type {

	case timeType: // Special case: time.Time
		var t time.Time
		t, _n, err = DecodeTime(bz)
		if slide(&bz, &n, _n) && err != nil {
			return
		}
		rv.Set(reflect.ValueOf(t))
		return

	default:
		for _, field := range info.Fields {

			// Get field rv and info.
			var frv = rv.Field(field.Index)
			var finfo *TypeInfo
			finfo, err = cdc.getTypeInfo_wlock(field.Type)
			if err != nil {
				return
			}

			// Decode into field rv.
			_n, err = cdc.decodeReflectBinary(bz, finfo, frv, field.FieldOptions)
			if slide(&bz, &n, _n) && err != nil {
				return
			}
		}
		return
	}
}

//----------------------------------------
// cdc.encodeReflectBinary

// This is the main entrypoint for encoding all types in binary form.  This
// function calls encodeReflectBinary*, and generally those functions should
// only call this one, for the prefix bytes are only written here.
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

	//----------------------------------------
	// Default

	default:
		panic(fmt.Sprintf("unsupported type %v", info.Type.Kind()))
	}

	return
}

func (cdc *Codec) encodeReflectBinaryInterface(w io.Writer, iinfo *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {

	if rv.IsNil() {
		_, err = w.Write([]byte{0x00, 0x00, 0x00, 0x00})
		return
	}

	// Get concrete non-pointer reflect value & type.
	var crv = rv.Elem()
	crv, err = derefForInterface(crv, iinfo)
	if err != nil {
		return
	}
	var crt = crv.Type()

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
	if iinfo.AlwaysDisambiguate {
		needDisamb = true
	} else if len(iinfo.Implementers[cinfo.Prefix]) > 1 {
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
		bz := []byte(nil)
		if rv.CanAddr() {
			bz = rv.Slice(0, length).Bytes()
		} else {
			bz = make([]byte, length)
			reflect.Copy(reflect.ValueOf(bz), rv) // XXX: looks expensive!
		}
		_, err = w.Write(bz)
		return

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
				return
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
			// Get field value and info.
			var frv = rv.Field(field.Index)
			var finfo *TypeInfo
			finfo, err = cdc.getTypeInfo_wlock(field.Type)
			if err != nil {
				return
			}
			// Write field value.
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

func NameToDisfix(name string) (db DisambBytes, pb PrefixBytes) {
	return nameToDisfix(name)
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
		if (DisambBytes{}).EqualBytes(bz[1:4]) {
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
func slide(bz *[]byte, n *int, _n int) bool {
	if _n < 0 || _n > len(*bz) {
		panic(fmt.Sprintf("impossible slide: len:%v _n:%v", len(*bz), _n))
	}
	*bz = (*bz)[_n:]
	*n += _n
	return true
}

// Dereference pointer transparently for interface iinfo.
// This also works for pointer-pointers.
func derefForInterface(crv reflect.Value, iinfo *TypeInfo) (reflect.Value, error) {
	if iinfo.Type.Kind() != reflect.Interface {
		panic("derefForInterface() expects interface type info")
	}
	// NOTE: Encoding pointer-pointers only work for no-method interfaces like
	// `interface{}`.
	for crv.Kind() == reflect.Ptr {
		crv = crv.Elem()
		if crv.Kind() == reflect.Interface {
			err := fmt.Errorf("Unexpected interface-pointer of type *%v for registered interface %v. Not supported yet.", crv.Type(), iinfo.Type)
			return crv, err
		}
		if !crv.IsValid() {
			err := fmt.Errorf("Illegal nil-pointer of type %v for registered interface %v. "+
				"For compatibility with other languages, nil-pointer interface values are forbidden.", crv.Type(), iinfo.Type)
			return crv, err
		}
	}
	return crv, nil
}

// constructConcreteType creates the concrete value as
// well as the corresponding settable value for it.
// Return irvSet which should be set on caller's interface rv.
func constructConcreteType(cinfo *TypeInfo) (crv, irvSet reflect.Value) {
	// Construct new concrete type.
	if cinfo.PointerPreferred {
		cPtrRv := reflect.New(cinfo.Type)
		crv = cPtrRv.Elem()
		irvSet = cPtrRv
	} else {
		crv = reflect.New(cinfo.Type).Elem()
		irvSet = crv
	}
	return
}
