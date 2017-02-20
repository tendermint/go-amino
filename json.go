package data

import (
	"encoding/json"
	"reflect"

	"github.com/pkg/errors"
)

type Mapper struct {
	kindToType map[string]reflect.Type `json:"-"`
	typeToKind map[reflect.Type]string `json:"-"`
}

func NewMapper() *Mapper {
	return &Mapper{
		kindToType: map[string]reflect.Type{},
		typeToKind: map[reflect.Type]string{},
	}
}

// RegisterInterface allows you to register multiple concrete types.
//
// Returns itself to allow calls to be chained
func (m *Mapper) RegisterInterface(kind string, data interface{}) *Mapper {
	typ := reflect.TypeOf(data)
	m.kindToType[kind] = typ
	m.typeToKind[typ] = kind
	return m
}

func (m *Mapper) getTarget(kind string) (interface{}, error) {
	typ, ok := m.kindToType[kind]
	if !ok {
		return nil, errors.Errorf("Unmarshaling into unknown type: %s", kind)
	}
	target := reflect.New(typ).Interface()
	return target, nil
}

func (m *Mapper) getKind(obj interface{}) (string, error) {
	typ := reflect.TypeOf(obj)
	kind, ok := m.typeToKind[typ]
	if !ok {
		return "", errors.Errorf("Marshalling from unknown type: %#v", obj)
	}
	return kind, nil
}

func (m *Mapper) Unmarshal(data []byte) (interface{}, error) {
	e := envelope{
		Msg: &json.RawMessage{},
	}
	err := json.Unmarshal(data, &e)
	if err != nil {
		return nil, err
	}
	// switch on the type, then unmarshal into that
	bytes := *e.Msg.(*json.RawMessage)
	res, err := m.getTarget(e.Kind)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &res)
	return res, err
}

func (m *Mapper) Marshal(data interface{}) ([]byte, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	kind, err := m.getKind(data)
	if err != nil {
		return nil, err
	}
	msg := json.RawMessage(raw)
	e := envelope{
		Kind: kind,
		Msg:  &msg,
	}
	return json.Marshal(e)
}

// envelope lets us switch on type
type envelope struct {
	Kind string      `json:"type"`
	Msg  interface{} `json:"msg"`
}
