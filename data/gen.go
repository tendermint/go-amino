package data

import (
	"fmt"
	"io"

	"github.com/clipperhouse/typewriter"
)

func init() {
	err := typewriter.Register(NewHolderWriter())
	if err != nil {
		panic(err)
	}
}

type HolderWriter struct{}

func NewHolderWriter() *HolderWriter {
	return &HolderWriter{}
}

func (sw *HolderWriter) Name() string {
	return "holder"
}

func (sw *HolderWriter) Imports(t typewriter.Type) (result []typewriter.ImportSpec) {
	return result
}

func (sw *HolderWriter) Write(w io.Writer, t typewriter.Type) error {
	tag, found := t.FindTag(sw)

	if !found {
		// nothing to be done
		return nil
	}

	license := `// Auto-generated adapters for happily unmarshaling interfaces
// Apache License 2.0
// Copyright (c) 2017 Ethan Frey (ethan.frey@tendermint.com)
`

	if _, err := w.Write([]byte(license)); err != nil {
		fmt.Println("write error")
		return err
	}

	// prepare parameters
	name := t.Name + "Holder"
	if len(tag.Values) > 0 {
		name = tag.Values[0].Name
	}
	m := model{Type: t, Holder: name, Inner: t.Name}

	// now, first main holder
	v := typewriter.TagValue{Name: "Holder"}
	htmpl, err := templates.ByTagValue(t, v)
	if err != nil {
		return err
	}
	if err := htmpl.Execute(w, m); err != nil {
		return err
	}

	// Now, add any implementations...
	v.Name = "Register"
	rtmpl, err := templates.ByTagValue(t, v)
	if err != nil {
		return err
	}
	for _, t := range tag.Values {
		if t.Name == "Impl" {
			for i, p := range t.TypeParameters {
				m.Impl = p
				m.Count = i + 1
				if err := rtmpl.Execute(w, m); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

type model struct {
	Type   typewriter.Type
	Holder string
	Inner  string
	Impl   typewriter.Type // fill in when adding for implementations
	Count  int
}
