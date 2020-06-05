package submodule2

import (
	"github.com/tendermint/go-amino/genproto"
)

var PackageInfo = genproto.NewPackageInfo(
	"github.com/tendermint/go-amino/genproto/example/submodule2",
	"submodule2",
).WithStructs(
	StructSM2{},
)
