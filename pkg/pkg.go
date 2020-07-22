package pkg

import (
	"fmt"
	"path"
	"reflect"
	"regexp"
	"runtime"
	"strings"
)

type Package struct {
	// General info
	GoPkgPath    string
	GoPkgName    string
	DirName      string
	Dependencies []*Package
	Types        []reflect.Type

	// Proto3 info
	P3GoPkgPath  string
	P3PkgName    string
	P3ImportPath string
	P3SchemaFile string
}

// Like amino.RegisterPackage (which is probably what you're looking for
// unless you are developking on go-amino dependencies), but without
// global amino registration.
//
// P3GoPkgPath (the import path for go files generated from protoc) are
// by default set to "<GoPkgPath>/pb", but can be overridden by
// WithP3GoPkgPath().  You may want to override this for main package,
// for the subdirectory "pb" doesn't produce a "main/pb" package.  See
// ./proto/example/pacakge.go for such usage.
//
// (This field is needed for improving the performance of
// encoding and decoding by using protoc generated go code, but is
// slated to be replaced by native Go generation.)
//
// GoPkgName is by default derived from gopkgPath, but can also be
// overridden with WithGoPkgName().
//
// P3ImportPath is what is imported in the p3 import spec.  Generally
// this is GoPkgPath + "/types.proto", but packages can override this
// behavior, and sometimes (e.g. for google.protobuf.Any) it is
// necessary to provide fixed values.  This is not the absolute path to
// the actual file.  That is P3SchemaFile.
//
// Panics if invalid arguments are given, such as slashes in p3pkgName,
// invalid go pkg paths, or a relative dirName.
func NewPackage(gopkgPath string, p3pkgName string, dirName string) *Package {
	if gopkgPath == "" && (p3pkgName != "" || dirName != "") {
		panic("if goPkgPath is empty, p3PkgName and dirName must also be")
	}
	if gopkgPath != "" {
		assertValidGoPkgPath(gopkgPath)
	}
	if p3pkgName != "" {
		assertValidP3PkgName(p3pkgName)
	}
	assertValidDirName(dirName)
	if gopkgPath != "" {
		pkg := &Package{
			GoPkgPath:    gopkgPath,
			GoPkgName:    DefaultPkgName(gopkgPath),
			DirName:      dirName,
			Dependencies: nil,
			Types:        nil,
			P3GoPkgPath:  path.Join(gopkgPath, "pb"),
			P3PkgName:    p3pkgName,
			P3ImportPath: path.Join(gopkgPath, "types.proto"),
			P3SchemaFile: path.Join(dirName, "types.proto"),
		}
		return pkg
	} else {
		pkg := &Package{
			Dependencies: nil,
			Types:        nil,
		}
		return pkg
	}
}

func (pkg *Package) String() string {
	return fmt.Sprintf("pkg.Pkg(%v@%v)", pkg.GoPkgPath, pkg.DirName)
}

func (pkg *Package) WithP3GoPkgPath(p3gopkg string) *Package {
	pkg.P3GoPkgPath = p3gopkg
	return pkg
}

func (pkg *Package) WithP3GoPkgName(name string) *Package {
	pkg.GoPkgName = name
	return pkg
}

func (pkg *Package) WithDependencies(deps ...*Package) *Package {
	pkg.Dependencies = append(pkg.Dependencies, deps...)
	return pkg
}

func (pkg *Package) WithTypes(objs ...interface{}) *Package {
	for _, obj := range objs {
		objType := reflect.TypeOf(obj)
		objDerefType := objType
		for objDerefType.Kind() == reflect.Ptr {
			objDerefType = objDerefType.Elem()
		}
		if objDerefType.PkgPath() != pkg.GoPkgPath {
			panic(fmt.Sprintf("unexpected package for %v, expected %v got %v for obj %v obj type %v", objDerefType, pkg.GoPkgPath, objDerefType.PkgPath(), obj, objType))
		}
		exists, err := pkg.HasType(objType)
		if exists {
			panic(err)
		}
		// NOTE: keep pointer info, as preference for deserialization.
		// This is how amino.Codec.RegisterTypeFrom() knows.
		pkg.Types = append(pkg.Types, objType)
	}
	return pkg
}

// This path will get imported instead of the default "types.proto"
// if this package is a dependency.  This is not the filesystem path,
// but the path imported within the proto schema file.  The filesystem
// path is .P3SchemaFile.
func (pkg *Package) WithP3ImportPath(path string) *Package {
	pkg.P3ImportPath = path
	return pkg
}

// This file will get imported instead of the default "types.proto" if this package is a dependency.
func (pkg *Package) WithP3SchemaFile(file string) *Package {
	pkg.P3SchemaFile = file
	return pkg
}

// err is always non-nil and includes some generic message.
// (since the caller may either expect the type in the package or not).
func (pkg *Package) HasType(rt reflect.Type) (exists bool, err error) {
	for _, rt2 := range pkg.Types {
		if rt == rt2 {
			return true, fmt.Errorf("type %v already registered with package", rt)
		}
		if rt.Kind() == reflect.Ptr && rt.Elem() == rt2 {
			return true, fmt.Errorf("non-pointer receiver registered in package but got %v", rt)
		}
		if rt2.Kind() == reflect.Ptr && rt == rt2.Elem() {
			return true, fmt.Errorf("pointer receiver registered in package but got %v", rt)
		}
	}
	return false, fmt.Errorf("type %v not registered with package", rt)
}

