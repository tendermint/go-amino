package codepress

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// NOTE: flips the order of expected and actual, because wtf mate.
func assertEquals(t *testing.T, expected interface{}, actual interface{}) {
	assert.Equal(t, actual, expected)
}

func TestEmpty(t *testing.T) {
	p := NewCodePress()
	assertEquals(t, p.Print(), "")
}

func TestBasic(t *testing.T) {
	p := NewCodePress()
	p.P("this ")
	p.P("is ")
	p.P("a test")
	assertEquals(t, p.Print(), "this is a test")
}

func TestBasicLn(t *testing.T) {
	p := NewCodePress()
	p.P("this ")
	p.P("is ")
	p.Pln("a test")
	assertEquals(t, p.Print(), "this is a test\n")
}

func TestNewlineStr(t *testing.T) {
	p := NewCodePress().SetNewlineStr("\r\n")
	p.P("this ")
	p.P("is ")
	p.Pln("a test")
	p.Pln("a test")
	p.Pln("a test")
	assertEquals(t, p.Print(), "this is a test\r\na test\r\na test\r\n")
}

func TestIndent(t *testing.T) {
	p := NewCodePress()
	p.P("first line ")
	p.Pln("{").I(func(p *CodePress) {
		p.Pln("second line")
		p.Pln("third line")
	}).P("}")
	assertEquals(t, p.Print(), `first line {
	second line
	third line
}`)
}

func TestIndent2(t *testing.T) {
	p := NewCodePress()
	p.P("first line ")
	p.Pln("{").I(func(p *CodePress) {
		p.P("second ")
		p.P("line")
		// Regardless of whether Pln or Ln is called on cp2,
		// the indented lines terminate with newlineDelim before
		// the next unindented line.
	}).P("}")
	assertEquals(t, p.Print(), `first line {
	second line
}`)
}

func TestIndent3(t *testing.T) {
	p := NewCodePress()
	p.P("first line ")
	p.Pln("{").I(func(p *CodePress) {
		p.P("second ")
		p.Pln("line")
	}).P("}")
	assertEquals(t, p.Print(), `first line {
	second line
}`)
}

func TestIndentLn(t *testing.T) {
	p := NewCodePress()
	p.P("first line ")
	p.Pln("{").I(func(p *CodePress) {
		p.Pln("second line")
		p.Pln("third line")
	}).Pln("}")
	assertEquals(t, p.Print(), `first line {
	second line
	third line
}
`)
}

func TestNestedIndent(t *testing.T) {
	p := NewCodePress()
	p.P("first line ")
	p.Pln("{").I(func(p *CodePress) {
		p.Pln("second line")
		p.Pln("third line")
		p.I(func(p *CodePress) {
			p.Pln("fourth line")
			p.Pln("fifth line")
		})
	}).Pln("}")
	assertEquals(t, p.Print(), `first line {
	second line
	third line
		fourth line
		fifth line
}
`)
}
