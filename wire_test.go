package wire_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/go-wire"
)

func TestMarshalBinary(t *testing.T) {
	var cdc = wire.NewCodec()

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

	b, err := cdc.MarshalBinary(s)
	assert.Nil(t, err)
	t.Logf("MarshalBinary(s) -> %X", b)

	var s2 SimpleStruct
	err = cdc.UnmarshalBinary(b, &s2)
	assert.Nil(t, err)
	assert.Equal(t, s, s2)
}

func TestUnmarshalBinaryReader(t *testing.T) {
	var cdc = wire.NewCodec()

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

	b, err := cdc.MarshalBinary(s)
	assert.Nil(t, err)
	t.Logf("MarshalBinary(s) -> %X", b)

	var s2 SimpleStruct
	err = cdc.UnmarshalBinaryReader(bytes.NewBuffer(b), &s2, 0)
	assert.Nil(t, err)

	assert.Equal(t, s, s2)
}

func TestUnmarshalBinaryReaderTooLong(t *testing.T) {
	var cdc = wire.NewCodec()

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

	b, err := cdc.MarshalBinary(s)
	assert.Nil(t, err)
	t.Logf("MarshalBinary(s) -> %X", b)

	var s2 SimpleStruct
	err = cdc.UnmarshalBinaryReader(bytes.NewBuffer(b), &s2, 1) // 1 byte limit is ridiculous.
	assert.NotNil(t, err)
}
