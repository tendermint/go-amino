package ex

import (
	"fmt"

	"github.com/tendermint/tmlibs/common"
)

// +gen holder:"Foo,Impl[Bling,*Fuzz],bb,ff"
type FooInner interface {
	Bar() int
}

type Bling struct{}

func (b Bling) Bar() int {
	return common.RandInt()
}

type Fuzz struct{}

func (f *Fuzz) Bar() int {
	fmt.Println("hello")
	return 42
}
