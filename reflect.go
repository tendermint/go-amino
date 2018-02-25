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

//----------------------------------------
// Typ3 and Typ4

type typ3 uint8
type typ4 uint8 // typ3 | 0x80 (pointer bit)

const (
	// typ3 types
	typ3_Varint     = typ3(0)
	typ3_8Byte      = typ3(1)
	typ3_ByteLength = typ3(2)
	typ3_Struct     = typ3(3)
	typ3_StructTerm = typ3(4)
	typ3_4Byte      = typ3(5)
	typ3_List       = typ3(6)
	typ3_Interface  = typ3(7)

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

// Like typeToTyp4 but include a pointer bit.
func typeToTyp4(rt reflect.Type, opts FieldOptions) (typ typ4) {

	// Transparently "dereference" pointer type.
	var pointer = false
	for rt.Kind() == reflect.Ptr {
		pointer = true
		rt = rt.Elem()
	}

	// Call actual logic.
	typ = typ4(typeToTyp3(rt, opts))

	// Set pointer bit to 1 if pointer.
	if pointer {
		typ |= typ4_Pointer
	}
	return
}

// CONTRACT: rt.Kind() != reflect.Ptr
func typeToTyp3(rt reflect.Type, opts FieldOptions) typ3 {
	switch rt.Kind() {
	case reflect.Interface:
		return typ3_Interface
	case reflect.Array, reflect.Slice:
		ert := rt.Elem()
		switch ert.Kind() {
		case reflect.Uint8:
			return typ3_ByteLength
		default:
			return typ3_List
		}
	case reflect.String:
		return typ3_ByteLength
	case reflect.Struct:
		return typ3_Struct
	case reflect.Int64, reflect.Uint64:
		if opts.BinVarint {
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

func isNilSafe(rv reflect.Value) bool {
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface,
		reflect.Map, reflect.Ptr, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}
