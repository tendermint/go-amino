package main

import (
	"fmt"
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
}
