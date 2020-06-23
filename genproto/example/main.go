package main

import (
	"github.com/tendermint/go-amino"
	"github.com/tendermint/go-amino/genproto"
	"github.com/tendermint/go-amino/genproto/example/submodule"
	"github.com/tendermint/go-amino/genproto/example/submodule2"
)

// amino
type StructA struct {
	fieldA int
	fieldB int
	FieldC int
	FieldD uint32
}

// amino
type StructB struct {
	fieldA int
	fieldB int
	FieldC int
	FieldD uint32
	FieldE submodule.StructSM
	FieldF StructA
	FieldG interface{}
}

func main() {

	packages := []*amino.Package{
		Package,
		submodule.Package,
		submodule2.Package,
	}

	// Defined in genproto.go.
	// These will generate .proto files next to
	// their .go origins.
	genproto.WriteProto3Schemas(packages...)

	genproto.WriteProtoBindings(packages...)

	/*

		// Before running protoc, gather
		// all dependencies as symlinks into the
		// ./proto directory.
		// NOTE: All dependencies will be symlinked!  This is unlike
		// WriteProto3Schemas, which only writes files for the listed
		// packages.
		genproto.MakeProtoFolder(packages)

		// Then,
		// > protoc -I=./proto --go_out=./pb types.proto
		// would generate .pb.go files in the ./pb folder,
		// with all imports symlinked to from the ./proto
		// folder.
		genproto.CallProtoc(
	*/
}
