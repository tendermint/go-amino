package genproto

// p3c.SetProjectRootGopkg("example.com/main")

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/tendermint/go-amino"
	anypb "google.golang.org/protobuf/types/known/anypb"
)

// TODO sort
//  * Proto3 import file paths are by default always full (including
//  domain) and basically the gopkg path.  This lets proto3 schema
//  import paths stay consistent even as dependency.
//  * In the go mod world, the user is expected to run an independent
//  tool to copy proto files to a proto folder from go mod dependencies.
//  This is provided by MakeProtoFilder().

// P3Context holds contextual information beyond the P3Doc.
//
// It holds all the package infos needed to derive the full P3doc,
// including p3 import paths, as well as where to write them,
// because all of that information is encapsulated in amino.Package.
//
// It also holds a local amino.Codec instance, with package registrations
// passed through.
type P3Context struct {
	// e.g. "github.com/tendermint/tendermint/abci/types" ->
	//   &Package{...}
	packages map[string]*amino.Package

	// TODO
	// // for beyond default "type.proto"
	// // e.g. "tendermint.abci.types" ->
	// //   []string{"github.com/tendermint/abci/types/types.proto"}}
	// moreP3Imports map[string][]string

	// This is only necessary to construct TypeInfo.
	cdc *amino.Codec
}

func NewP3Context() *P3Context {
	p3c := &P3Context{
		packages: make(map[string]*amino.Package),
		cdc:      amino.NewCodec(),
	}
	// Register a singletone package for Any.
	p3c.RegisterPackage(amino.NewPackage(
		"google.golang.org/protobuf/types/known/anypb",
		"google.protobuf",
		"",
	).WithP3ImportPath("google/protobuf/any.proto").
		WithP3SchemaFile("").
		WithTypes(&anypb.Any{}))
	return p3c
}

func (p3c *P3Context) RegisterPackage(pkg *amino.Package) {
	pkgs := pkg.CrawlPackages(nil)
	for _, pkg := range pkgs {
		p3c.registerPackage(pkg)
	}
}

func (p3c *P3Context) registerPackage(pkg *amino.Package) {
	if found, ok := p3c.packages[pkg.GoPkgPath]; ok {
		if found != pkg {
			panic(fmt.Errorf("found conflicting package mappkgng, %v -> %v but trying to overwrite with -> %v", pkg.GoPkgPath, found, pkg))
		}
	}
	p3c.packages[pkg.GoPkgPath] = pkg
}

func (p3c *P3Context) GetPackage(gopkg string) *amino.Package {
	pkg, ok := p3c.packages[gopkg]
	if !ok {
		panic(fmt.Sprintf("package info unrecognized for %v (not registered directly nor indirectly as dependency", gopkg))
	}
	return pkg
}

// Crawls the packages and flattens all dependencies.
func (p3c *P3Context) GetAllPackages() (res []*amino.Package) {
	seen := map[*amino.Package]struct{}{}
	for _, pkg := range p3c.packages {
		pkgs := pkg.CrawlPackages(seen)
		res = append(res, pkgs...)
	}
	return
}

func (p3c *P3Context) ValidateBasic() {
	// TODO: do verifications across packages.
	// pkgs := p3c.GetAllPackages()
}

// TODO: This could live as a method of the package, and only crawl the
// dependencies of that package.  But a method implemented on P3Context
// should function like this and print an intelligent error.
func (p3c *P3Context) GetP3ImportPath(p3type P3Type) string {
	p3pkg := p3type.GetPackage()
	pkgs := p3c.GetAllPackages()
	for _, pkg := range pkgs {
		if pkg.P3PkgName == p3pkg {
			if pkg.HasName(p3type.GetName()) {
				return pkg.P3ImportPath
			}
		}
	}
	panic(fmt.Sprintf("proto3 type %v not recognized", p3type))
}

