package main

/*
NOTE: golang's templating system is not ergonomic for code generation.
It's actually easier to write a custom generator like here.
*/

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
)

type TContext struct {
	rnd *rand.Rand
}

func NewTContext() *TContext {
	return &TContext{
		rnd: rand.New(rand.NewSource(0)),
	}
}

func (ctx *TContext) RandID(prefix string) string {
	return prefix + "_" + ctx.RandStr(8)
}

// NOTE: Copied from tendermint/libs/common/random.go
func (ctx *TContext) RandStr(length int) string {
	const strChars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz" // 62 characters
	chars := []byte{}
MAIN_LOOP:
	for {
		val := ctx.rnd.Int63()
		for i := 0; i < 10; i++ {
			v := int(val & 0x3f) // rightmost 6 bits
			if v >= 62 {         // only 62 characters in strChars
				val >>= 6
				continue
			} else {
				chars = append(chars, strChars[v])
				if len(chars) == length {
					break MAIN_LOOP
				}
				val >>= 6
			}
		}
	}
	return string(chars)
}

/*
 - w: writer
 - err: any error
 - val: value variable of any type
*/

type TEncoder interface {
	TEncode(ctx *TContext, ref string) (code string)
}

type IntType struct{}

// ref: the reference to the value being encoded.
func (_ IntType) TEncode(ctx *TContext, ref string) string {
	return FMT(`{
	var buf [10]byte
	n := binary.PutVarint(buf[:], {{REF}})
	_, err = w.Write(buf[0:n])
}`,
		"REF", ref)
}

type StructFieldType struct {
	Name       string
	Number     int
	Type       TEncoder
	Kind       reflect.Kind
	WriteEmpty bool
}

// ref: reference to the struct
func (field StructFieldType) TEncode(ctx *TContext, ref string) string {
	name := field.Name
	done := ctx.RandID("done")
	fref := ref + "." + name
	return FMT(`{
	// Struct field .{{NAME}}
	// Maybe skip?
	if ({{COND}}) {
		goto {{DONE}}
	}
	pos1 := w.Len()
	// Write field number & typ3
	// TODO
	pos2 := w.Len()
	// Write field value
	{{CODE}}
	pos3 := w.Len()
	if (!{{ISPTR}} && !{{WRITEEMPTY}} && pos2 == pos3-1 && w.PeekLastByte() == 0x00) {
		w.Truncate(pos1)
	}
{{DONE}}:
}`,
		"NAME", name,
		"COND", field.TEncodeSkipCond(ctx, fref),
		"CODE", _INDENT(1, field.Type.TEncode(ctx, fref)),
		"DONE", done,
		"ISPTR", fmt.Sprintf("%v", field.Kind == reflect.Ptr),
		"WRITEEMPTY", fmt.Sprintf("%v", field.WriteEmpty),
	)
}

// ref: reference to the struct field
func (field StructFieldType) TEncodeSkipCond(ctx *TContext, fref string) string {
	// If the value is nil or empty, do not encode.
	// Field values that are Empty struct are not skipped here,
	// but rather via StructFieldType.TEncode.
	switch field.Kind {
	case reflect.Ptr:
		return FMT("{{FREF}} == nil", "FREF", fref)
	case reflect.Bool:
		return FMT("{{FREF}} == false", "FREF", fref)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return FMT("{{FREF}} == 0", "FREF", fref)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return FMT("{{FREF}} == 0", "FREF", fref)
	case reflect.String:
		return FMT("len({{FREF}}) == 0", "FREF", fref)
	case reflect.Chan, reflect.Map, reflect.Slice:
		return FMT("{{FREF}} == nil || len({{FREF}}) == 0", "FREF", fref)
	case reflect.Func, reflect.Interface:
		return FMT("{{FREF}} == nil", "FREF", fref)
	default:
		return "true"
	}
}

type StructType struct {
	Fields []StructFieldType
}

func (st StructType) TEncode(ctx *TContext, ref string) string {
	// XXX How do we deal with *not* encoding empty structs?
	// -> I guess by remembering w.Len() or something and backtracking.
	// -> Well, that isn't supposed to happen here.
	// -> it should happen for field values, not for the whole struct.

	fieldBlocks := []string{}
	for _, field := range st.Fields {
		fieldBlocks = append(fieldBlocks, field.TEncode(ctx, ref))
	}

	return FMT(`{
	// Struct
	{{CODE}}
}`,
		"CODE", _INDENT(1, strings.Join(fieldBlocks, "\n")))
}

func main() {
	t_int := IntType{}
	t_struct := StructType{
		Fields: []StructFieldType{
			{Name: "Foo", Number: 0, Type: t_int, Kind: reflect.Int},
			{Name: "Bar", Number: 1, Type: t_int, Kind: reflect.Int32},
		},
	}
	ctx := NewTContext()
	code := t_struct.TEncode(ctx, "var")
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
