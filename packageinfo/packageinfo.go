package packageinfo

import (
	"fmt"
	"path"
	"reflect"
	"runtime"
	"strings"
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
func NewPackageInfo(gopkg string, p3pkg string, dirname string) *PackageInfo {
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
			panic(fmt.Sprintf("unexpected package for %v, expected %v got %v", objDerefType, pi.GoPkg, objDerefType.PkgPath()))
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

func (pi *PackageInfo) ValidateBasic() {
	if pi.GoPkg == "" {
		panic("go pkg can't be empty")
	}
	if pi.Dirname == "" {
		panic("dirname can't be empty")
	}
	if pi.P3Pkg == "" {
		panic("p3 pkg can't be empty")
	}
	if strings.Contains(pi.P3Pkg, "/") {
		panic("p3 pkg can't contain any slashes")
	} // TODO use REGEX
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
	return path.Join(pi.P3Pkg, rt.Name())
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
