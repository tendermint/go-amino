package wire

import (
	"fmt"
	"reflect"
	"time"
)

//----------------------------------------
// Constants

var timeType = reflect.TypeOf(time.Time{})

const RFC3339Millis = "2006-01-02T15:04:05.000Z" // forced microseconds
const printLog = false

/*
Type | Meaning      | Used For
---- | ------------ | --------
0    | Varint       | bool, byte, [u]int16, and varint-[u]int[64/32]
1    | 8 byte       | int64, uint64, float64(unsafe)
2    | Byte length  | string, bytes, raw?
3    | Start struct | conceptually, '{'
4    | End struct   | conceptually, '}'; always a single byte, 0x04
5    | 4 byte       | int32, uint32, float32(unsafe)
6    | List         | array, slice; followed by 1 or more bytes encoding `<type-bytes>`,
     |              | then `<uvarint(num-items)>`
7    | Interface    | value starts with `<prefix-bytes>` or `<disfix-bytes>`
*/

type typ3 uint8
type typ4 uint8 // typ3 | 0x80 (pointer bit)

const (
	// typ3 types
	typ3_Varint      = typ3(0)
	typ3_8Byte       = typ3(1)
	typ3_ByteLength  = typ3(2)
	typ3_StartStruct = typ3(3)
	typ3_EndStruct   = typ3(4)
	typ3_4Byte       = typ3(5)
	typ3_List        = typ3(6)
	typ3_Interface   = typ3(7)

	// typ4 bit
	typ4_Pointer = typ4(0x08)
)

//----------------------------------------
// encode: see binary-encode.go and json-encode.go
// decode: see binary-decode.go and json-decode.go

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

func typeToTyp4(rt runtime.Type, fopts FieldOptions) (typ4 typ4) {

	// Transparently "dereference" pointer type.
	var pointer bool = false
	for rt.Kind() == reflect.Ptr {
		pointer = true
		rt = rt.Elem()
	}

	// Call actual logic.
	typ4 = typ4(typeToTyp3(rt, fopts))

	// Set pointer bit to 1 if pointer.
	if pointer {
		typ4 |= typ4_Pointer
	}
	return
}

// CONTRACT: rt.Kind() != reflect.Ptr
func typeToTyp3(rt runtime.Type, fopts FieldOptions) (typ typ3) {
	switch rt.Kind() {
	case reflect.Interface:
		return typ3_Interface
	case reflect.Array, reflect.Slice:
		ert := rt.Elem()
		switch ert.Kind() {
		case reflect.Uint8:
			return typ3_ByteLength
		default:
			return type3_List
		}
	case reflect.String:
		return typ3_ByteLength
	case reflect.Struct:
		return typ3_StartStruct
	case reflect.Int64, reflect.Uint64:
		if fopts.BinVarint {
			return typ3_Varint
		}
		return typ3_8Byte
	case reflect.Float64:
		return typ3_8Byte
	case reflect.Int32, reflect.Uint32, reflect.Float32:
		return typ3_4Byte
	case reflect.Int16, reflect.Int8, reflect.Int,
		reflect.Uint16, reflect.Uint8, reflect.Uint, reflect.Bool:
		return typ3_Varint
	default:
		panic(fmt.Sprintf("unsupported field type %v", rt))
	}
}

func encodeFieldNumberAndTyp3s(w io.Writer, num int32, typ3s []typ3) (err error) {
	var typ = typ3s[0]
	if (typ & 0xF8) > 0 {
		panic(fmt.Sprintf("invalid typ3 bytes %X (see first typ3 byte)" + typ3s))
	}
	if num < 0 || num > (1<<29-1) {
		panic(fmt.Sprintf("invalid field number %v" + num))
	}
	value := (int64(num) << 3) | typ

	// Write uvarint value for field and first typ3 byte.
	var buf [10]byte
	n := binary.PutUvarint(buf[:], value)
	buf = buf[0:n]
	_, err = w.Write(buf)

	// Write remaining typ3 bytes.
	if len(typ3s) > 1 {
		_, err = w.Write(typ3s[1:])
		if err != nil {
			return
		}
	}
	return
}

func isNilSafe(rv reflect.Value) bool {
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface,
		reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}
