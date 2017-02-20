package data

import (
	"encoding/json"
	"reflect"

	"github.com/pkg/errors"
)

// Envelope lets us switch on type
type Envelope struct {
	Kind       string                  `json:"type"`
	Msg        interface{}             `json:"msg"`
	kindToType map[string]reflect.Type `json:"-"`
	typeToKind map[reflect.Type]string `json:"-"`
}

// NewEnvelope must be called to construct a base envelope to unmarshall info
func NewEnvelope() *Envelope {
	return &Envelope{
		kindToType: map[string]reflect.Type{},
		typeToKind: map[reflect.Type]string{},
	}
}

// RegisterInterface allows you to register multiple concrete types
// to one Envelope.
//
// The configured envelope should be saved as a singleton in that
// package and copied with New() in order to unmarshal json.
//
// Returns itself to allow calls to be chained
func (e *Envelope) RegisterInterface(kind string, data interface{}) *Envelope {
	typ := reflect.TypeOf(data)
	e.kindToType[kind] = typ
	e.typeToKind[typ] = kind
	return e
}

func (e Envelope) New() Envelope {
	return Envelope{
		kindToType: e.kindToType,
		typeToKind: e.typeToKind,
	}
}

type Bling struct {
	Name string
}

func (e *Envelope) getTarget(kind string) (interface{}, error) {
	typ, ok := e.kindToType[kind]
	if !ok {
		return nil, errors.Errorf("Unmarshaling into unknown type: %s", kind)
	}
	target := reflect.New(typ).Interface()
	return target, nil
}

func (e *Envelope) UnmarshalJSON(data []byte) error {
	type Alias Envelope
	e.Msg = &json.RawMessage{}
	err := json.Unmarshal(data, (*Alias)(e))
	if err != nil {
		return err
	}
	// switch on the type, then unmarshal into that
	bytes := *e.Msg.(*json.RawMessage)
	e.Msg, err = e.getTarget(e.Kind)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, &e.Msg)
}
