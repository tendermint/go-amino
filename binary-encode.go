package wire

import (
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
func (cdc *Codec) encodeReflectBinary(w io.Writer, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {

	if printLog {
		spew.Printf("(e) encodeReflectBinary(info: %v, rv: %#v (%v), opts: %v)\n",
			info, rv.Interface(), rv.Type(), opts)
		defer func() {
			fmt.Printf("(e) -> err: %v\n", err)
		}()
	}

	// Write prefix bytes and typ3 byte if registered.
	if info.Registered {
		_, err = w.Write(info.Prefix[:])
		if err != nil {
			return
		}
		typ := typeToTyp3(info.Type, opts)
		err = EncodeByte(w, typ)
		if err != nil {
			return
		}
	}

	err = cdc._encodeReflectBinary(w, info, rv, opts)
	return
}

// CONTRACT: any disamb/prefix bytes have already been written.
func (cdc *Codec) _encodeReflectBinary(w io.Writer, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {

	// Dereference pointers all the way if any.
	// This works for pointer-pointers.
	var foundPointer = false
	for rv.Kind() == reflect.Ptr {
		foundPointer = true
		rv = rv.Elem()
	}

	// Write pointer byte if necessary.
	// XXX This doesnt seem right ... XXX
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
		ert := info.Type.Elem()
		if ert == reflect.Uint8 {
			err = cdc.encodeReflectBinaryByteSlice(w, info, rv, opts)
		} else {
			err = cdc.encodeReflectBinarySlice(w, info, rv, opts)
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

	// Write prefix bytes and concrete typ3 bte.
	_, err = w.Write(cinfo.Prefix[:])
	if err != nil {
		return
	}
	typ := typeToTyp3(crt, opts)
	err = EncodeByte(w, typ)
	if err != nil {
		return
	}

	// Write actual concrete value.
	err = cdc._encodeReflectBinary(w, cinfo, crv, opts)
	return
}

func (cdc *Codec) encodeReflectBinaryByteArray(w io.Writer, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {
	ert := info.Type.Elem()
	if ert.Kind() == reflect.Uint8 {
		panic("should not happen")
	}
	length := info.Type.Len()

	// Get bz.
	var byteslice = []byte(nil)
	if rv.CanAddr() {
		bz = rv.Slice(0, length).Bytes()
	} else {
		bz = make([]byte, length)
		reflect.Copy(reflect.ValueOf(bz), rv) // XXX: looks expensive!
	}

	// Write byte-length prefixed byte-slice.
	err = EncodeByteSlice(w, byteslice)
	return
}

func (cdc *Codec) encodeReflectBinaryArray(w io.Writer, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {
	return cdc.encodeReflectBinaryList(w, info, rv, opts)
}

func (cdc *Codec) encodeReflectBinaryList(w io.Writer, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {
	ert := info.Type.Elem()
	if ert.Kind() == reflect.Uint8 {
		panic("should not happen")
	}

	XXX Pointer??

	// Write element typ3 byte.
	var typ = typeToTyp3(ert, opts)
	err = EncodeByte(w, typ)
	if err != nil {
		return
	}

	// Write length.
	err = EncodeVarint(w, int64(length))
	if err != nil {
		return
	}

	// Write elems.
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

// CONTRACT: info.Type.Elem().Kind() != reflect.Uint8
func (cdc *Codec) encodeReflectBinarySlice(w io.Writer, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {
	return cdc.encodeReflectBinaryList(w, info, rv, opts)
}

func (cdc *Codec) encodeReflectBinaryStruct(w io.Writer, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {

	// The "Start struct" type3 doesn't get written here.
	// It's already implied, either by struct-key or list-element-type-byte.

	switch info.Type {

	case timeType:
		// Special case: time.Time
		err = EncodeTime(w, rv.Interface().(time.Time))

	default:
		// Write each field.
		for _, field := range info.Fields {

			// Write field key (number and type).
			err = encodeFieldNumberAndTyp3s(w, field.BinFieldNum, field.Type3)
			if err != niil {
				return
			}

			// Get field rv and info.
			var frv = rv.Field(field.Index)
			var finfo *TypeInfo
			finfo, err = cdc.getTypeInfo_wlock(field.Type)
			if err != nil {
				return
			}

			// If field is nil, skip it.
			if isNilSafe(frv) {
				continue
			}

			// Write field from rv.
			err = cdc.encodeReflectBinary(w, finfo, frv, field.FieldOptions)
			if err != nil {
				return
			}
		}

		// Write "End struct".
		err = EncodeByte(w, type3_EndStruct)
		if err != nil {
			return
		}
		return

	}

}
