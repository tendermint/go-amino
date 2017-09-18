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

package data

import (
	"github.com/pkg/errors"
	wire "github.com/tendermint/go-wire"
)

type binaryMapper struct {
	base  interface{}
	impls []wire.ConcreteType
}

func newBinaryMapper(base interface{}) *binaryMapper {
	return &binaryMapper{
		base: base,
	}
}

// registerImplementation allows you to register multiple concrete types.
//
// We call wire.RegisterInterface with the entire (growing list) each time,
// as we do not know when the end is near.
func (m *binaryMapper) registerImplementation(data interface{}, kind string, b byte) {
	m.impls = append(m.impls, wire.ConcreteType{O: data, Byte: b})
	wire.RegisterInterface(m.base, m.impls...)
}

// ToWire is a convenience method to serialize with go-wire
// error is there to keep the same interface as json, but always nil
func ToWire(o interface{}) ([]byte, error) {
	return wire.BinaryBytes(o), nil
}

// FromWire is a convenience method to deserialize with go-wire
func FromWire(d []byte, o interface{}) error {
	return errors.WithStack(
		wire.ReadBinaryBytes(d, o))
}
