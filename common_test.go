package data_test

import (
	"strings"

	data "github.com/tendermint/go-data"
)

/** These are some sample types to test parsing **/

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

type BigThing struct {
	Name  string
	Age   int
	Dings Fooer
}

/** This is parse code: todo - autogenerate **/

var parser data.Mapper

type FooerS struct {
	Fooer
}

func init() {
	parser = data.NewMapper(FooerS{}).
		RegisterInterface("bar", 0x01, Bar{}).
		RegisterInterface("baz", 0x02, Baz{})
}

func (f FooerS) MarshalJSON() ([]byte, error) {
	return parser.ToJSON(f.Fooer)
}

func (f *FooerS) UnmarshalJSON(data []byte) (err error) {
	parsed, err := parser.FromJSON(data)
	if err == nil {
		f.Fooer = parsed.(Fooer)
	}
	return
}

// Set is a helper to deal with wrapped interfaces
func (f *FooerS) Set(foo Fooer) {
	f.Fooer = foo
}

/** end parse code **/
