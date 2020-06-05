package main

import (
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
	// Defined in genproto.go
	genproto.WriteProto3Schemas(
		PackageInfo,
		submodule.PackageInfo,
		submodule2.PackageInfo,
	)
}
