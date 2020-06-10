package genproto

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	sm1 "github.com/tendermint/go-amino/genproto/example/submodule"
)

func TestBasic(t *testing.T) {
	p3c := NewP3Context()
	p3doc := P3Doc{}
	obj := sm1.StructSM{}
	p3message, err := p3c.GenerateProto3MessagePartial(&p3doc, reflect.TypeOf(obj))
	assert.Nil(t, err)
	assertEquals(t, p3message.Print(), `message StructSM {
	int64 FieldA = 1;
	string FieldB = 2;
	go_amino.genproto.example.submodule2.StructSM2 FieldC = 3;
}
`)

	assertEquals(t, p3doc.Print(), `syntax = "proto3";

// imports
import "vendor/github.com/tendermint/go-amino/genproto/example/submodule2/types.proto";
`)

	p3doc, err = p3c.GenerateProto3Schema("", reflect.TypeOf(obj))
	assert.Nil(t, err)
	assertEquals(t, p3doc.Print(), `syntax = "proto3";

// imports
import "vendor/github.com/tendermint/go-amino/genproto/example/submodule2/types.proto";

// messages
message StructSM {
	int64 FieldA = 1;
	string FieldB = 2;
	go_amino.genproto.example.submodule2.StructSM2 FieldC = 3;
}`)
}
