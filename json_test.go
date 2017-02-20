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

/** This is parse code: todo - autogenerate **/

var parser *data.Mapper

func init() {
	parser = data.NewMapper().
		RegisterInterface("bar", Bar{}).
		RegisterInterface("baz", Baz{})
}

type FooerS struct {
	Fooer
}

func (f FooerS) MarshalJSON() ([]byte, error) {
	return parser.Marshal(f.Fooer)
}

func (f *FooerS) UnmarshalJSON(data []byte) (err error) {
	parsed, err := parser.Unmarshal(data)
	if err == nil {
		f.Fooer = parsed.(Fooer)
	}
	return
}

/** end parse code **/

func TestSimpleJSON(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	c := Bar{Name: "Fly"}
	assert.Equal("Bar Fly", c.Foo())

	wrap := FooerS{c}
	d, err := json.Marshal(wrap)
	require.Nil(err, "%+v", err)

	parsed := FooerS{}
	err = json.Unmarshal(d, &parsed)
	require.Nil(err, "%+v", err)
	assert.Equal("Bar Fly", parsed.Foo())
	bar, ok := parsed.Fooer.(*Bar)
	assert.True(ok)
	assert.Equal("Fly", bar.Name)

	c2 := Baz{Name: "For Bar"}
	assert.Equal("Foz Baz", c2.Foo())

	wrap2 := FooerS{c2}
	d, err = json.Marshal(wrap2)
	require.Nil(err, "%+v", err)

	parsed = FooerS{}
	err = json.Unmarshal(d, &parsed)
	require.Nil(err, "%+v", err)
	assert.Equal("Foz Baz", parsed.Foo())
}
