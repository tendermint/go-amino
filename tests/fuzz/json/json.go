package fuzz_json

import (
	"github.com/tendermint/go-wire"
	"github.com/tendermint/go-wire/tests"
)

//-------------------------------------
// Non-interface go-fuzz tests
// See https://github.com/dvyukov/go-fuzz
// (Test that deserialize never panics)

func Fuzz(data []byte) int {
	cdc := wire.NewCodec()
	cst := tests.ComplexSt{}
	err := cdc.UnmarshalJSON(data, &cst)
	if err != nil {
		return 0
	}
	return 1
}