// Given a codec and some reflection type, generate the Proto3 message
// (partial) schema.  Imports are added to p3doc.
func (p3c *P3Context) GenerateProto3MessagePartial(p3doc *P3Doc, rt reflect.Type) (p3msg P3Message, err error) {

	if p3doc.Package == "" {
		err = errors.New("cannot generate message partials in the root package \"\".")
		return
	}

	info, err := p3c.cdc.GetTypeInfo(rt)
	if err != nil {
		return
	} else if info.Type.Kind() != reflect.Struct {
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
			p3c.reflectTypeToP3Type(field.TypeInfo.ReprType.Type)
		// If the p3 field package is the same, omit the prefix.
		if p3FieldType.GetPackage() == p3doc.Package {
			p3FieldMessageType := p3FieldType.(P3MessageType)
			p3FieldMessageType.SetOmitPackage()
			p3FieldType = p3FieldMessageType
		}
		// If the field package different, add the import to p3doc.
		if field.TypeInfo.ReprType.Package.GoPkgPath != pkgPath {
			if p3FieldType.GetPackage() != "" {
				importPath := p3c.GetP3ImportPath(p3FieldType)
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
func (p3c *P3Context) GenerateProto3SchemaForTypes(pkg *amino.Package, rtz ...reflect.Type) (p3doc P3Doc, err error) {

	if pkg.P3PkgName == "" {
		err = errors.New("cannot generate schema in the root package \"\".")
		return
	}

	// Set the package.
	p3doc.Package = pkg.P3PkgName
	p3doc.GoPackage = pkg.P3GoPkgPath

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
func (p3c *P3Context) WriteProto3SchemaForTypes(filename string, pkg *amino.Package, rtz ...reflect.Type) (err error) {
	fmt.Printf("writing proto3 schema to %v for package %v\n", filename, pkg)
	p3doc, err := p3c.GenerateProto3SchemaForTypes(pkg, rtz...)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filename, []byte(p3doc.Print()), 0644)
	return err
}

// If rt is a struct, the returned proto3 type is a P3MessageType.
// `rt` should be the representation type in case IsAminoMarshaler.
func (p3c *P3Context) reflectTypeToP3Type(rt reflect.Type) (p3type P3Type, repeated bool) {
	// dereference type, in case pointer.
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		if rt.Kind() == reflect.Ptr {
			panic("nested pointers not supported")
		}
	}

	switch rt.Kind() {
	case reflect.Interface:
		return P3AnyType, false
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
		// Look up the p3pkg type from the go path, using P3Context.
		pkg := p3c.GetPackage(rt.PkgPath())
		return NewP3MessageType(pkg.P3PkgName, rt.Name()), false
	default:
		panic("unexpected rt kind")
	}

}

// Writes in the same directory as the origin package.
func WriteProto3Schema(pkg *amino.Package) {
	p3c := NewP3Context()
	p3c.RegisterPackage(pkg)
	p3c.ValidateBasic()
	filename := path.Join(pkg.DirName, "types.proto")
	err := p3c.WriteProto3SchemaForTypes(filename, pkg, pkg.Types...)
	if err != nil {
		panic(err)
	}
}

// Symlinks .proto files from pkg info to dirname, keeping the go path
// structure as expected, <dirName>/path/to/gopkg/types.proto.
// If Pkg.DirName is empty, the package is considered "well known", and
// the mapping is not made.
func MakeProtoFolder(pkg *amino.Package, dirName string) {
	fmt.Printf("making proto3 schema folder for package %v\n", pkg)
	p3c := NewP3Context()
	p3c.RegisterPackage(pkg)

	// Populate mapping.
	// p3 import path -> p3 import file (abs path).
	// e.g. "github.com/.../types.proto" ->
	// "/gopath/pkg/mod/.../types.proto"
	var p3imports = map[string]string{}
	for _, dpkg := range p3c.GetAllPackages() {
		if dpkg.P3SchemaFile == "" {
			// Skip well known packages like google.protobuf.Any
			continue
		}
		p3path := dpkg.P3ImportPath
		if p3path == "" {
			panic("P3ImportPath cannot be empty")
		}
		p3file := dpkg.P3SchemaFile
		p3imports[p3path] = p3file
	}

	// Check validity.
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		panic(fmt.Sprintf("directory %v does not exist", dirName))
	}

	// Make symlinks.
	for p3path, p3file := range p3imports {
		loc := path.Join(dirName, p3path)
		locdir := path.Dir(loc)
		// Ensure that paths exist.
		if _, err := os.Stat(locdir); os.IsNotExist(err) {
			err = os.MkdirAll(locdir, os.ModePerm)
			if err != nil {
				panic(err)
			}
		}
		// Delete existing symlink.
		if _, err := os.Stat(loc); !os.IsNotExist(err) {
			err := os.Remove(loc)
			if err != nil {
				panic(err)
			}
		}
		// Write symlink.
		err := os.Symlink(p3file, loc)
		if err != nil {
			panic(err)
		}
	}
}

// Uses pkg.P3GoPkgPath to determine where the compiled file goes.  If
// pkg.P3GoPkgPath is a subpath of pkg.GoPkgPath, then it will be
// written in the relevant subpath in pkg.DirName.
// `protosDir`: folder where .proto files for all dependencies live.
func RunProtoc(pkg *amino.Package, protosDir string) {
	if !strings.HasSuffix(pkg.P3SchemaFile, ".proto") {
		panic(fmt.Sprintf("expected P3Importfile to have .proto suffix, got %v", pkg.P3SchemaFile))
	}
	inDir := filepath.Dir(pkg.P3SchemaFile)
	inFile := filepath.Base(pkg.P3SchemaFile)
	outDir := path.Join(inDir, "pb")
	outFile := inFile[:len(inFile)-6] + ".pb.go"
	// Ensure that paths exist.
	if _, err := os.Stat(outDir); os.IsNotExist(err) {
		err = os.MkdirAll(outDir, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
	// First generate output to a temp dir.
	tempDir, err := ioutil.TempDir("", "amino-genproto")
	if err != nil {
		return
	}
	// Run protoc
	cmd := exec.Command("protoc", "-I="+inDir, "-I="+protosDir, "--go_out="+tempDir, pkg.P3SchemaFile)
	fmt.Println("running protoc: ", cmd.String())
	cmd.Stdin = nil
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err = cmd.Run()
	if err != nil {
		panic(err)
	}

	// Copy file from tempDir to outDir.
	copyFile(
		path.Join(tempDir, pkg.P3GoPkgPath, outFile),
		path.Join(outDir, outFile),
	)
}

func copyFile(src string, dst string) {
	// Read all content of src to data
	data, err := ioutil.ReadFile(src)
	if err != nil {
		panic(err)
	}
	// Write data to dst
	err = ioutil.WriteFile(dst, data, 0644)
	if err != nil {
		panic(err)
	}
}
