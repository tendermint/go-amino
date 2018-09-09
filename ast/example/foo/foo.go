package foo

import (
	"github.com/tendermint/go-amino/ast/example/bar"
)

type Foo struct {
	A string
	B *bar.Bar
}

func (_ Foo) something() {}

func Something() {}
