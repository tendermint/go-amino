package wire_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/go-wire"
)

type SimpleStruct struct {
	String string
	Bytes  []byte
	Time   time.Time
}

func newSimpleStruct() SimpleStruct {
	s := SimpleStruct{
		String: "hello",
		Bytes:  []byte("goodbye"),
		Time:   time.Now().UTC().Truncate(time.Millisecond), // strip monotonic and timezone.
	}
	return s
}

func TestMarshalUnmarshalBinaryPointer0(t *testing.T) {

	var s = newSimpleStruct()
	cdc := wire.NewCodec()
	b, err := cdc.MarshalBinary(s) // no indirection
	assert.Nil(t, err)

	var s2 SimpleStruct
	err = cdc.UnmarshalBinary(b, &s2) // no indirection
	assert.Nil(t, err)
	assert.Equal(t, s, s2)

}

func TestMarshalUnmarshalBinaryPointer1(t *testing.T) {

	var s = newSimpleStruct()
	cdc := wire.NewCodec()
	b, err := cdc.MarshalBinary(&s) // extra indirection
	assert.Nil(t, err)

	var s2 SimpleStruct
	err = cdc.UnmarshalBinary(b, &s2) // no indirection
	assert.Nil(t, err)
	assert.Equal(t, s, s2)

}

func TestMarshalUnmarshalBinaryPointer2(t *testing.T) {

	var s = newSimpleStruct()
	var ptr = &s
	cdc := wire.NewCodec()
	b, err := cdc.MarshalBinary(&ptr) // double extra indirection
	assert.Nil(t, err)

	var s2 SimpleStruct
	err = cdc.UnmarshalBinary(b, &s2) // no indirection
	assert.Nil(t, err)
	assert.Equal(t, s, s2)

}

func TestMarshalUnmarshalBinaryPointer3(t *testing.T) {

	var s = newSimpleStruct()
	cdc := wire.NewCodec()
	b, err := cdc.MarshalBinary(s) // no indirection
	assert.Nil(t, err)

	var s2 *SimpleStruct
	err = cdc.UnmarshalBinary(b, &s2) // extra indirection
	assert.NotNil(t, err, "unmarshal should not accept extra indirection")
	// assert.Equal(t, s, s2)

}

func TestMarshalUnmarshalBinaryPointer4(t *testing.T) {

	var s = newSimpleStruct()
	var ptr = &s
	cdc := wire.NewCodec()
	b, err := cdc.MarshalBinary(&ptr) // extra indirection
	assert.Nil(t, err)

	var s2 *SimpleStruct
	err = cdc.UnmarshalBinary(b, &s2) // extra indirection
	assert.NotNil(t, err, "unmarshal should not accept extra indirection")
	// assert.Equal(t, s, s2)

}
