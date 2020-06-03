package genproto

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/go-amino"
	sm1 "github.com/tendermint/go-amino/genproto/example/submodule"
)

func TestBasic(t *testing.T) {
	p3c := NewP3Context()
	cdc := amino.NewCodec()
	obj := sm1.StructSM{}
	p3message, err := p3c.GenerateProto3MessagePartial(cdc, reflect.TypeOf(obj))
	assert.Nil(t, err)
	assertEquals(t, p3message.Print(), `message StructSM {
	int64 FieldA = 1;
	string FieldB = 2;
	tendermint.go-amino.genproto.example.submodule2.StructSM2 FieldC = 3;
}
`)

	p3doc, err := p3c.GenerateProto3Schema(cdc, reflect.TypeOf(obj))
	assert.Nil(t, err)
	assertEquals(t, p3doc.Print(), `syntax = "proto3";
import "vendor/github.com/tendermint/go-amino/genproto/example/submodule2/types.proto";

message StructSM {
	int64 FieldA = 1;
	string FieldB = 2;
	tendermint.go-amino.genproto.example.submodule2.StructSM2 FieldC = 3;
}
`)
}
