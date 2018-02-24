package wire

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/davecgh/go-spew/spew"
)

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

	// Read prefix bytes and typ3 byte if registered.
	if info.Registered {
		// Strip the prefix bytes after checking it.
		if len(bz) < PrefixBytesLen {
			err = errors.New("EOF skipping prefix bytes.")
			return
		}
		// Check and consume prefix bytes.
		if !info.Prefix.EqualBytes(bz) {
			panic("should not happen")
		}
		bz = bz[PrefixBytesLen:]
		n += PrefixBytesLen
		// Check and consume typ3 byte.
		err = decodeTyp3AndCheck(info.Type, &bz, opts)
		if err != nil {
			return
		}
	}

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
		ert := info.Type.Elem()
		if ert == reflect.Uint8 {
			_n, err = cdc.decodeReflectBinaryByteArray(bz, info, rv, opts)
			n += _n
		} else {
			_n, err = cdc.decodeReflectBinaryArray(bz, info, rv, opts)
			n += _n
		}
		return

	case reflect.Slice:
		ert := info.Type.Elem()
		if ert == reflect.Uint8 {
			_n, err = cdc.decodeReflectBinaryByteSlice(bz, info, rv, opts)
			n += _n
		} else {
			_n, err = cdc.decodeReflectBinarySlice(bz, info, rv, opts)
			n += _n
		}
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

	// Get concrete type info from disfix/prefix.
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

	// Check and consume typ3 byte.
	// It cannot be a typ4 byte because it cannot be nil.
	err = decodeTyp3AndCheck(cinfo.Type, &bz, opts)
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
func (cdc *Codec) decodeReflectBinaryByteArray(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) (n int, err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}
	ert := info.Type.Elem()
	if ert.Kind() != reflect.Uint8 {
		panic("should not happen")
	}
	length := info.Type.Len()
	if len(bz) < length {
		return 0, fmt.Errorf("Insufficient bytes to decode [%v]byte.", length)
	}

	// Read byte-length prefixed byteslice.
	var byteslice, _n = []byte(nil), int64(0)
	byteslice, _n, err = DecodeByteSlice(bz)
	if slide(&bz, &n, _n) && err != nil {
		return
	}
	if len(byteslice) != length {
		err = errors.New("Mismatched byte array length: Expected %v, got %v",
			length, len(byteslice))
		return
	}

	// Copy read byteslice to rv array.
	reflect.Copy(rv, reflect.ValueOf(byteslice))
	return
}

// CONTRACT: rv.CanAddr() is true.
func (cdc *Codec) decodeReflectBinaryArray(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) (n int, err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}
	ert := info.Type.Elem()
	if ert.Kind() == reflect.Uint8 {
		panic("should not happen")
	}
	length := info.Type.Len()
	einfo := *TypeInfo(nil)
	einfo, err = cdc.getTypeInfo_wlock(ert)
	if err != nil {
		return
	}

	// Check and consume typ4 byte.
	var ptr, err = decodeTyp4AndCheck(ert, &bz, opts)
	if err != nil {
		return
	}

	// Read number of items.
	var count, _n, err = DecodeVarint(bz)
	if slide(&bz, &n, _n) && err != nil {
		return
	}
	if count != length {
		err = errors.New("Expected num items of %v, decoded %v", length, count)
		return
	}

	// NOTE: Unlike decodeReflectBinarySlice,
	// there is nothing special to do for
	// zero-length arrays.  Is that even possible?

	// Read each item.
	for i := 0; i < length; i++ {
		var erv, _n = rv.Index(i), int(0)
		// Maybe read nil.
		if ptr {
			numNil := int64(0)
			numNil, _n, err = decodeNumNilBytes(bz)
			if slide(&bz, &n, _n) && err != nil {
				return
			}
			if numNil == 0 {
				// Good, continue decoding item.
			} else if numNil == 1 {
				// Set nil/zero.
				erv.Set(reflect.Zero(erv.Type()))
				continue
			} else {
				panic("should not happen")
			}
		}
		// Decode non-nil value.
		_n, err = cdc.decodeReflectBinary(bz, einfo, erv, opts)
		if slide(&bz, &n, _n) && err != nil {
			return
		}
	}
	return
}

