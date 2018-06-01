package amino

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
// The value may be a nil interface, but not a nil pointer.
// The following contracts apply to all similar encode methods.
// CONTRACT: rv is not a pointer
// CONTRACT: rv is valid.
func (cdc *Codec) encodeReflectBinary(w io.Writer, info *TypeInfo, rv reflect.Value, isRoot bool, fopts FieldOptions) (err error) {
	if rv.Kind() == reflect.Ptr {
		panic("should not happen")
	}
	if !rv.IsValid() {
		panic("should not happen")
	}
	if printLog {
		spew.Printf("(E) encodeReflectBinary(info: %v, rv: %#v (%v), isRoot: %v, fopts: %v)\n",
			info, rv.Interface(), rv.Type(), isRoot, fopts)
		defer func() {
			fmt.Printf("(E) -> err: %v\n", err)
		}()
	}

	// Handle override if rv implements json.Marshaler.
	if info.IsAminoMarshaler {
		// First, encode rv into repr instance.
		var rrv, rinfo = reflect.Value{}, (*TypeInfo)(nil)
		rrv, err = toReprObject(rv)
		if err != nil {
			return
		}
		rinfo, err = cdc.getTypeInfo_wlock(info.AminoMarshalReprType)
		if err != nil {
			return
		}
		// Then, encode the repr instance.
		err = cdc.encodeReflectBinary(w, rinfo, rrv, isRoot, fopts)
		return
	}

	switch info.Type.Kind() {

	//----------------------------------------
	// Complex

	case reflect.Interface:
		err = cdc.encodeReflectBinaryInterface(w, info, rv, isRoot, fopts)

	case reflect.Array:
		if info.Type.Elem().Kind() == reflect.Uint8 {
			err = cdc.encodeReflectBinaryByteArray(w, info, rv, fopts)
		} else {
			err = cdc.encodeReflectBinaryList(w, info, rv, fopts)
		}

	case reflect.Slice:
		if info.Type.Elem().Kind() == reflect.Uint8 {
			err = cdc.encodeReflectBinaryByteSlice(w, info, rv, fopts)
		} else {
			err = cdc.encodeReflectBinaryList(w, info, rv, fopts)
		}

	case reflect.Struct:
		err = cdc.encodeReflectBinaryStruct(w, info, rv, isRoot, fopts)

	//----------------------------------------
	// Signed

	case reflect.Int64:
		if fopts.BinVarint {
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
		if fopts.BinVarint {
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
		if !fopts.Unsafe {
			err = errors.New("Amino float* support requires `amino:\"unsafe\"`.")
			return
		}
		err = EncodeFloat64(w, rv.Float())

	case reflect.Float32:
		if !fopts.Unsafe {
			err = errors.New("Amino float* support requires `amino:\"unsafe\"`.")
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

func (cdc *Codec) encodeReflectBinaryInterface(w io.Writer, iinfo *TypeInfo, rv reflect.Value, isRoot bool, fopts FieldOptions) (err error) {
	if printLog {
		fmt.Println("(e) encodeReflectBinaryInterface")
		defer func() {
			fmt.Printf("(e) -> err: %v\n", err)
		}()
	}

	// Special case when rv is nil, write 0x0000.
	if rv.IsNil() {
		_, err = w.Write([]byte{0x00, 0x00})
		return
	}

	// Get concrete non-pointer reflect value & type.
	var crv, isPtr, isNilPtr = derefPointers(rv.Elem())
	if isPtr && crv.Kind() == reflect.Interface {
		// See "MARKER: No interface-pointers" in codec.go
		panic("should not happen")
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
	var typ = typeToTyp3(crt, fopts)
	_, err = w.Write(cinfo.Prefix.WithTyp3(typ).Bytes())
	if err != nil {
		return
	}

	// Write actual concrete value.
	err = cdc.encodeReflectBinary(w, cinfo, crv, isRoot, fopts)
	return
}

func (cdc *Codec) encodeReflectBinaryByteArray(w io.Writer, info *TypeInfo, rv reflect.Value, fopts FieldOptions) (err error) {
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

func (cdc *Codec) encodeReflectBinaryList(w io.Writer, info *TypeInfo, rv reflect.Value, fopts FieldOptions) (err error) {
	if printLog {
		fmt.Println("(e) encodeReflectBinaryList")
		defer func() {
			fmt.Printf("(e) -> err: %v\n", err)
		}()
	}
	ert := info.Type.Elem()
	if ert.Kind() == reflect.Uint8 {
		panic("should not happen")
	}
	einfo, err := cdc.getTypeInfo_wlock(ert)
	if err != nil {
		return
	}

	// If elem is not already a ByteLength type, write in packed form.
	// This is a Proto wart due to Proto backwards compatibility issues.
	// Amino2 will probably migrate to use the List typ3.
	typ3 := typeToTyp3(einfo.Type, fopts)
	if typ3 != Typ3_ByteLength {
		// JAE: Proto's byte-length prefixing incurs alloc cost on the encoder.
		buf := bytes.NewBuffer()
		// Write elems in packed form.
		for i := 0; i < rv.Len(); i++ {
			// Get dereferenced element value and info.
			var erv, isDefault = isDefaultValue(rv.Index(i))
			if isDefault {
				// Nothing to encode, so the length is 0.
				err = EncodeByte(buf, byte(0x00))
				if err != nil {
					return
				}
			} else {
				// Write the element value.
				// It may be a nil interface, but not a nil pointer.
				err = cdc.encodeReflectBinary(buf, einfo, erv, false, fopts)
				if err != nil {
					return
				}
			}
		}
		// Write byte-length prefixed byteslice.
		err = EncodeByteSlice(w, buf.Bytes())
		return
	} else {
		// Write elems in unpacked form.
		for i := 0; i < rv.Len(); i++ {
			// Written elements as repeated field key of the parent struct.
			err = encodeFieldNumberAndTyp3(w, fopts.BinFieldNum, Typ3_ByteLength)
			if err != nil {
				return
			}
			// Get dereferenced element value and info.
			var erv, isDefault = isDefaultValue(rv.Index(i))
			if isDefault {
				// Nothing to encode, so the length is 0.
				err = EncodeByte(w, byte(0x00))
				if err != nil {
					return
				}
			} else {
				// Write the element value to a buffer.
				// It may be a nil interface, but not a nil pointer.
				buf := bytes.NewBuffer()
				err = cdc.encodeReflectBinary(buf, einfo, erv, false, fopts)
				if err != nil {
					return
				}
				// Write byte-length prefixed byteslice.
				err = EncodeByteSlice(w, buf.Bytes())
				return
			}
		}
	}

	return
}

// CONTRACT: info.Type.Elem().Kind() == reflect.Uint8
func (cdc *Codec) encodeReflectBinaryByteSlice(w io.Writer, info *TypeInfo, rv reflect.Value, fopts FieldOptions) (err error) {
	if printLog {
		fmt.Println("(e) encodeReflectBinaryByteSlice")
		defer func() {
			fmt.Printf("(e) -> err: %v\n", err)
		}()
	}
	ert := info.Type.Elem()
	if ert.Kind() != reflect.Uint8 {
		panic("should not happen")
	}

	// Write byte-length prefixed byte-slice.
	var byteslice = rv.Bytes()
	err = EncodeByteSlice(w, byteslice)
	return
}

func (cdc *Codec) encodeReflectBinaryStruct(w io.Writer, info *TypeInfo, rv reflect.Value, isRoot bool, fopts FieldOptions) (err error) {
	if printLog {
		fmt.Println("(e) encodeReflectBinaryBinaryStruct")
		defer func() {
			fmt.Printf("(e) -> err: %v\n", err)
		}()
	}

	// Proto3 incurs a cost in writing non-root structs.
	// Here we incur it for root structs as well for ease of dev.
	buf := bytes.NewBuffer()

	switch info.Type {

	case timeType:
		// Special case: time.Time
		err = EncodeTime(buf, rv.Interface().(time.Time))
		if err != nil {
			return
		}

	default:
		for _, field := range info.Fields {
			// Get type info for field.
			var finfo *TypeInfo
			finfo, err = cdc.getTypeInfo_wlock(field.Type)
			if err != nil {
				return
			}
			// Get dereferenced field value and info.
			var frv, isDefault = isDefaultValue(rv.Field(field.Index))
			if isDefault {
				// Do not encode default value fields.
				continue
			}
			if field.UnpackedList {
				// Write repeated field entries for each list item.
				err = cdc.encodeReflectBinaryList(buf, finfo, frv, field.FieldOptions)
				if err != nil {
					return
				}
			} else {
				// Write field key (number and type).
				err = encodeFieldNumberAndTyp3(buf, field.BinFieldNum, field.BinTyp3)
				if err != nil {
					return
				}
				// Write field from rv.
				err = cdc.encodeReflectBinary(buf, finfo, frv, false, field.FieldOptions)
				if err != nil {
					return
				}
			}
		}
	}

	// Write byte-length prefixed byteslice.
	err = EncodeByteSlice(w, buf.Bytes())
	return
}

//----------------------------------------
// Misc.

// Write field key.
func encodeFieldNumberAndTyp3(w io.Writer, num uint32, typ Typ3) (err error) {
	if (typ & 0xF8) != 0 {
		panic(fmt.Sprintf("invalid Typ3 byte %v", typ))
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
