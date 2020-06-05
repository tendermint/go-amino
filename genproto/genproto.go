package genproto

// p3c.SetProjectRootGopkg("example.com/main")

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"reflect"

	"github.com/tendermint/go-amino"
)

// TODO sort
//  * Proto3 import file paths are by default always full (including domain)
//    and basically the p3importPrefix plus the gopkg path.
//    This lets proto3 schema import paths stay consistent even as dependency.
//  * In the go mod world, the user is expected to run an independent tool
//    to copy proto files to the p3importPrefix folder from go mod dependencies.

// P3Context holds contextual information beyond the P3Doc.
//
// It holds all the package infos needed to derive the full P3doc,
// including p3 import paths, as well as where to write them,
// because all of that information is encapsulated in genproto.PackageInfo.
type P3Context struct {
	// e.g. "github.com/tendermint/tendermint/abci/types" ->
	//   &PackageInfo{...}
	go2PackageInfo map[string]*PackageInfo

	// TODO
	// // for beyond default "type.proto"
	// // e.g. "tendermint.abci.types" ->
	// //   []string{"proto/github.com/tendermint/abci/types/types.proto"}}
	// moreP3Imports map[string][]string

	// Proto 3 schema files are found in
	// "{p3importPrefix}{gopkg}/types.proto"
	p3importPrefix string

	// Singleton for this P3Context.
	// Interface/disfix info not relevant for proto3, so there is no need for
	// the user to access this codec.  In other words, the default codec is
	// sufficient.
	cdc *amino.Codec
}

func NewP3Context() *P3Context {
	return &P3Context{
		go2PackageInfo: make(map[string]*PackageInfo),
		p3importPrefix: "proto/",
		cdc:            amino.NewCodec(),
	}
}

func (p3c *P3Context) RegisterPackage(pi *PackageInfo) {
	infos := crawlPackageInfos(pi, nil)
	for _, info := range infos {
		p3c.registerPackage(info)
	}
}

func (p3c *P3Context) registerPackage(pi *PackageInfo) {
	if found, ok := p3c.go2PackageInfo[pi.GoPkg]; ok {
		if found != pi {
			panic(fmt.Errorf("found conflicting package mapping, %v -> %v but trying to overwrite with -> %v", pi.GoPkg, found, pi))
		}
	}
	p3c.go2PackageInfo[pi.GoPkg] = pi
}

func (p3c *P3Context) GetPackageInfo(gopkg string) *PackageInfo {
	pi, ok := p3c.go2PackageInfo[gopkg]
	if !ok {
		panic(fmt.Sprintf("package info unrecognized for %v (not registered directly nor indirectly as dependency", gopkg))
	}
	return pi
}

// For a given package info, crawl and discover all package infos.
func crawlPackageInfos(info *PackageInfo, seen map[*PackageInfo]struct{}) (res []*PackageInfo) {
	if seen == nil {
		seen = map[*PackageInfo]struct{}{}
	}
	var crawl func(info *PackageInfo)
	crawl = func(info *PackageInfo) {
		seen[info] = struct{}{}
		for _, dependency := range info.Dependencies {
			if _, ok := seen[dependency]; ok {
				continue
			}
			crawl(dependency)
		}
	}
	crawl(info)
	for info, _ := range seen {
		res = append(res, info)
	}
	return res
}

// Crawls the package infos and flattens all dependencies.
func (p3c *P3Context) GetAllPackageInfos() (res []*PackageInfo) {
	seen := map[*PackageInfo]struct{}{}
	for _, info := range p3c.go2PackageInfo {
		infos := crawlPackageInfos(info, seen)
		res = append(res, infos...)
	}
	return
}

func (p3c *P3Context) ValidateBasic() {
	infos := p3c.GetAllPackageInfos()
	for _, info := range infos {
		info.ValidateBasic()
	}
}

