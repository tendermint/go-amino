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

package wire_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	wire "github.com/tendermint/go-wire"
)

func TestEndToEndReflectBinary(t *testing.T) {

	type Receiver interface{}
	type bcMessage struct {
		Message string
		Height  int
	}

	type bcResponse struct {
		Status  int
		Message string
	}

	type bcStatus struct {
		Peers int
	}

	/*bm := &bcMessage{Message: "ABC", Height: 100}
	unregistered, err := wire.MarshalBinary(bm)
	assert.Nil(t, err)
	fmt.Println("### normal", unregistered)*/

	wire2 := wire.NewCodec()
	wire2.RegisterInterface((*Receiver)(nil), nil)
	wire2.RegisterConcrete(&bcMessage{}, "bcMessage", nil)
	wire2.RegisterConcrete(&bcResponse{}, "bcResponse", nil)
	wire2.RegisterConcrete(&bcStatus{}, "bcStatus", nil)
	fmt.Println("registered")

	fmt.Println("-------")
	bm := &bcMessage{Message: "ABC", Height: 100}

	bmBytes, err := wire2.MarshalBinary(bm)
	assert.Nil(t, err)
	fmt.Println("### registered bytes", bmBytes)
	return
	t.Logf("Encoded: %x\n", bmBytes)

	var rcvr Receiver
	err = wire2.UnmarshalBinary(bmBytes, &rcvr)
	assert.Nil(t, err)
	bm2 := rcvr.(*bcMessage)
	t.Logf("Decoded: %#v\n", bm2)

	assert.Equal(t, bm, bm2)
}
