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
	p3message, err := p3c.GenerateProto3MessageSchema(cdc, reflect.TypeOf(obj))
	assert.Nil(t, err)
	t.Log(p3message.Print())

	// XXX
}
