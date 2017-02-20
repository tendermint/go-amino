package data

import wire "github.com/tendermint/go-wire"

type BinaryMapper struct {
	base  interface{}
	impls []wire.ConcreteType
}

func NewBinaryMapper(base interface{}) *BinaryMapper {
	return &BinaryMapper{
		base: base,
	}
}

// RegisterInterface allows you to register multiple concrete types.
//
// We call wire.RegisterInterface with the entire (growing list) each time,
// as we do not know when the end is near.
func (m *BinaryMapper) RegisterInterface(kind string, b byte, data interface{}) {
	m.impls = append(m.impls, wire.ConcreteType{O: data, Byte: b})
	wire.RegisterInterface(m.base, m.impls...)
}
