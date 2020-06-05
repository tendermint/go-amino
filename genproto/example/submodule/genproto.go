package submodule

import (
	"github.com/tendermint/go-amino/genproto"
	"github.com/tendermint/go-amino/genproto/example/submodule2"
)

var PackageInfo = genproto.NewPackageInfo(
	"github.com/tendermint/go-amino/genproto/example/submodule",
	"submodule",
).WithDependencies(
	submodule2.PackageInfo,
).WithStructs(
	StructSM{},
)
