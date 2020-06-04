package genproto

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/tendermint/go-amino/press"
)

//----------------------------------------

// NOTE: The goal is not complete Proto3 compatibility (unless there is
// widespread demand for maintaining this repo for that purpose).  Rather, the
// point is to define enough such that the subset that is needed for Amino
// Go->Proto3 is supported.  For example, there is explicitly no plan to
// support the automatic conversion of Proto3->Go, so not all features need to
// be supported.
// NOTE: enums are not supported, as Amino's philosophy is that value checking
// should primarily be done on the application side.

type P3Type interface {
	AssertIsP3Type()
}

func (P3ScalarType) AssertIsP3Type()  {}
func (P3MessageType) AssertIsP3Type() {}

type P3ScalarType string

const (
	P3ScalarTypeDouble   P3ScalarType = "double"
	P3ScalarTypeFloat    P3ScalarType = "float"
	P3ScalarTypeInt32    P3ScalarType = "int32"
	P3ScalarTypeInt64    P3ScalarType = "int64"
	P3ScalarTypeUint32   P3ScalarType = "uint32"
	P3ScalarTypeUint64   P3ScalarType = "uint64"
	P3ScalarTypeSint32   P3ScalarType = "sint32"
	P3ScalarTypeSint64   P3ScalarType = "sint64"
	P3ScalarTypeFixed32  P3ScalarType = "fixed32"
	P3ScalarTypeFixed64  P3ScalarType = "fixed64"
	P3ScalarTypeSfixed32 P3ScalarType = "sfixed32"
	P3ScalarTypeSfixed64 P3ScalarType = "sfixed64"
	P3ScalarTypeBool     P3ScalarType = "bool"
	P3ScalarTypeString   P3ScalarType = "string"
	P3ScalarTypeBytes    P3ScalarType = "bytes"
)

type P3MessageType struct {
	Package string // proto3 package name, optional.
	Name    string // message name.
}

func NewP3MessageType(pkg string, name string) P3MessageType {
	if name == string(P3ScalarTypeDouble) ||
		name == string(P3ScalarTypeFloat) ||
		name == string(P3ScalarTypeInt32) ||
		name == string(P3ScalarTypeInt64) ||
		name == string(P3ScalarTypeUint32) ||
		name == string(P3ScalarTypeUint64) ||
		name == string(P3ScalarTypeSint32) ||
		name == string(P3ScalarTypeSint64) ||
		name == string(P3ScalarTypeFixed32) ||
		name == string(P3ScalarTypeFixed64) ||
		name == string(P3ScalarTypeSfixed32) ||
		name == string(P3ScalarTypeSfixed64) ||
		name == string(P3ScalarTypeBool) ||
		name == string(P3ScalarTypeString) ||
		name == string(P3ScalarTypeBytes) {
		panic(fmt.Sprintf("field type %v already defined", name))
	}
	// check name
	if len(name) == 0 {
		panic("custom p3 type name can't be empty")
	}
	return P3MessageType{Package: pkg, Name: name}
}

func (p3mt P3MessageType) String() string {
	if p3mt.Package == "" {
		return p3mt.Name
	} else {
		return fmt.Sprintf("%v.%v", p3mt.Package, p3mt.Name)
	}
}

// NOTE: P3Doc and its fields are meant to hold basic AST-like information.  No
// validity checking happens here... it should happen before these values are
// set.
type P3Doc struct {
	Package  string // XXX
	Comment  string
	Imports  []P3Import
	Messages []P3Message
	// Enums []P3Enums // enums not supported, no need.
}

type P3Import struct {
	Path string
	// Public bool // not used (yet)
}

type P3Message struct {
	Comment string
	Name    string
	Fields  []P3Field
}

type P3Field struct {
	Comment  string
	Repeated bool
	Type     P3Type
	Name     string
	Number   uint32
}

//----------------------------------------
// Functions for printing P3 objects

// NOTE: P3Doc imports must be set correctly.
func (doc P3Doc) Print() string {
	p := press.NewPress()
	return strings.TrimSpace(doc.PrintCode(p).Print())
}

func (doc P3Doc) PrintCode(p *press.Press) *press.Press {
	p.Pl("syntax = \"proto3\";")
	if doc.Package != "" {
		p.Pl("package %v;", doc.Package)
	}
	// Print comments, if any.
	p.Ln()
	if doc.Comment != "" {
		printComments(p, doc.Comment)
		p.Ln()
	}
	// Print imports, if any.
	for i, imp := range doc.Imports {
		if i == 0 {
			p.Pl("// imports")
		}
		imp.PrintCode(p)
		if i == len(doc.Imports)-1 {
			p.Ln()
		}
	}
	// Print message schemas, if any.
	for i, msg := range doc.Messages {
		if i == 0 {
			p.Pl("// messages")
		}
		msg.PrintCode(p)
		if i == len(doc.Messages)-1 {
			p.Ln()
		}
	}
	return p
}

func (imp P3Import) PrintCode(p *press.Press) *press.Press {
	p.Pl("import %v;", strconv.Quote(imp.Path))
	return p
}

func (msg P3Message) Print() string {
	p := press.NewPress()
	return msg.PrintCode(p).Print()
}

func (msg P3Message) PrintCode(p *press.Press) *press.Press {
	printComments(p, msg.Comment)
	p.Pl("message %v {", msg.Name).I(func(p *press.Press) {
		for _, fld := range msg.Fields {
			fld.PrintCode(p)
		}
	}).Pl("}")
	return p
}

func (fld P3Field) PrintCode(p *press.Press) *press.Press {
	printComments(p, fld.Comment)
	if fld.Repeated {
		p.Pl("repeated %v %v = %v;", fld.Type, fld.Name, fld.Number)
	} else {
		p.Pl("%v %v = %v;", fld.Type, fld.Name, fld.Number)
	}
	return p
}

func printComments(p *press.Press, comment string) {
	if comment == "" {
		return
	}
	commentLines := strings.Split(comment, "\n")
	for _, line := range commentLines {
		p.Pl("// %v", line)
	}
}
