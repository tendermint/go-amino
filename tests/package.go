package tests

import (
	"github.com/tendermint/go-amino/pkg"
)

// Creates one much like amino.RegisterPackage, but without registration.
// This is needed due to circular dependency issues for dependencies of Amino.
// Another reason to strive for many independent modules.
var Package = pkg.NewPackage(
	"github.com/tendermint/go-amino/tests",
	"tests",
	pkg.GetCallersDirName(),
).WithDependencies().WithTypes(
	EmptyStruct{},
	PrimitivesStruct{},
	ShortArraysStruct{},
	ArraysStruct{},
	SlicesStruct{},
	PointersStruct{},
	PointerSlicesStruct{},
	//NestedPointersStruct{},
	ComplexSt{},
	EmbeddedSt1{},
	EmbeddedSt2{},
	EmbeddedSt3{},
	EmbeddedSt4{},
	EmbeddedSt5{},
	IntDef(0),
	IntAr{},
	IntSl(nil),
	ByteAr{},
	ByteSl(nil),
	PrimitivesStructDef{},
	PrimitivesStructSl(nil),
	PrimitivesStructAr{},
	Concrete1{},
	Concrete2{},
	ConcreteTypeDef{},
	ConcreteWrappedBytes{},
	&InterfaceFieldsStruct{},
)
