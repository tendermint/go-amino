package submodule2

import (
	"github.com/tendermint/go-amino"
)

var PackageInfo = amino.RegisterPackageInfo(
	"github.com/tendermint/go-amino/genproto/example/submodule2",
	"submodule2",
).WithTypes(
	StructSM2{},
)
