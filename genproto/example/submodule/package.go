package submodule

import (
	"github.com/tendermint/go-amino"
	"github.com/tendermint/go-amino/genproto/example/submodule2"
)

var PackageInfo = amino.RegisterPackageInfo(
	"github.com/tendermint/go-amino/genproto/example/submodule",
	"submodule",
).WithDependencies(
	submodule2.PackageInfo,
).WithTypes(
	StructSM{},
)
