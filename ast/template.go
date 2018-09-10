package main

import (
	"fmt"
	"strings"
)

/*
 - w: writer
 - err: any error
 - val: value variable of any type
*/

type TEncoder interface {
	TEncode(ref string) (code string)
}

type TypeInt struct{}

// ref: the reference to the value being encoded.
func (_ TypeInt) TEncode(ref string) string {
	return FMT(`{
	var buf [10]byte
	n := binary.PutVarint(buf[:], {{REF}})
	_, err = w.Write(buf[0:n])
}`,
		"REF", ref)
}

type TypeStructField struct {
	Name   string
	Number int
	Type   TEncoder
	// TODO: pointer, etc.
}

func (field TypeStructField) TEncode(ref string) string {
	// XXX What if the field is nil and we need to skip?
	// -> we can do field.Type.TShouldSkip(ref).

	name := field.Name
	fref := ref + "." + name
	return FMT(`{
	// Struct field .{{NAME}}
	// Write field number
	// TODO
	// Write field value
	{{CODE}}
}`,
		"NAME", name,
		"CODE", _INDENT(1, field.Type.TEncode(fref)))
}

type TypeStruct struct {
	Fields []TypeStructField
}

func (st TypeStruct) TEncode(ref string) string {
	// XXX How do we deal with *not* encoding empty structs?
	// -> I guess by remembering w.Len() or something and backtracking.

	fieldBlocks := []string{}
	for _, field := range st.Fields {
		fieldBlocks = append(fieldBlocks, field.TEncode(ref))
	}

	return FMT(`{
	// Struct
	{{CODE}}
}`,
		"CODE", _INDENT(1, strings.Join(fieldBlocks, "\n")))
}

func main() {
	t_int := TypeInt{}
	t_struct := TypeStruct{
		Fields: []TypeStructField{
			{Name: "Foo", Number: 0, Type: t_int},
			{Name: "Bar", Number: 1, Type: t_int},
		},
	}

	code := t_struct.TEncode("var")
	fmt.Println(code)
}

//----------------------------------------

func FMT(tmpl string, args ...string) string {
	if len(args)%2 == 1 {
		panic("FMT args should be key/value pairs")
	}
	s := tmpl
	for i := 0; i < len(args); i += 2 {
		varname := "{{" + args[i] + "}}"
		varval := args[i+1]
		s = strings.Replace(s, varname, varval, -1)
	}
	return s
}

func INDENT(n int, text string) string {
	lines := strings.Split(text, "\n")
	s := ""
	indent := strings.Repeat("\t", n)
	for i, line := range lines {
		if i > 0 {
			s += "\n"
		}
		s += indent + line
	}
	return s
}

// Does not indent the first line.
func _INDENT(n int, text string) string {
	lines := strings.Split(text, "\n")
	s := ""
	indent := strings.Repeat("\t", n)
	for i, line := range lines {
		if i > 0 {
			s += "\n"
		}
		if i > 0 {
			s += indent + line
		} else {
			s += line
		}
	}
	return s
}
