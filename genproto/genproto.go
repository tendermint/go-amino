package genproto

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/tendermint/go-amino"
)

// P3Context holds contextual information beyond the P3Doc.
//
// It holds a mapping from Go package paths to proto3 package names.  This
// struct doesn't keep track of third party proto file locations, that is
// considered orthogonal.
// NOTE: It is assumed that amino structs declared in the same Go package share
// the same proto3 package name.  This contract should not change.
type P3Context struct {
	// e.g. "github.com/tendermint/tendermint/abci/types" ->
	//   "tendermint.abci.types"
	go2p3pkg map[string]string

	// e.g. "tendermint.abci.types" -> []string{"vendor/github.com/tendermint/abci/types/types.proto"}}
	p3imports map[string][]string
	// By default, proto 3 schema files are found in "{p3importPrefix}{gopkg}/types.proto"
	p3importPrefix string
}

func NewP3Context() *P3Context {
	return &P3Context{
		go2p3pkg:       make(map[string]string),
		p3imports:      make(map[string][]string),
		p3importPrefix: "vendor/",
	}
}

func (p3c *P3Context) RegisterPackageMapping(gopkg string, p3pkg string, imports []string) error {
	if foundP3pkg, ok := p3c.go2p3pkg[gopkg]; ok {
		if foundP3pkg != p3pkg {
			return fmt.Errorf("found conflicting package mapping, %v -> %v but trying to overwrite with -> %v", gopkg, foundP3pkg, p3pkg)
		}
	}
	p3c.go2p3pkg[gopkg] = p3pkg
	p3c.p3imports[p3pkg] = append(p3c.p3imports[p3pkg], imports...)
	return nil
}

func (p3c *P3Context) GetP3Package(gopkg string) (p3pkg string, ok bool) {
	p3pkg, ok = p3c.go2p3pkg[gopkg]
	return
}

// If not found, p3pkg is derived automatically from the gopkg, and ok is still true.
// There is no conflict resolution.
// If no import files for p3pkg were registered, the default "types.proto" in the
// "/vendors" directory is used.
func (p3c *P3Context) GetP3PackageOrDefault(gopkg string) (p3pkg string) {
	p3pkg, ok := p3c.go2p3pkg[gopkg]
	if !ok {
		p3pkg = DefaultP3pkgFromGopkg(gopkg)
		// precautionary warning
		for gopkgFound, p3pkgFound := range p3c.go2p3pkg {
			if p3pkgFound == p3pkg {
				// Mapping isn't 1:1, more than one gopaths could map to the
				// same proto3 package, but is it what you want?
				fmt.Printf("WARNING, proto3 package %v already registered with %v, but also is derived from %v", p3pkg, gopkgFound, gopkg)
			}
		}
		// If no files for p3pkg were registered,
		// also derive the p3 schema file import file.
		if _, ok := p3c.p3imports[p3pkg]; !ok {
			p3import := p3c.p3importPrefix + gopkg + "/types.proto" // TODO make safe
			p3c.p3imports[p3pkg] = []string{p3import}
		}
		return
	}
	return
}

func DefaultP3pkgFromGopkg(gopkg string) string {
	if gopkg == "" {
		panic("gopkg cannot be empty")
	}
	parts := strings.Split(gopkg, "/")
	// if parts[0] is domain TODO
	afterDomain := parts[1:]
	return strings.Join(afterDomain, ".")
}

// Given a codec and some reflection type, generate the Proto3 message
// (partial) schema.
func (p3c *P3Context) GenerateProto3MessagePartial(cdc *amino.Codec, rt reflect.Type) (p3msg P3Message, err error) {

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

	p3msg.Name = info.Type.Name()

	for _, field := range info.StructInfo.Fields {
		p3Type, p3FieldRepeated :=
			p3c.reflectTypeToP3Type(cdc, field.Type)
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

// Given the arguments, create a new P3Doc.
func (p3c *P3Context) GenerateProto3Schema(cdc *amino.Codec, rtz ...reflect.Type) (p3doc P3Doc, err error) {
	for _, rt := range rtz {
		p3msg, err := p3c.GenerateProto3MessagePartial(cdc, rt)
		if err != nil {
			return P3Doc{}, err
		}
		p3doc.Messages = append(p3doc.Messages, p3msg)
	}
	for _, filenames := range p3c.p3imports {
		for _, filename := range filenames {
			p3imp := P3Import{Path: filename}
			p3doc.Imports = append(p3doc.Imports, p3imp)
		}
	}
	return p3doc, nil
}

// NOTE: if rt is a struct, the returned proto3 type is
// a P3MessageType.
func (p3c *P3Context) reflectTypeToP3Type(cdc *amino.Codec, rt reflect.Type) (p3type P3Type, repeated bool) {

	var info *amino.TypeInfo
	var err error
	info, err = cdc.GetTypeInfo(rt)
	if err != nil {
		return
	}

	switch rt.Kind() {
	case reflect.Bool:
		return P3ScalarTypeBool, false
	case reflect.Int:
		return P3ScalarTypeInt64, false
	case reflect.Int8:
		return P3ScalarTypeInt32, false
	case reflect.Int16:
		return P3ScalarTypeInt32, false
	case reflect.Int32:
		return P3ScalarTypeInt32, false
	case reflect.Int64:
		return P3ScalarTypeInt64, false
	case reflect.Uint:
		return P3ScalarTypeUint64, false
	case reflect.Uint8:
		return P3ScalarTypeUint32, false
	case reflect.Uint16:
		return P3ScalarTypeUint32, false
	case reflect.Uint32:
		return P3ScalarTypeUint32, false
	case reflect.Uint64:
		return P3ScalarTypeUint64, false
	case reflect.Float32:
		return P3ScalarTypeFloat, false
	case reflect.Float64:
		return P3ScalarTypeDouble, false
	case reflect.Complex64, reflect.Complex128:
		panic("complex types not yet supported")
	case reflect.Array, reflect.Slice:
		switch rt.Elem().Kind() {
		case reflect.Uint8:
			return P3ScalarTypeBytes, false
		default:
			elemP3Type, elemRepeated := p3c.reflectTypeToP3Type(cdc, rt.Elem())
			if elemRepeated {
				panic("multi-dimensional arrays not yet supported")
			}
			return elemP3Type, true
		}
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr,
		reflect.UnsafePointer:
		panic("chan, func, map, and pointers are not supported")
	case reflect.String:
		return P3ScalarTypeString, false
	case reflect.Struct:
		// TODO if the package is the same as the container's package,
		// no need to set the p3pkg name, it can be empty.
		p3pkg := p3c.GetP3PackageOrDefault(info.Type.PkgPath())
		fmt.Println("---", info.Type.Name())
		return NewP3MessageType(p3pkg, info.Type.Name()), false
	default:
		panic("unexpected rt kind")
	}

}
