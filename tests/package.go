package tests

import (
	"github.com/tendermint/go-amino/pkg"
)

// Creates one much like amino.RegisterPackage, but without registration.
// This is needed due to circular dependency issues for dependencies of Amino.
// Another reason to strive for many independent modules.
// NOTE: Register new repr types here as well.
// NOTE: This package registration is independent of test registration.
// See tests/common.go StructTypes etc to add to tests.
var Package = pkg.NewPackage(
	"github.com/tendermint/go-amino/tests",
	"tests",
	pkg.GetCallersDirName(),
).WithDependencies().WithTypes(
	EmptyStruct{},
	PrimitivesStruct{},
	ShortArraysStruct{},
	ArraysStruct{},
	ArraysArraysStruct{},
	SlicesStruct{},
	SlicesSlicesStruct{},
	PointersStruct{},
	PointerSlicesStruct{},
	//NestedPointersStruct{},
	ComplexSt{},
	EmbeddedSt1{},
	EmbeddedSt2{},
	EmbeddedSt3{},
	EmbeddedSt4{},
	EmbeddedSt5{},
	AminoMarshalerStruct1{},
	ReprStruct1{},
	AminoMarshalerStruct2{},
	ReprElem2{},
	AminoMarshalerStruct3{},
	AminoMarshalerInt4(0),
	AminoMarshalerInt5(0),
	AminoMarshalerStruct6{},
	AminoMarshalerStruct7{},
	ReprElem7{},
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
