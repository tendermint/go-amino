// Copyright 2017 Tendermint. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ex

import (
	"fmt"

	"github.com/tendermint/tmlibs/common"
)

// +gen wrapper:"Foo,Impl[Bling,*Fuzz],blng,fzz"
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
                                                                                                                                                                                                                                                                                                                  