// CONTRACT: rv.CanAddr() is true.
func (cdc *Codec) decodeReflectBinaryByteSlice(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) (n int, err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}
	ert := info.Type.Elem()
	if ert.Kind() != reflect.Uint8 {
		panic("should not happen")
	}

	// Read byte-length prefixed byteslice.
	var byteslice, _n = []byte(nil), int64(0)
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
}

// CONTRACT: rv.CanAddr() is true.
func (cdc *Codec) decodeReflectBinarySlice(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) (n int, err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}
	ert := info.Type.Elem()
	if ert.Kind() == reflect.Uint8 {
		panic("should not happen")
	}
	einfo := *TypeInfo(nil)
	einfo, err = cdc.getTypeInfo_wlock(ert)
	if err != nil {
		return
	}

	// Check and consume typ4 byte.
	var ptr, err = decodeTyp4AndCheck(ert, &bz, opts)
	if err != nil {
		return
	}

	// Read number of items.
	var count, _n, err = DecodeVarint(bz)
	if slide(&bz, &n, _n) && err != nil {
		return
	}
	if count < 0 {
		err = errors.New("Invalid negative slice length")
		return
	}

	// Special case when length is 0.
	// NOTE: We prefer nil slices.
	if count == 0 {
		rv.Set(info.ZeroValue)
		return
	}

	// Read each item.
	// NOTE: Unlike decodeReflectBinaryArray,
	// we need to construct a new slice before
	// we populate it. Arrays on the other hand
	// reserve space in the value itself.
	var esrt = reflect.SliceOf(ert) // TODO could be optimized.
	var srv = reflect.MakeSlice(esrt, int(count), int(count))
	for i := 0; i < length; i++ {
		var erv, _n = rv.Index(i), int(0)
		// Maybe read nil.
		if ptr {
			numNil := int64(0)
			numNil, _n, err = decodeNumNilBytes(bz)
			if slide(&bz, &n, _n) && err != nil {
				return
			}
			if numNil == 0 {
				// Good, continue decoding item.
			} else if numNil == 1 {
				// Set nil/zero.
				erv.Set(reflect.Zero(erv.Type()))
				continue
			} else {
				panic("should not happen")
			}
		}
		// Decode non-nil value.
		_n, err = cdc.decodeReflectBinary(bz, einfo, erv, opts)
		if slide(&bz, &n, _n) && err != nil {
			return
		}
	}
	rv.Set(srv)
	return
}

// CONTRACT: rv.CanAddr() is true.
func (cdc *Codec) decodeReflectBinaryStruct(bz []byte, info *TypeInfo, rv reflect.Value, _ FieldOptions) (n int, err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}
	_n := 0 // nolint: ineffassign

	// The "Start struct" type3 doesn't get read here.
	// It's already implied, either by struct-key or list-element-type-byte.

	switch info.Type {

	case timeType:
		// Special case: time.Time
		var t time.Time
		t, _n, err = DecodeTime(bz)
		if slide(&bz, &n, _n) && err != nil {
			return
		}
		rv.Set(reflect.ValueOf(t))
		return

	default:
		// Read each field.
		for _, field := range info.Fields {

			// Read field key (number and type).
			var fieldNum, typ = int64(0), typ3(0x00)
			fieldNum, typ, _n, err = decodeFieldNumberAndTyp3(bz)
			if field.BinFieldNum < fieldNum {
				// Set nil field value.
				rv.Set(reflect.Zero(rv.Type()))
				continue
				// Do not slide, we will read it again.
			}
			if slide(&bz, &n, _n) && err != nil {
				return
			}
			// NOTE: In the future, we'll support upgradeability.
			// So in the future, this may not match,
			// so we will need to remove this sanity check.
			if field.BinFieldNum != fieldNum {
				err = errors.New(fmt.Sprintf("Expected field number %v, got %v", field.BinFieldNum, fieldNum))
				return
			}
			typ3Wanted := typeToTyp3(field.Type, field.FieldOptions)
			if typ != typ3Wanted {
				err = errors.New(fmt.Sprintf("Expected field type %X, got %X", typ3Wanted, typ))
				return
			}

			// Get field rv and info.
			var frv = rv.Field(field.Index)
			var finfo *TypeInfo
			finfo, err = cdc.getTypeInfo_wlock(field.Type)
			if err != nil {
				return
			}

			// Decode field into rv.
			_n, err = cdc.decodeReflectBinary(bz, finfo, frv, field.FieldOptions)
			if slide(&bz, &n, _n) && err != nil {
				return
			}
		}

		// Read "End struct".
		// NOTE: In the future, we'll need to break out of a loop
		// when encoutering an EndStruct typ3 byte.
		typ, _n, err = DecodeByte(bz)
		if slide(&bz, &n, _n) && err != nil {
			return
		}
		if typ != typ3_EndStruct {
			err = errors.New(fmt.Sprintf("Expected End struct typ3 byte, got %X", typ))
			return
		}
		return
	}
}

