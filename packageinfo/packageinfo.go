package packageinfo

import (
	"fmt"
	"path"
	"reflect"
	"regexp"
	"runtime"
)

type PackageInfo struct {
	GoPkg        string
	Dirname      string
	P3Pkg        string
	Dependencies []*PackageInfo
	Types        []reflect.Type
}

// Like amino.RegisterPackage (which is probably what you're looking for unless
// you are developing on go-amino dependencies), but without global amino
// registration.
// Panics if invalid arguments are given, such as slashes in p3pkg, invalid go
// pkg paths, or a relative dirname.
func NewPackageInfo(gopkg string, p3pkg string, dirname string) *PackageInfo {
	assertValidGoPkg(gopkg)
	assertValidP3Pkg(p3pkg)
	assertValidDirname(dirname)
	return &PackageInfo{
		GoPkg:        gopkg,
		Dirname:      dirname,
		P3Pkg:        p3pkg,
		Dependencies: nil,
		Types:        nil,
	}
}

func (pi *PackageInfo) WithDependencies(deps ...*PackageInfo) *PackageInfo {
	pi.Dependencies = append(pi.Dependencies, deps...)
	return pi
}

func (pi *PackageInfo) WithTypes(objs ...interface{}) *PackageInfo {
	for _, obj := range objs {
		objType := reflect.TypeOf(obj)
		objDerefType := objType
		for objDerefType.Kind() == reflect.Ptr {
			objDerefType = objDerefType.Elem()
		}
		if objDerefType.PkgPath() != pi.GoPkg {
			panic(fmt.Sprintf("unexpected package for %v, expected %v got %v for obj %v obj type %v", objDerefType, pi.GoPkg, objDerefType.PkgPath(), obj, objType))
		}
		exists, err := pi.HasType(objType)
		if exists {
			panic(err)
		}
		// NOTE: keep pointer info, as preference for deserialization.
		// This is how amino.Codec.RegisterTypeFrom() knows.
		pi.Types = append(pi.Types, objType)
	}
	return pi
}

// err is always non-nil and includes some generic message.
// (since the caller may either expect the type in the package or not).
func (pi *PackageInfo) HasType(rt reflect.Type) (exists bool, err error) {
	for _, rt2 := range pi.Types {
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

// panics of rt was not registered
func (pi *PackageInfo) NameForType(rt reflect.Type) string {
	exists, err := pi.HasType(rt)
	if !exists {
		panic(err)
	}
	return fmt.Sprintf("%v.%v", pi.P3Pkg, rt.Name())
}

// panics of rt was not registered
func (pi *PackageInfo) TypeURLForType(rt reflect.Type) string {
	name := pi.NameForType(rt)
	return "/" + name
}

//----------------------------------------

// Utility for whoever is making a NewPackageInfo manually.
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
