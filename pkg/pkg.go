package pkg

import (
	"fmt"
	"path"
	"reflect"
	"regexp"
	"runtime"
)

type Package struct {
	GoPkg        string
	Dirname      string
	P3Pkg        string
	P3Import     string
	Dependencies []*Package
	Types        []reflect.Type
}

// Like amino.RegisterPackage (which is probably what you're looking for unless
// you are developkgng on go-amino dependencies), but without global amino
// registration.
// Panics if invalid arguments are given, such as slashes in p3pkg, invalid go
// pkg paths, or a relative dirname.
func NewPackage(gopkg string, p3pkg string, dirname string) *Package {
	assertValidGoPkg(gopkg)
	assertValidP3Pkg(p3pkg)
	assertValidDirname(dirname)
	return &Package{
		GoPkg:        gopkg,
		Dirname:      dirname,
		P3Pkg:        p3pkg,
		P3Import:     "",
		Dependencies: nil,
		Types:        nil,
	}
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
		if objDerefType.PkgPath() != pkg.GoPkg {
			panic(fmt.Sprintf("unexpected package for %v, expected %v got %v for obj %v obj type %v", objDerefType, pkg.GoPkg, objDerefType.PkgPath(), obj, objType))
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

// These files will get imported instead of the default "types.proto" if this package is a dependency.
func (pkg *Package) WithP3Import(p3import string) *Package {
	pkg.P3Import = p3import
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
		rt = derefType(rt)
		if rt.Name() == name {
			return true
		}
	}
	return false
}

func (pkg *Package) HasFullName(name string) (exists bool) {
	for _, rt := range pkg.Types {
		rt = derefType(rt)
		if pkg.FullNameForType(rt) == name {
			return true
		}
	}
	return false
}

// panics of rt was not registered.
func (pkg *Package) FullNameForType(rt reflect.Type) string {
	rt = derefType(rt)
	exists, err := pkg.HasType(rt)
	if !exists {
		panic(err)
	}
	return fmt.Sprintf("%v.%v", pkg.P3Pkg, rt.Name())
}

// panics of rt (or a pointer to it) was not registered.
func (pkg *Package) TypeURLForType(rt reflect.Type) string {
	name := pkg.FullNameForType(rt)
	return "/" + name
}

//----------------------------------------

// Utility for whoever is making a NewPackage manually.
func GetCallersDirname() string {
	var dirname = "" // derive from caller.
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		panic("could not get caller to derive caller's package directory")
	}
	dirname = path.Dir(filename)
	if filename == "" || dirname == "" {
		panic("could not derive caller's package directory")
	}
	return dirname
}

var (
	RE_DOMAIN     = `[[:alnum:]-_]+[[:alnum:]-_.]+\.[a-zA-Z]{2,4}`
	RE_GOPKG_PART = `[[:alpha:]-_]+`
	RE_GOPKG      = fmt.Sprintf(`(?:%v|%v)(?:/%v)*`, RE_DOMAIN, RE_GOPKG_PART, RE_GOPKG_PART)
	RE_P3PKG_PART = `[[:alpha:]_]+`
	RE_P3PKG      = fmt.Sprintf(`%v(?:\.:%v)*`, RE_P3PKG_PART, RE_P3PKG_PART)
)

func assertValidGoPkg(gopkg string) {
	matched, err := regexp.Match(RE_GOPKG, []byte(gopkg))
	if err != nil {
		panic(err)
	}
	if !matched {
		panic(fmt.Sprintf("not a valid go package path: %v", gopkg))
	}
}

func assertValidP3Pkg(p3pkg string) {
	matched, err := regexp.Match(RE_P3PKG, []byte(p3pkg))
	if err != nil {
		panic(err)
	}
	if !matched {
		panic(fmt.Sprintf("not a valid proto3 package path: %v", p3pkg))
	}
}

// The dirname is only used to tell code generation tools where to put them.  I
// suppose the default could be empty for convenience, as long as it isn't a
// relative path that tries to access parent directories.
func assertValidDirname(dirname string) {
	if dirname == "" {
		// Default dirname of empty is allowed, for convenience.
		// Any generated files would be written in the current directory.
		// Dirname should not be set to "." or "./".
		return
	}
	if !path.IsAbs(dirname) {
		panic(fmt.Sprintf("dirname if present should be absolute, but got %v", dirname))
	}
	if path.Dir(dirname+"/dummy") != dirname {
		panic(fmt.Sprintf("dirname not canonical. got %v, expected %v", dirname, path.Dir(dirname+"/dummy")))
	}
}

func derefType(rt reflect.Type) (drt reflect.Type) {
	drt = rt
	for drt.Kind() == reflect.Ptr {
		drt = drt.Elem()
	}
	return
}
