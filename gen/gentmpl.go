package data

import (
	"strings"

	"github.com/clipperhouse/typewriter"
)

var templates = typewriter.TemplateSlice{
	holder,
	register,
}

// this is the template for generating the go-data wrappers of an interface
var holder = &typewriter.Template{
	Name:           "Holder",
	TypeConstraint: typewriter.Constraint{},
	FuncMap:        fmap,
	Text: `
type {{.Holder}} struct {
  {{.Inner}} "json:\"unwrap\""
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

/*** below are bindings for each implementation ***/
`,
}

var register = &typewriter.Template{
	Name:           "Register",
	TypeConstraint: typewriter.Constraint{},
	FuncMap:        fmap,
	Text: `
func init() {
  {{.Holder}}Mapper.RegisterImplementation({{ if .Impl.Pointer }}&{{ end }}{{.Impl.Name}}{}, "{{.ImplType }}", 0x{{.Count}})
}

func (hi {{ if .Impl.Pointer }}*{{ end }}{{.Impl.Name}}) Wrap() {{.Holder}} {
  return {{.Holder}}{hi}
}
`,
}

var fmap = map[string]interface{}{
	"ToLower": strings.ToLower,
}