//----------------------------------------

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

// Read field key.
func decodeFieldNumberAndTyp3(bz []byte) (num int32, typ typ3, n int, err error) {

	// Read uvarint value.
	var value int64
	uvalue := uint64(0)
	uvalue, n, err = DecodeUvarint(bz)
	if err != nil {
		return
	}
	value = int64(uvalue)

	// Decode first typ3 byte.
	typ = uint8(value & 0x07)

	// Decode num.
	num64 = value >> 3
	if num64 < 0 || num64 > (1<<29-1) {
		err = errors.New(fmt.Sprintf("invalid field num %v", num64))
		return
	}
	num = int32(num64)
	return
}

// Check and consume typ4 byte and error if it doesn't match rt.
func decodeTyp4AndCheck(rt reflect.Type, bzPtr *[]byte, opts FieldOptions) (ptr bool, err error) {
	var bz = *bzPtr
	var typ, _n = typ4(0x00), int64(0)
	typ, _n, err = decodeTyp4(bz)
	if slide(&bz, &n, _n) && err != nil {
		return
	}
	typWanted := typeToTyp4(rt, opts)
	if typWanted != typ {
		err = errors.New(fmt.Sprintf("Typ4 mismatch.  Expected %X, got %X", typWanted, typ))
		return
	}
	ptr = bool(typ & 0x80)
	return
}

// Read typ4 byte.
func decodeTyp4(bz []byte) (typ typ4, n int, err error) {
	if len(bz) == 0 {
		err = errors.New(fmt.Sprintf("EOF reading typ4 bytes"))
		return
	}
	if bz[0]&0xF0 != 0 {
		err = errors.New(fmt.Sprintf("Invalid non-zero nibble reading typ4 bytes"))
		return
	}
	typ = bz[0]
	n = 1
	return
}

// Check and consume typ3 byte and error if it doesn't match rt.
func decodeTyp3AndCheck(rt reflect.Type, bzPtr *[]byte, opts FieldOptions) (err error) {
	var bz = *bzPtr
	var typ, _n = typ4(0x00), int64(0)
	typ, _n, err = decodeTyp3(bz)
	if slide(&bz, &n, _n) && err != nil {
		return
	}
	typWanted := typ3(typeToTyp4(rt, opts) & 0x07) // TODO typeToTyp3?
	if typWanted != typ {
		err = fmt.Errorf("Typ3 mismatch.  Expected %X, got %X", typWanted, typ)
		return
	}
	return
}

// Read typ3 byte.
func decodeTyp3(bz []byte) (typ typ3, n int, err error) {
	if len(bz) == 0 {
		err = errors.New(fmt.Sprintf("EOF reading typ3 bytes"))
		return
	}
	if bz[0]&0xF8 != 0 {
		err = errors.New(fmt.Sprintf("Invalid non-zero nibble reading typ3 bytes"))
		return
	}
	typ = bz[0]
	n = 1
	return
}

// Read a uvarint that encodes the number of nil items to skip.  NOTE:
// Currently does not support any number besides 0 (not nil) and 1 (nil).  All
// other values will error.
func decodeNumNilBytes(bz []byte) (numNil int64, n int, err error) {
	if len(bz[0]) == 0 {
		err = errors.New("EOF reading nil byte(s)")
		return
	}
	if bz[0] == 0x00 {
		numNil, n = 0, 1
		return
	}
	if bz[0] == 0x01 {
		numNil, n = 1, 1
		return
	}
	n, err = 0, fmt.Errorf("Unexpected nil byte %X (sparse lists not supported)", bz[0])
	return
}
