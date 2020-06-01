package gengo

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/tendermint/go-amino"
	"github.com/tendermint/go-amino/codepress"
)

type SampleStruct struct {
	Blah1 string
	Blah2 string
}

func TestBasic(t *testing.T) {
	p := codepress.NewCodePress()
	fmt.Println(p)
	ss := SampleStruct{"cat", "dog"}

	cdc := amino.NewCodec()
	ssType := reflect.TypeOf(ss)
	info, err := cdc.GetTypeInfo(ssType)
	if err != nil {
		panic(err)
	}
	PrintStructFieldEncoder(p, "b1", info.Fields[0])
	fmt.Println("----")
	fmt.Println(p.Print())
	fmt.Println("----")
}