func (p3c *P3Context) GetImportPath(p3type P3Type) string {
	p3pkg := p3type.GetPackage()
	infos := p3c.GetAllPackageInfos()
	for _, info := range infos {
		if info.P3Pkg == p3pkg {
			return path.Join(info.Dirname, "types.proto")
		}
	}
	panic(fmt.Sprintf("proto3 package %v not recognized", p3pkg))
}

// Given a codec and some reflection type, generate the Proto3 message
// (partial) schema.  Imports are added to p3doc.
func (p3c *P3Context) GenerateProto3MessagePartial(p3doc *P3Doc, rt reflect.Type) (p3msg P3Message, err error) {

	var info *amino.TypeInfo
	info, err = p3c.cdc.GetTypeInfo(rt)
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
		p3FieldType, p3FieldRepeated :=
			p3c.reflectTypeToP3Type(field.Type)
		// If the p3 field package is the same, omit the prefix.
		fmt.Println(">>", p3FieldType, ">>", p3doc, ">>", field.Type)
		if p3FieldType.GetPackage() == p3doc.Package {
			p3FieldMessageType := p3FieldType.(P3MessageType)
			p3FieldMessageType.SetOmitPackage()
			p3FieldType = p3FieldMessageType
		}
		// If the field package different, add the import to p3doc.
		if field.Type.PkgPath() != pkgPath {
			if p3FieldType.GetPackage() != "" {
				importPath := p3c.GetImportPath(p3FieldType)
				p3doc.AddImport(importPath)
			}
		}
		p3Field := P3Field{
			Repeated: p3FieldRepeated,
			Type:     p3FieldType,
			Name:     field.Name,
			Number:   field.FieldOptions.BinFieldNum,
		}
		p3Field.Repeated = p3FieldRepeated
		p3msg.Fields = append(p3msg.Fields, p3Field)
	}

	return
}

// Given the arguments, create a new P3Doc.
// pkg is optional.
func (p3c *P3Context) GenerateProto3Schema(p3pkg string, rtz ...reflect.Type) (p3doc P3Doc, err error) {

	// Set the package.
	p3doc.Package = p3pkg

	// Set Message schemas.
	for _, rt := range rtz {
		p3msg, err := p3c.GenerateProto3MessagePartial(&p3doc, rt)
		if err != nil {
			return P3Doc{}, err
		}
		p3doc.Messages = append(p3doc.Messages, p3msg)
	}

	return p3doc, nil
}

// Convenience.
func (p3c *P3Context) WriteProto3Schema(filename string, p3pkg string, rtz ...reflect.Type) (err error) {
	fmt.Printf("writing proto3 schema to %v for package %v\n", filename, p3pkg)
	p3doc, err := p3c.GenerateProto3Schema(p3pkg, rtz...)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filename, []byte(p3doc.Print()), 0644)
	return err
}

// If rt is a struct, the returned proto3 type is a P3MessageType.
func (p3c *P3Context) reflectTypeToP3Type(rt reflect.Type) (p3type P3Type, repeated bool) {

	// If the kind is an interface type,
	// just return an any.
	// The p3c.cdc is a blank one with no registrations,
	// so there isn't much else we can do atm.
	if rt.Kind() == reflect.Interface {
		return P3AnyType, false
	}

	var info *amino.TypeInfo
	var err error
	info, err = p3c.cdc.GetTypeInfo(rt)
	if err != nil {
		panic(err)
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
			elemP3Type, elemRepeated := p3c.reflectTypeToP3Type(rt.Elem())
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
		// Look up the p3pkg type from p3 context.
		pkginfo := p3c.GetPackageInfo(info.Type.PkgPath())
		return NewP3MessageType(pkginfo.P3Pkg, info.Type.Name()), false
	default:
		panic("unexpected rt kind")
	}

}

func WriteProto3Schemas(infos ...*PackageInfo) {
	for _, info := range infos {
		p3c := NewP3Context()
		p3c.RegisterPackage(info)
		p3c.ValidateBasic()
		filename := path.Join(info.Dirname, "types.proto")
		err := p3c.WriteProto3Schema(filename, info.P3Pkg, info.StructTypes...)
		if err != nil {
			panic(err)
		}
	}
}
