package wire_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/go-wire"
)

func TestMarshalGlobal(t *testing.T) {

	type SimpleStruct struct {
		String string
		Bytes  []byte
		Time   time.Time
	}

	s := SimpleStruct{
		String: "hello",
		Bytes:  []byte("goodbye"),
		Time:   time.Now().UTC().Truncate(time.Millisecond), // strip monotonic and timezone.
	}

	b, err := wire.MarshalBinary(s)
	assert.Nil(t, err)

	var s2 SimpleStruct
	err = wire.UnmarshalBinary(b, &s2)
	assert.Nil(t, err)
	assert.Equal(t, s, s2)

}
