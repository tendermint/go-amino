package main

import (
	"fmt"
	"reflect"

	"github.com/tendermint/go-amino"
	"github.com/tendermint/go-amino/genproto"
	"github.com/tendermint/go-amino/genproto/example/submodule"
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
}

func main() {
	fmt.Println("dontcare")

	// Testing that amino structs defined in main also work.
	p3c := genproto.NewP3Context()
	// p3c.RegisterPackageMapping("github.com/tendermint/go-amino/genproto/example/submodule", "example.submodule", []string{"example/submodule/types.go"})
	p3c.RegisterPackageMapping("github.com/tendermint/go-amino/genproto/example/submodule", "example.submodule", nil)
	cdc := amino.NewCodec()
	err := p3c.WriteProto3Schema("types.proto", "main", cdc,
		reflect.TypeOf(StructA{}),
		reflect.TypeOf(StructB{}))
	if err != nil {
		panic(err)
	}
}
