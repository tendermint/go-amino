package press

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// NOTE: actual first.
func assertEquals(t *testing.T, actual interface{}, expected interface{}) {
	assert.Equal(t, expected, actual)
}

func TestEmpty(t *testing.T) {
	p := NewPress()
	assertEquals(t, p.Print(), "")
}

func TestBasic(t *testing.T) {
	p := NewPress()
	p.P("this ")
	p.P("is ")
	p.P("a test")
	assertEquals(t, p.Print(), "this is a test")
}

func TestBasicLn(t *testing.T) {
	p := NewPress()
	p.P("this ")
	p.P("is ")
	p.Pl("a test")
	assertEquals(t, p.Print(), "this is a test\n")
}

func TestNewlineStr(t *testing.T) {
	p := NewPress().SetNewlineStr("\r\n")
	p.P("this ")
	p.P("is ")
	p.Pl("a test")
	p.Pl("a test")
	p.Pl("a test")
	assertEquals(t, p.Print(), "this is a test\r\na test\r\na test\r\n")
}

func TestIndent(t *testing.T) {
	p := NewPress()
	p.P("first line ")
	p.Pl("{").I(func(p *Press) {
		p.Pl("second line")
		p.Pl("third line")
	}).P("}")
	assertEquals(t, p.Print(), `first line {
	second line
	third line
}`)
}

func TestIndent2(t *testing.T) {
	p := NewPress()
	p.P("first line ")
	p.Pl("{").I(func(p *Press) {
		p.P("second ")
		p.P("line")
		// Regardless of whether Pl or Ln is called on cp2,
		// the indented lines terminate with newlineDelim before
		// the next unindented line.
	}).P("}")
	assertEquals(t, p.Print(), `first line {
	second line
}`)
}

func TestIndent3(t *testing.T) {
	p := NewPress()
	p.P("first line ")
	p.Pl("{").I(func(p *Press) {
		p.P("second ")
		p.Pl("line")
	}).P("}")
	assertEquals(t, p.Print(), `first line {
	second line
}`)
}

func TestIndentLn(t *testing.T) {
	p := NewPress()
	p.P("first line ")
	p.Pl("{").I(func(p *Press) {
		p.Pl("second line")
		p.Pl("third line")
	}).Pl("}")
	assertEquals(t, p.Print(), `first line {
	second line
	third line
}
`)
}

func TestNestedIndent(t *testing.T) {
	p := NewPress()
	p.P("first line ")
	p.Pl("{").I(func(p *Press) {
		p.Pl("second line")
		p.Pl("third line")
		p.I(func(p *Press) {
			p.Pl("fourth line")
			p.Pl("fifth line")
		})
	}).Pl("}")
	assertEquals(t, p.Print(), `first line {
	second line
	third line
		fourth line
		fifth line
}
`)
}
