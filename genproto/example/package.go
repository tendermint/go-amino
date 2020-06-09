package main

import (
	"github.com/tendermint/go-amino"
	"github.com/tendermint/go-amino/genproto/example/submodule"
)

var PackageInfo = amino.RegisterPackageInfo(
	"main",
	"main",
).WithDependencies(
	submodule.PackageInfo,
).WithTypes(
	StructA{},
	StructB{},
)
