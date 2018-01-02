package wire_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	wire "github.com/tendermint/go-wire"
)

func TestMarshalGlobal(t *testing.T) {
	s := SimpleStruct{
		String: "hello",
		Bytes:  []byte("goodbye"),
		Time:   time.Now().Truncate(time.Millisecond),
	}

	b, err := wire.MarshalBinary(s)
	assert.Nil(t, err)

	var s2 SimpleStruct
	err = wire.UnmarshalBinary(b, &s2)
	assert.Nil(t, err)
	assert.Equal(t, s, s2)

}
