package amino_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/go-amino"
)

func TestNilSliceEmptySlice(t *testing.T) {
	var cdc = amino.NewCodec()

	type TestStruct struct {
		A []byte
		B []int
		C [][]byte
		D [][]int
		E []*[]byte
		F []*[]int
	}
	nnb, nni := []byte(nil), []int(nil)
	eeb, eei := []byte{}, []int{}

	a := TestStruct{
		A: nnb,
		B: nni,
		C: [][]byte{nnb},
		D: [][]int{nni},
		E: []*[]byte{nil},
		F: []*[]int{nil},
	}
	b := TestStruct{
		A: eeb,
		B: eei,
		C: [][]byte{eeb},
		D: [][]int{eei},
		E: []*[]byte{&nnb},
		F: []*[]int{&nni},
	}
	c := TestStruct{
		A: eeb,
		B: eei,
		C: [][]byte{eeb},
		D: [][]int{eei},
		E: []*[]byte{&eeb},
		F: []*[]int{&eei},
	}

	abz := cdc.MustMarshalBinary(a)
	bbz := cdc.MustMarshalBinary(b)
	cbz := cdc.MustMarshalBinary(c)

	assert.Equal(t, abz, bbz, "a != b")
	assert.Equal(t, abz, cbz, "a != c")
}
