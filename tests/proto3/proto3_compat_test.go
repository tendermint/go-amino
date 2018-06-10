// +build extensive_tests

// only built if manually enforced
package proto3

import (
	"testing"

	"github.com/golang/protobuf/proto"
	pbf "github.com/golang/protobuf/proto/proto3_proto"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/go-amino"
	p3 "github.com/tendermint/go-amino/tests/proto3/proto"
)

// This file checks basic proto3 compatibility by checking encoding of some test-vectors generated by

// TODO(ismail): add a semi-automatic way to test for (more or less full) compatibility to proto3, ideally,
// using their .proto test files: https://github.com/golang/protobuf/tree/master/proto

// List of differences:
// panic: floating point types are unsafe for go-amino

func TestEncodeAminoDecodeProto(t *testing.T) {
	cdc := amino.NewCodec()
	// we have to define our own struct for amino enc because the proto3 test files contains floating types
	type Msg struct {
		Name     string
		Hilarity pbf.Message_Humour
	}
	m := pbf.Message{Name: "Cosmos"}
	ab, err := cdc.MarshalBinaryBare(Msg{Name: "Cosmos"})
	assert.NoError(t, err, "unexpected error")

	pb, err := proto.Marshal(&m)
	assert.NoError(t, err, "unexpected error")
	// This works:
	assert.Equal(t, pb, ab, "encoding doesn't match")

	m = pbf.Message{Name: "Cosmos", Hilarity: pbf.Message_PUNS}
	ab, err = cdc.MarshalBinaryBare(Msg{Name: "Cosmos", Hilarity: pbf.Message_PUNS})
	assert.NoError(t, err, "unexpected error")

	pb, err = proto.Marshal(&m)
	assert.NoError(t, err, "unexpected error")

	// This does not work (same if we drop Name and only have the int32 field):
	//assert.Equal(t, pb, ab, "encoding doesn't match")

	m2 := pbf.Nested{Bunny: "foo", Cute: true}
	ab, err = cdc.MarshalBinaryBare(m2)
	assert.NoError(t, err, "unexpected error")

	pb, err = proto.Marshal(&m2)
	assert.NoError(t, err, "unexpected error")
	assert.Equal(t, pb, ab, "encoding doesn't match")

	// in amino we encode golang int32 as fixed size
	// in protobuf int32 is varint encoded by default
	// so we cant't just use the line below:
	// ab, err = cdc.MarshalBinaryBare(p3.Test32{Foo: 150, Bar: 150})
	// instead:
	type test32 struct {
		Foo int32
		Bar int
	}
	ab, err = cdc.MarshalBinaryBare(test32{Foo: 150, Bar: 150})
	assert.NoError(t, err, "unexpected error")
	pb, err = proto.Marshal(&p3.Test32{Foo: 150, Bar: 150})
	assert.NoError(t, err, "unexpected error")
	assert.Equal(t, pb, ab, "encoding doesn't match")

	// in amino we encode golang int32 as fixed size
	// in protobuf int32 is varint encoded by default
	// so we cant't just encode the below with amino and
	// expect the same outcome:
	//
	// varint := p3.TestInt32Varint{Int32: 150}
	//
	// instead we define a type that will be also be varint encoded in amino:
	type testInt32Varint struct {
		Int32 int
	}
	varint := testInt32Varint{Int32: 150}
	ab, err = cdc.MarshalBinaryBare(varint)
	assert.NoError(t, err, "unexpected error")
	pb, err = proto.Marshal(&p3.TestInt32Varint{Int32: 150})
	assert.NoError(t, err, "unexpected error")
	assert.Equal(t, pb, ab, "varint encoding doesn't match")

	var amToP3 p3.TestInt32Varint
	err = proto.Unmarshal(ab, &amToP3)
	assert.NoError(t, err, "unexpected error")
	assert.Equal(t, uint64(varint.Int32), uint64(amToP3.Int32))

	fixed32 := p3.TestInt32Fixed{Fixed32: 150}
	ab, err = cdc.MarshalBinaryBare(fixed32)
	assert.NoError(t, err, "unexpected error")
	pb, err = proto.Marshal(&fixed32)
	assert.NoError(t, err, "unexpected error")
	assert.Equal(t, pb, ab, "fixed32 encoding doesn't match")

	byteMsg := pbf.Message{Data: []byte("hello cosmos")}
	type bm struct {
		Name       string
		Hilarity   pbf.Message_Humour
		HeightInCm uint32
		Data       []byte
	}
	aminoByteMsg := bm{Data: []byte("hello cosmos")}
	ab, err = cdc.MarshalBinaryBare(aminoByteMsg)
	assert.NoError(t, err, "unexpected error")
	pb, err = proto.Marshal(&byteMsg)
	assert.NoError(t, err, "unexpected error")
	assert.Equal(t, pb, ab, "[]byte encoding doesn't match")

	// there is no way to varsize encode (u)int64 in amino?
	type testUInt64Varint struct {
		Int64 uint64
	}
	varint64 := testUInt64Varint{Int64: 150}
	ab, err = cdc.MarshalBinaryBare(varint64)
	assert.NoError(t, err, "unexpected error")
	pb, err = proto.Marshal(&p3.TestFixedInt64{Int64: 150})
	assert.NoError(t, err, "unexpected error")
	assert.Equal(t, pb, ab, "varint64 encoding doesn't match")
}
