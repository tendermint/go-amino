package wire

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/davecgh/go-spew/spew"
)

//----------------------------------------
// cdc.encodeReflectBinary

// This is the main entrypoint for encoding all types in binary form.  This
// function calls encodeReflectBinary*, and generally those functions should
// only call this one, for the prefix bytes are only written here.
// CONTRACT: rv is not a pointer (nil is already handled).
// CONTRACT: rv is valid.
func (cdc *Codec) encodeReflectBinary(w io.Writer, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {
	if rv.Kind() == reflect.Ptr {
		panic("should not happen")
	}
	if !rv.IsValid() {
		panic("should not happen")
	}

	if printLog {
		spew.Printf("(e) encodeReflectBinary(info: %v, rv: %#v (%v), opts: %v)\n",
			info, rv.Interface(), rv.Type(), opts)
		defer func() {
			fmt.Printf("(e) -> err: %v\n", err)
		}()
	}

	// Maybe write prefix+typ3 bytes.
	if info.Registered {
		var typ = typeToTyp4(info.Type, opts).Typ3()
		_, err = w.Write(info.Prefix.WithTyp3(typ).Bytes())
		if err != nil {
			return
		}
	}

	err = cdc._encodeReflectBinary(w, info, rv, opts)
	return
}

// CONTRACT: rv is not a pointer (nil is already handled).
// CONTRACT: rv is valid.
// CONTRACT: any disamb/prefix+typ3 bytes have already been written.
func (cdc *Codec) _encodeReflectBinary(w io.Writer, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {
	if rv.Kind() == reflect.Ptr {
		panic("should not happen")
	}
	if !rv.IsValid() {
		panic("should not happen")
	}

	// Handle override if rv implements json.Marshaler.
	if info.IsWireMarshaler {
		// First, encode rv into repr instance.
		var rrv, rinfo = reflect.Value{}, (*TypeInfo)(nil)
		rrv, err = toReprObject(rv)
		if err != nil {
			return
		}
		rinfo, err = cdc.getTypeInfo_wlock(info.WireMarshalReprType)
		if err != nil {
			return
		}
		// Then, encode the repr instance.
		err = cdc._encodeReflectBinary(w, rinfo, rrv, opts)
		return
	}

	switch info.Type.Kind() {

	//----------------------------------------
	// Complex

	case reflect.Interface:
		err = cdc.encodeReflectBinaryInterface(w, info, rv, opts)

	case reflect.Array:
		if info.Type.Elem().Kind() == reflect.Uint8 {
			err = cdc.encodeReflectBinaryByteArray(w, info, rv, opts)
		} else {
			err = cdc.encodeReflectBinaryList(w, info, rv, opts)
		}

	case reflect.Slice:
		if info.Type.Elem().Kind() == reflect.Uint8 {
			err = cdc.encodeReflectBinaryByteSlice(w, info, rv, opts)
		} else {
			err = cdc.encodeReflectBinaryList(w, info, rv, opts)
		}

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

	// Special case when rv is nil, write 0x0000.
	if rv.IsNil() {
		_, err = w.Write([]byte{0x00, 0x00})
		return
	}

	// Get concrete non-pointer reflect value & type.
	var crv, isPtr, isNilPtr = derefPointers(rv)
	if isPtr && crv.Kind() == reflect.Interface {
		panic(fmt.Sprintf("Unexpected interface-pointer of type *%v for registered interface %v. Not supported yet.", crv.Type(), iinfo.Type))
	}
	if isNilPtr {
		panic(fmt.Sprintf("Illegal nil-pointer of type %v for registered interface %v. "+
			"For compatibility with other languages, nil-pointer interface values are forbidden.", crv.Type(), iinfo.Type))
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

	// Write disambiguation bytes if needed.
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

	// Write prefix+typ3 bytes.
	var typ = typeToTyp3(crt, opts)
	_, err = w.Write(cinfo.Prefix.WithTyp3(typ).Bytes())
	if err != nil {
		return
	}

	// Write actual concrete value.
	err = cdc._encodeReflectBinary(w, cinfo, crv, opts)
	return
}

func (cdc *Codec) encodeReflectBinaryByteArray(w io.Writer, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {
	ert := info.Type.Elem()
	if ert.Kind() != reflect.Uint8 {
		panic("should not happen")
	}
	length := info.Type.Len()

	// Get byteslice.
	var byteslice = []byte(nil)
	if rv.CanAddr() {
		byteslice = rv.Slice(0, length).Bytes()
	} else {
		byteslice = make([]byte, length)
		reflect.Copy(reflect.ValueOf(byteslice), rv) // XXX: looks expensive!
	}

	// Write byte-length prefixed byteslice.
	err = EncodeByteSlice(w, byteslice)
	return
}

func (cdc *Codec) encodeReflectBinaryList(w io.Writer, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {
	ert := info.Type.Elem()
	if ert.Kind() == reflect.Uint8 {
		panic("should not happen")
	}

	// Write element Typ4 byte.
	var typ = typeToTyp4(ert, opts)
	err = EncodeByte(w, byte(typ))
	if err != nil {
		return
	}

	// Write length.
	err = EncodeUvarint(w, uint64(rv.Len()))
	if err != nil {
		return
	}

	// Write elems.
	var einfo *TypeInfo
	einfo, err = cdc.getTypeInfo_wlock(ert)
	if err != nil {
		return
	}
	for i := 0; i < rv.Len(); i++ {
		// Get dereferenced element value and info.
		var erv, isNil = isNilSafe(rv.Index(i))
		if typ.IsPointer() {
			// We must write a byte to denote whether element is nil.
			var isNil bool
			erv, isNil = isNilSafe(erv) // NOTE: sets erv without shadowing it.
			if isNil {
				// Value is nil.
				// e.g. nil pointer, nil slice, pointer to nil slice, pointer to nil pointer.
				// Write 0x01 for nil.
				_, err = w.Write([]byte{0x01})
				continue
			} else {
				// Value is not nil.
				// Write 0x00 for not nil.
				_, err = w.Write([]byte{0x00})
			}
		} else {
			// Do not write nil byte.
			if isNil {
				continue
			}
		}
		// Write the element value, it isn't nil.
		err = cdc.encodeReflectBinary(w, einfo, erv, opts)
		if err != nil {
			return
		}
	}
	return
}

// CONTRACT: info.Type.Elem().Kind() == reflect.Uint8
func (cdc *Codec) encodeReflectBinaryByteSlice(w io.Writer, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {
	ert := info.Type.Elem()
	if ert.Kind() != reflect.Uint8 {
		panic("should not happen")
	}

	// Write byte-length prefixed byte-slice.
	var byteslice = rv.Bytes()
	err = EncodeByteSlice(w, byteslice)
	return
}

func (cdc *Codec) encodeReflectBinaryStruct(w io.Writer, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {

	// The "Struct" Typ3 doesn't get written here.
	// It's already implied, either by struct-key or list-element-type-byte.

	switch info.Type {

	case timeType:
		// Special case: time.Time
		err = EncodeTime(w, rv.Interface().(time.Time))
		return

	default:
		for _, field := range info.Fields {
			// Get dereferenced field value and info.
			var frv, isNil = isNilSafe(rv.Field(field.Index))
			if isNil {
				continue // Do not encode nil fields.
			}
			var finfo *TypeInfo
			finfo, err = cdc.getTypeInfo_wlock(field.Type)
			if err != nil {
				return
			}
			// TODO Maybe allow omitempty somehow.
			// Write field key (number and type).
			err = encodeFieldNumberAndTyp3(w, field.BinFieldNum, field.BinTyp3)
			if err != nil {
				return
			}
			// Write field from rv.
			err = cdc.encodeReflectBinary(w, finfo, frv, field.FieldOptions)
			if err != nil {
				return
			}
		}

		// Write "StructTerm".
		err = EncodeByte(w, byte(Typ3_StructTerm))
		if err != nil {
			return
		}
		return

	}

}

//----------------------------------------
// Misc.

// Write field key.
func encodeFieldNumberAndTyp3(w io.Writer, num uint32, typ Typ3) (err error) {
	if (typ & 0xF8) != 0 {
		panic(fmt.Sprintf("invalid Typ3 byte %X", typ))
	}
	if num < 0 || num > (1<<29-1) {
		panic(fmt.Sprintf("invalid field number %v", num))
	}

	// Pack Typ3 and field number.
	var value64 = (uint64(num) << 3) | uint64(typ)

	// Write uvarint value for field and Typ3.
	var buf [10]byte
	n := binary.PutUvarint(buf[:], value64)
	_, err = w.Write(buf[0:n])
	return
}
