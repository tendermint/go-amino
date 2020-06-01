package genproto

import (
	"fmt"
	"strings"

	"github.com/tendermint/go-amino/press"
)

// NOTE: The goal is not complete Proto3 compatibility (unless there is
// widespread demand for maintaining this repo for that purpose).  Rather, the
// point is to define enough such that the subset that is needed for Amino
// Go->Proto3 is supported.  For example, there is explicitly no plan to
// support the automatic conversion of Proto3->Go, so not all features need to
// be supported.
// NOTE: enums are not supported, as Amino's philosophy is that value checking
// should primarily be done on the application side.

type P3Type string

const (
	P3TypeDouble   P3Type = "double"
	P3TypeFloat    P3Type = "float"
	P3TypeInt32    P3Type = "int32"
	P3TypeInt64    P3Type = "int64"
	P3TypeUint32   P3Type = "uint32"
	P3TypeUint64   P3Type = "uint64"
	P3TypeSint32   P3Type = "sint32"
	P3TypeSint64   P3Type = "sint64"
	P3TypeFixed32  P3Type = "fixed32"
	P3TypeFixed64  P3Type = "fixed64"
	P3TypeSfixed32 P3Type = "sfixed32"
	P3TypeSfixed64 P3Type = "sfixed64"
	P3TypeBool     P3Type = "bool"
	P3TypeString   P3Type = "string"
	P3TypeBytes    P3Type = "bytes"
)

func NewCustomP3Type(typeName string) P3Type {
	if typeName == string(P3TypeDouble) ||
		typeName == string(P3TypeFloat) ||
		typeName == string(P3TypeInt32) ||
		typeName == string(P3TypeInt64) ||
		typeName == string(P3TypeUint32) ||
		typeName == string(P3TypeUint64) ||
		typeName == string(P3TypeSint32) ||
		typeName == string(P3TypeSint64) ||
		typeName == string(P3TypeFixed32) ||
		typeName == string(P3TypeFixed64) ||
		typeName == string(P3TypeSfixed32) ||
		typeName == string(P3TypeSfixed64) ||
		typeName == string(P3TypeBool) ||
		typeName == string(P3TypeString) ||
		typeName == string(P3TypeBytes) {
		panic(fmt.Sprintf("field type %v already defined", typeName))
	}
	// check typeName
	if len(typeName) == 0 {
		panic("custom p3 type name can't be empty")
	}
	return P3Type(typeName)
}

type P3Doc struct {
	Comment  string
	Messages []P3Message
	// Enums []P3Enums // enums not supported, no need.
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

func (doc P3Doc) Print() string {
	p := press.NewPress()
	return doc.PrintCode(p).Print()
}

func (doc P3Doc) PrintCode(p *press.Press) *press.Press {
	p.Pl("syntax = \"proto3\";")
	printComments(p, doc.Comment)
	for _, msg := range doc.Messages {
		p.Ln()
		msg.PrintCode(p)
	}
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
