package data

import (
	"encoding/json"
	"reflect"

	"github.com/pkg/errors"
)

type JSONMapper struct {
	kindToType map[string]reflect.Type
	typeToKind map[reflect.Type]string
}

func NewJSONMapper(base interface{}) *JSONMapper {
	return &JSONMapper{
		kindToType: map[string]reflect.Type{},
		typeToKind: map[reflect.Type]string{},
	}
}

// RegisterInterface allows you to register multiple concrete types.
//
// Returns itself to allow calls to be chained
func (m *JSONMapper) RegisterInterface(kind string, b byte, data interface{}) {
	typ := reflect.TypeOf(data)
	m.kindToType[kind] = typ
	m.typeToKind[typ] = kind
}

func (m *JSONMapper) getTarget(kind string) (interface{}, error) {
	typ, ok := m.kindToType[kind]
	if !ok {
		return nil, errors.Errorf("Unmarshaling into unknown type: %s", kind)
	}
	target := reflect.New(typ).Interface()
	return target, nil
}

func (m *JSONMapper) getKind(obj interface{}) (string, error) {
	typ := reflect.TypeOf(obj)
	kind, ok := m.typeToKind[typ]
	if !ok {
		return "", errors.Errorf("Marshalling from unknown type: %#v", obj)
	}
	return kind, nil
}

func (m *JSONMapper) FromJSON(data []byte) (interface{}, error) {
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

func (m *JSONMapper) ToJSON(data interface{}) ([]byte, error) {
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
