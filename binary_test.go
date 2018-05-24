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

func TestNewFieldBackwardsCompatibility(t *testing.T) {
	type V1 struct {
		String  string
		String2 string
	}

	type V2 struct {
		String  string
		String2 string
		// new field in V2:
		Int int
	}

	type SomeStruct struct {
		Sth int
	}

	type V3 struct {
		String string `json:"string"`
		// different from V1 starting here:
		Int  int
		Some SomeStruct
	}

	cdc := amino.NewCodec()

	v2 := V2{String: "hi", String2: "cosmos", Int: 4}
	bz, err := cdc.MarshalBinaryBare(v2)
	assert.Nil(t, err, "unexpected error while encoding V2: %v", err)

	var v1 V1
	err = cdc.UnmarshalBinaryBare(bz, &v1)
	assert.Nil(t, err, "unexpected error %v", err)
	assert.Equal(t, v1, V1{"hi", "cosmos"},
		"backwards compatibility failed: didn't yield expected result ...")

	v3 := V3{String: "tender", Int: 2014, Some: SomeStruct{Sth: 84}}
	bz2, err := cdc.MarshalBinaryBare(v3)
	assert.Nil(t, err, "unexpected error")

	err = cdc.UnmarshalBinaryBare(bz2, &v1)
	// this might change later but we include this case to document the current behaviour:
	assert.NotNil(t, err, "expected an error here because of changed order of fields")

	// we still expect that decoding worked to some extend (until above error occurred):
	assert.Equal(t, v1, V1{"tender", "cosmos"})
}

func TestAppendedSomeBytesAfterStructTerm(t *testing.T) {
	// like above test but we add some bytes and expect an error:
	type V1 struct {
		String  string
		String2 string
	}

	type V2 struct {
		String  string
		String2 string
		// new field in V2:
		Int int
	}

	cdc := amino.NewCodec()

	v2 := V2{String: "to the", String2: "cosmos", Int: 4}
	bz, err := cdc.MarshalBinaryBare(v2)
	assert.Nil(t, err, "unexpected error while encoding V2: %v", err)

	bzOrig := bz
	// just add some meaningless bytes:
	bz = append(bz, []byte{0xde, 0xad, 0xbe, 0xef}...)
	var v1 V1
	err = cdc.UnmarshalBinaryBare(bz, &v1)
	// expected: unmarshal didn't read all bytes. Expected to read 23, only read 19
	assert.NotNil(t, err, "expected error %v", err)

	assert.Equal(t, v1, V1{"to the", "cosmos"},
		"backwards compatibility failed: didn't yield expected result ...")

	// append a structTerm:
	bz = append(bzOrig, byte(amino.Typ3_StructTerm))

	err = cdc.UnmarshalBinaryBare(bz, &v1)
	// expected: unmarshal didn't read all bytes. Expected to read 20, only read 19
	assert.NotNil(t, err, "expected error %v", err)

	// append a struct start at the end (we should break before and not finish reading here):
	bz = append(bzOrig, byte(amino.Typ3_Struct))

	err = cdc.UnmarshalBinaryBare(bz, &v1)
	// expected: unmarshal didn't read all bytes. Expected to read 20, only read 19
	assert.NotNil(t, err, "expected error %v", err)
}
