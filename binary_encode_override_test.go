package amino_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/go-amino"
)

type Thing struct {
	Name string
}

func (thing Thing) MarshalBinary() ([]byte, error) {
	return []byte(thing.Name), nil
}

func (thing *Thing) UnmarshalBinary(bz []byte) error {
	thing.Name = string(bz)
	return nil
}

func TestMarshalBinaryOverrideBare(t *testing.T) {
	var cdc = amino.NewCodec()
	cdc.RegisterConcrete(&Thing{}, "amino/thing", nil)

	thing1 := Thing{Name: "a"}

	bz, err := cdc.MarshalBinaryBare(thing1)
	assert.Nil(t, err)
	assert.Equal(t, bz, []byte{140, 74, 30, 175, 97})

	var thing2 Thing
	err = cdc.UnmarshalBinaryBare(bz, &thing2)
	assert.Nil(t, err)
	assert.Equal(t, thing1, thing2)
}

func TestMarshalBinaryOverrideLengthPrefixed(t *testing.T) {
	var cdc = amino.NewCodec()
	cdc.RegisterConcrete(&Thing{}, "amino/thing", nil)

	thing1 := Thing{Name: "a"}

	bz, err := cdc.MarshalBinaryLengthPrefixed(thing1)
	assert.Nil(t, err)
	assert.Equal(t, bz, []byte{5, 140, 74, 30, 175, 97})

	var thing2 Thing
	err = cdc.UnmarshalBinaryLengthPrefixed(bz, &thing2)
	assert.Nil(t, err)
	assert.Equal(t, thing1, thing2)
}

type Bytes [16]byte

func (bytes Bytes) MarshalBinary() ([]byte, error) {
	bz := make([]byte, 17)
	copy(bz[:1], []byte{16})
	copy(bz[1:], bytes[:])
	return bz, nil
}

func (bytes *Bytes) UnmarshalBinary(bz []byte) error {
	copy(bytes[:], bz[1:])
	return nil
}

func TestMarshalBinaryOverrideBytes(t *testing.T) {
	var cdc = amino.NewCodec()
	cdc.RegisterConcrete(&Bytes{}, "amino/bytes", nil)

	bytes1 := Bytes{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}

	bz, err := cdc.MarshalBinaryBare(bytes1)
	assert.Nil(t, err)
	assert.Equal(t, bz, []byte{207, 109, 94, 111, 16, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})

	var bytes2 Bytes
	err = cdc.UnmarshalBinaryBare(bz, &bytes2)
	assert.Nil(t, err)
	assert.Equal(t, bytes1, bytes2)
}