func (pkg *Package) HasName(name string) (exists bool) {
	for _, rt := range pkg.Types {
		if rt.Kind() == reflect.Ptr {
			rt = rt.Elem()
		}
		if rt.Name() == name {
			return true
		}
	}
	return false
}

func (pkg *Package) HasFullName(name string) (exists bool) {
	for _, rt := range pkg.Types {
		if rt.Kind() == reflect.Ptr {
			rt = rt.Elem()
		}
		if pkg.FullNameForType(rt) == name {
			return true
		}
	}
	return false
}

// panics of rt was not registered.
func (pkg *Package) FullNameForType(rt reflect.Type) string {
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	exists, err := pkg.HasType(rt)
	if !exists {
		panic(err)
	}
	return fmt.Sprintf("%v.%v", pkg.P3PkgName, rt.Name())
}

// panics of rt (or a pointer to it) was not registered.
func (pkg *Package) TypeURLForType(rt reflect.Type) string {
	name := pkg.FullNameForType(rt)
	return "/" + name
}

// Finds a dependency package from the gopkg.  Well known packages are
// not known here, so some dependencies may not show up, such as for
// google.protobuf.Any for any interface fields.
// For that, use a P3Context.GetPackage().
func (pkg *Package) GetDependency(gopkg string) (*Package, error) {
	all := pkg.CrawlPackages(nil)
	for _, pkg := range all {
		if pkg.GoPkgPath == gopkg {
			return pkg, nil
		}
	}
	return nil, fmt.Errorf("Go package not declared a (in)direct dependency of %v",
		pkg.GoPkgPath)
}

// For a given package info, crawl and discover all package infos.
func (pkg *Package) CrawlPackages(seen map[*Package]struct{}) (res []*Package) {
	if seen == nil {
		seen = map[*Package]struct{}{}
	}
	var crawl func(pkg *Package)
	crawl = func(pkg *Package) {
		seen[pkg] = struct{}{}
		for _, dependency := range pkg.Dependencies {
			if _, ok := seen[dependency]; ok {
				continue
			}
			crawl(dependency)
		}
	}
	crawl(pkg)
	for pkg, _ := range seen {
		res = append(res, pkg)
	}
	return res
}

//----------------------------------------

// Utility for whoever is making a NewPackage manually.
func GetCallersDirName() string {
	var dirName = "" // derive from caller.
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		panic("could not get caller to derive caller's package directory")
	}
	dirName = path.Dir(filename)
	if filename == "" || dirName == "" {
		panic("could not derive caller's package directory")
	}
	return dirName
}

var (
	RE_DOMAIN     = `[[:alnum:]-_]+[[:alnum:]-_.]+\.[a-zA-Z]{2,4}`
	RE_GOPKG_PART = `[[:alpha:]-_]+`
	RE_GOPKG      = fmt.Sprintf(`(?:%v|%v)(?:/%v)*`, RE_DOMAIN, RE_GOPKG_PART, RE_GOPKG_PART)
	RE_P3PKG_PART = `[[:alpha:]_]+`
	RE_P3PKG      = fmt.Sprintf(`%v(?:\.:%v)*`, RE_P3PKG_PART, RE_P3PKG_PART)
)

func assertValidGoPkgPath(gopkgPath string) {
	matched, err := regexp.Match(RE_GOPKG, []byte(gopkgPath))
	if err != nil {
		panic(err)
	}
	if !matched {
		panic(fmt.Sprintf("not a valid go package path: %v", gopkgPath))
	}
}

func assertValidP3PkgName(p3pkgName string) {
	matched, err := regexp.Match(RE_P3PKG, []byte(p3pkgName))
	if err != nil {
		panic(err)
	}
	if !matched {
		panic(fmt.Sprintf("not a valid proto3 package path: %v", p3pkgName))
	}
}

// The dirName is only used to tell code generation tools where to put them.  I
// suppose the default could be empty for convenience, as long as it isn't a
// relative path that tries to access parent directories.
func assertValidDirName(dirName string) {
	if dirName == "" {
		// Default dirName of empty is allowed, for convenience.
		// Any generated files would be written in the current directory.
		// DirName should not be set to "." or "./".
		return
	}
	if !path.IsAbs(dirName) {
		panic(fmt.Sprintf("dirName if present should be absolute, but got %v", dirName))
	}
	if path.Dir(dirName+"/dummy") != dirName {
		panic(fmt.Sprintf("dirName not canonical. got %v, expected %v", dirName, path.Dir(dirName+"/dummy")))
	}
}

func DefaultPkgName(gopkgPath string) (name string) {
	parts := strings.Split(gopkgPath, "/")
	last := parts[len(parts)-1]
	parts = strings.Split(last, "-")
	name = parts[len(parts)-1]
	name = strings.ToLower(name)
	return name
}
