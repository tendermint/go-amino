package main

import (
	"github.com/tendermint/go-amino"
	"github.com/tendermint/go-amino/genproto/example/submodule"
)

var Package = amino.RegisterPackage(
	amino.NewPackage(
		"main",
		"main",
		amino.GetCallersDirname(),
	).WithGoP3PkgPath(
		"github.com/tendermint/go-amino/genproto/example/pb",
	).WithDependencies(
		submodule.Package,
	).WithTypes(
		StructA{},
		StructB{},
	),
)
