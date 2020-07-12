package main

import (
	"github.com/tendermint/go-amino/genproto"
	"github.com/tendermint/go-amino/tests"
)

func main() {
	pkg := tests.Package
	genproto.WriteProto3Schema(pkg)
	genproto.MakeProtoFolder(pkg, "proto")
	genproto.RunProtoc(pkg, "proto")
	genproto.WriteProtoBindings(pkg)
}
