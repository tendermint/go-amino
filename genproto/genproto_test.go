package genproto

import (
	"reflect"
	"testing"

	"github.com/jaekwon/testify/assert"
	sm1 "github.com/tendermint/go-amino/genproto/example/submodule"
)

func TestBasic(t *testing.T) {
	p3c := NewP3Context()
	p3c.RegisterPackage(sm1.Package)
	p3doc := P3Doc{Package: "test"}
	obj := sm1.StructSM{}
	p3message, err := p3c.GenerateProto3MessagePartial(&p3doc, reflect.TypeOf(obj))
	assert.Nil(t, err)
	assert.Equal(t, p3message.Print(), `message StructSM {
	int64 FieldA = 1;
	string FieldB = 2;
	submodule2.StructSM2 FieldC = 3;
}
`)

	assert.Equal(t, p3doc.Print(), `syntax = "proto3";
package test;

// imports
import "proto/github.com/tendermint/go-amino/genproto/example/submodule2/types.proto";`)

	p3doc, err = p3c.GenerateProto3Schema("test", reflect.TypeOf(obj))
	assert.Nil(t, err)
	assert.Equal(t, p3doc.Print(), `syntax = "proto3";
package test;

// imports
import "proto/github.com/tendermint/go-amino/genproto/example/submodule2/types.proto";

// messages
message StructSM {
	int64 FieldA = 1;
	string FieldB = 2;
	submodule2.StructSM2 FieldC = 3;
}`)
}
