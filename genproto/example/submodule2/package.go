package submodule2

import (
	"github.com/tendermint/go-amino"
)

var Package = amino.RegisterPackage(
	amino.NewPackage(
		"github.com/tendermint/go-amino/genproto/example/submodule2",
		"submodule2",
		amino.GetCallersDirname(),
	).WithTypes(
		StructSM2{},
	),
)
