package data

import "github.com/clipperhouse/typewriter"

// this is the template for generating the go-data wrappers of an interface
var tmpl = &typewriter.Template{
	Name:           "Holder",
	TypeConstraint: typewriter.Constraint{},
	Text: `
import (
  "github.com/tendermint/go-wire/data"
)

type {{.Holder}} struct {
  {{.Inner}}
}

var {{.Holder}}Mapper = data.NewMapper({{.Holder}}{})

func (h {{.Holder}}) MarshalJSON() ([]byte, error) {
  return {{.Holder}}Mapper.ToJSON(h.{{.Inner}})
}

func (h *{{.Holder}}) UnmarshalJSON(data []byte) (err error) {
  parsed, err := {{.Holder}}Mapper.FromJSON(data)
  if err == nil && parsed != nil {
    h.{{.Inner}} = parsed.({{.Inner}})
  }
  return err
}

// Unwrap recovers the concrete interface safely (regardless of levels of embeds)
func (h {{.Holder}}) Unwrap() {{.Inner}} {
  hi := h.{{.Inner}}
  for wrap, ok := hi.({{.Holder}}); ok; wrap, ok = hi.({{.Holder}}) {
    hi = wrap.{{.Inner}}
  }
  return hi
}

func (h {{.Holder}}) Empty() bool {
  return h.{{.Inner}} == nil
}
`,
}

// RegisterImplementation(PubKeyEd25519{}, NameEd25519, TypeEd25519).
// RegisterImplementation(PubKeySecp256k1{}, NameSecp256k1, TypeSecp256k1)
