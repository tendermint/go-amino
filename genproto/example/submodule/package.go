package submodule

import (
	"github.com/tendermint/go-amino"
	"github.com/tendermint/go-amino/genproto/example/submodule2"
)

var Package = amino.RegisterPackage(
	amino.NewPackage(
		"github.com/tendermint/go-amino/genproto/example/submodule",
		"submodule",
		amino.GetCallersDirname(),
	).WithDependencies(
		submodule2.Package,
	).WithTypes(
		StructSM{},
	),
)
