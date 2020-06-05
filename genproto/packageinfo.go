package genproto

import (
	"path"
	"reflect"
	"runtime"
)

type PackageInfo struct {
	GoPkg        string
	Dirname      string
	P3Pkg        string
	Dependencies []*PackageInfo
	StructTypes  []reflect.Type
}

// NOTE: Dirname is derived from the caller, using runtime caller analysis.
// This function must be called from within gopkg, with no function decoration
// or indirection.
// If you must, refactor this method and create a new one this calls, which
// takes shift.
func NewPackageInfo(gopkg string, p3pkg string) *PackageInfo {

	var dirname = "" // derive from caller.
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		panic("could not get caller to derive caller's package directory")
	}
	dirname = path.Dir(filename)
	if filename == "" || dirname == "" {
		panic("could not derive caller's package directory")
	}
	return &PackageInfo{
		GoPkg:        gopkg,
		Dirname:      dirname,
		P3Pkg:        p3pkg,
		Dependencies: nil,
		StructTypes:  nil,
	}
}

func (pi *PackageInfo) WithDependencies(deps ...*PackageInfo) *PackageInfo {
	pi.Dependencies = append(pi.Dependencies, deps...)
	return pi
}

func (pi *PackageInfo) WithStructs(structs ...interface{}) *PackageInfo {
	for _, str := range structs {
		strType := reflect.TypeOf(str)
		if strType.Kind() != reflect.Struct {
			panic("proto structs only in WithStructs()")
		}
		pi.StructTypes = append(pi.StructTypes, strType)
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
}
