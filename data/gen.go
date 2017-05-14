package data

import (
	"bytes"
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
	for _, t := range tag.Values {
		if t.Name == "Impl" {
			for _, p := range t.TypeParameters {
				fmt.Printf("param: %#v\n", p)
			}
		}
	}

	license := `// Auto-generated adapters for happily unmarshaling interfaces
// Apache License 2.0
// Copyright (c) 2017 Ethan Frey (ethan.frey@tendermint.com)
`

	if _, err := w.Write([]byte(license)); err != nil {
		fmt.Println("write error")
		return err
	}

	ptmpl, err := tmpl.Parse()
	if err != nil {
		return err
	}

	// prepare parameters
	name := t.Name + "Holder"
	if len(tag.Values) > 0 {
		name = tag.Values[0].Name
	}
	m := model{Type: t, Holder: name, Inner: t.Name}

	bw := bytes.NewBuffer(nil)
	if err := ptmpl.Execute(bw, m); err != nil {
		return err
	}
	fmt.Print(license)
	byt := bw.Bytes()
	fmt.Println(string(byt))
	w.Write(byt)

	// if err := ptmpl.Execute(w, m); err != nil {
	//   return err
	// }

	return nil
}

type model struct {
	Type   typewriter.Type
	Holder string
	Inner  string
}
