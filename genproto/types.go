package genproto

import (
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

type P3FieldType string

const (
	P3FieldTypeDouble   P3FieldType = "double"
	P3FieldTypeFloat    P3FieldType = "float"
	P3FieldTypeInt32    P3FieldType = "int32"
	P3FieldTypeInt64    P3FieldType = "int64"
	P3FieldTypeUInt32   P3FieldType = "uint32"
	P3FieldTypeUInt64   P3FieldType = "uint64"
	P3FieldTypeSInt32   P3FieldType = "sint32"
	P3FieldTypeFixed32  P3FieldType = "fixed32"
	P3FieldTypeFixed64  P3FieldType = "fixed64"
	P3FieldTypeSFixed32 P3FieldType = "sfixed32"
	P3FieldTypeSFixed64 P3FieldType = "sfixed64"
	P3FieldTypeSInt64   P3FieldType = "sint64"
	P3FieldTypeBool     P3FieldType = "bool"
	P3FieldTypeString   P3FieldType = "string"
	P3FieldTypeBytes    P3FieldType = "bytes"
)

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
	Type     P3FieldType
	Name     string
	Number   int
	Repeated bool
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
