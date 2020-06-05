package main

import (
	"github.com/tendermint/go-amino/genproto"
	"github.com/tendermint/go-amino/genproto/example/submodule"
)

var PackageInfo = genproto.NewPackageInfo(
	"main",
	"main",
).WithDependencies(
	submodule.PackageInfo,
).WithStructs(
	StructA{},
	StructB{},
)
