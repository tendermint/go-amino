package data_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	data "github.com/tendermint/go-data"
)

type Fooer interface {
	Foo() string
}

type Bar struct {
	Name string
}

func (b Bar) Foo() string {
	return "Bar " + b.Name
}

type Baz struct {
	Name string
}

func (b Baz) Foo() string {
	return strings.Replace(b.Name, "r", "z", -1)
}

var parser *data.Envelope

func init() {
	parser = data.NewEnvelope().
		RegisterInterface("bar", Bar{}).
		RegisterInterface("baz", Baz{})
}

func TestSimpleJSON(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	env := data.Envelope{
		Kind: "bar",
		Msg:  Bar{Name: "Fly"},
	}
	d, err := json.Marshal(env)
	require.Nil(err, "%+v", err)

	parsed := parser.New()
	err = json.Unmarshal(d, &parsed)
	require.Nil(err, "%+v", err)
	assert.Equal("bar", parsed.Kind)
	bar, ok := parsed.Msg.(*Bar)
	assert.True(ok)
	assert.Equal("Fly", bar.Name)
}
