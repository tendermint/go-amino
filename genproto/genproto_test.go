package genproto

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	sm1 "github.com/tendermint/go-amino/genproto/example/submodule"
)

func TestBasic(t *testing.T) {
	p3c := NewP3Context()
	obj := sm1.StructSM{}
	p3message, err := p3c.GenerateProto3MessagePartial(reflect.TypeOf(obj))
	assert.Nil(t, err)
	assertEquals(t, p3message.Print(), `message StructSM {
	int64 FieldA = 1;
	string FieldB = 2;
	go_amino.genproto.example.submodule2.StructSM2 FieldC = 3;
}
`)

	p3doc, err := p3c.GenerateProto3Schema("", reflect.TypeOf(obj))
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

func TestDefaultP3pkgFromGopkg(t *testing.T) {
	p3c := NewP3Context()
	p3c.RegisterPackageMapping("github.com/tendermint/tendermint/go-amino/example", "example", nil)

	testDefault := func(gopkg string, expected string) {
		assertEquals(t, DeriveDefaultP3pkgFromGopkg(p3c, gopkg), expected)
	}

	// NOTE: add desired mapping invariants here and make the function intelligent.
	testDefault("github.com/tendermint/tendermint/go-amino", "tendermint.go_amino")
	testDefault("github.com/tendermint/tendermint/go-amino/example", "example")
	testDefault("github.com/tendermint/tendermint/go-amino/example/foo", "example.foo")
	testDefault("github.com/tendermint/tendermint/go-amino/example/foo-bar", "example.foo_bar")
	testDefault("google.golang.org/protobuf/types/known/anypb", "protobuf.types.known.anypb")
	testDefault("go/ast", "go.ast")
	testDefault("math/big", "math.big")
	testDefault("sort", "sort")
}
