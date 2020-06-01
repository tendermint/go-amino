package genproto

import (
	"errors"
	"reflect"

	"github.com/tendermint/go-amino"
)

// Given a codec and some reflection type, generate the Proto3 message
// (partial) schema.
//
func GenerateProto3MessageSchema(cdc *amino.Codec, rt reflect.Type) (p3msg P3Message, err error) {

	var info *amino.TypeInfo
	info, err = cdc.GetTypeInfo(rt)
	if err != nil {
		return
	}
	if info.Type.Kind() != reflect.Struct {
		err = errors.New("only structs can generate proto3 message schemas")
		return
	}

	// When fields include other declared structs,
	// we need to know whether it's an external reference
	// (with corresponding imports in the proto3 schema)
	// or an internal reference (with no imports necessary).
	var pkgPath = rt.PkgPath()
	if pkgPath == "" {
		err = errors.New("can only generate proto3 message schemas from user-defined package-level declared structs")
		return
	}

	for _, field := range info.StructInfo.Fields {
		p3Type, p3FieldRepeated :=
			reflectTypeToP3Type(cdc, field.Type)
		p3Field := P3Field{
			Repeated: p3FieldRepeated,
			Type:     p3Type,
			Name:     field.Name,
			Number:   field.FieldOptions.BinFieldNum,
		}
		p3Field.Repeated = p3FieldRepeated
		p3msg.Fields = append(p3msg.Fields, p3Field)
	}

	return
}

// NOTE: if rt is a struct, the returned proto3 type is
// NewCustomP3Type(<amino name>).
func reflectTypeToP3Type(cdc *amino.Codec, rt reflect.Type) (p3type P3Type, repeated bool) {

	var info *amino.TypeInfo
	var err error
	info, err = cdc.GetTypeInfo(rt)
	if err != nil {
		return
	}

	switch rt.Kind() {
	case reflect.Bool:
		return P3TypeBool, false
	case reflect.Int:
		return P3TypeInt64, false
	case reflect.Int8:
		return P3TypeInt32, false
	case reflect.Int16:
		return P3TypeInt32, false
	case reflect.Int32:
		return P3TypeInt32, false
	case reflect.Int64:
		return P3TypeInt64, false
	case reflect.Uint:
		return P3TypeUint64, false
	case reflect.Uint8:
		return P3TypeUint32, false
	case reflect.Uint16:
		return P3TypeUint32, false
	case reflect.Uint32:
		return P3TypeUint32, false
	case reflect.Uint64:
		return P3TypeUint64, false
	case reflect.Float32:
		return P3TypeFloat, false
	case reflect.Float64:
		return P3TypeDouble, false
	case reflect.Complex64, reflect.Complex128:
		panic("complex types not yet supported")
	case reflect.Array, reflect.Slice:
		switch rt.Elem().Kind() {
		case reflect.Uint8:
			return P3TypeBytes, false
		default:
			elemP3Type, elemRepeated := reflectTypeToP3Type(cdc, rt.Elem())
			if elemRepeated {
				panic("multi-dimensional arrays not yet supported")
			}
			return elemP3Type, true
		}
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr,
		reflect.UnsafePointer:
		panic("chan, func, map, and pointers are not supported")
	case reflect.String:
		return P3TypeString, false
	case reflect.Struct:
		// XXX if the package is different than the current package...
		return NewCustomP3Type(info.Type.Name()), false
	default:
		panic("unexpected rt kind")
	}

}
