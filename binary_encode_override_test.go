package amino_test

import (
	"github.com/stretchr/testify/assert"
	amino "github.com/tendermint/go-amino"
	"testing"
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

func TestMarshalBinaryOverride(t *testing.T) {
	var cdc = amino.NewCodec()
	cdc.RegisterConcrete(&Thing{}, "amino/thing", nil)

	thing1 := Thing{Name: "a"}

	bz, err := cdc.MarshalBinaryBare(thing1)
	assert.Nil(t, err)

	var thing2 Thing
	err = cdc.UnmarshalBinaryBare(bz, &thing2)
	assert.Nil(t, err)
	assert.Equal(t, thing1, thing2)
}
