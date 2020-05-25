package codepress

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/tendermint/go-amino/libs/detrand"
)

type line struct {
	indentStr string // the fully expanded indentation string
	value     string // the original contents of the line
}

func newLine(indentStr string, value string) line {
	return line{indentStr, value}
}

func (l line) String() string {
	return l.indentStr + l.value
}

// shortcut
var fmt_ = fmt.Sprintf

// CodePress is a tool for printing code.
// CodePress is not concurrency safe.
type CodePress struct {
	rnd          *rand.Rand // for generating a random variable names.
	indentPrefix string     // current indent prefix.
	indentDelim  string     // a tab or spaces, whatever.
	newlineStr   string     // should probably just remain "\n".
	lines        []line     // accumulated lines from printing.
}

func NewCodePress() *CodePress {
	return &CodePress{
		rnd:          rand.New(rand.NewSource(0)),
		indentPrefix: "",
		indentDelim:  "\t",
		newlineStr:   "\n",
		lines:        nil,
	}
}

func (cp *CodePress) SetIndentDelim(s string) *CodePress {
	cp.indentDelim = s
	return cp
}

func (cp *CodePress) SetNewlineStr(s string) *CodePress {
	cp.newlineStr = s
	return cp
}

// Main function for printing something on the press.
func (cp *CodePress) P(s string, args ...interface{}) *CodePress {
	var l *line
	if len(cp.lines) == 0 {
		// Make a new line.
		cp.lines = []line{newLine(cp.indentPrefix, "")}
	}
	// Get ref to last line.
	l = &(cp.lines[len(cp.lines)-1])
	l.value += fmt.Sprintf(s, args...)
	return cp
}

// Appends a new line.
// It is also possible to print newline characters direclty,
// but CodePress doesn't treat them as newlines for the sake of indentation.
func (cp *CodePress) Ln() *CodePress {
	cp.lines = append(cp.lines, newLine(cp.indentPrefix, ""))
	return cp
}

// Convenience for P() followed by Nl().
func (cp *CodePress) Pln(s string, args ...interface{}) *CodePress {
	return cp.P(s, args...).Ln()
}

// auto-indents cp2, appends concents to cp.
// Panics if the last call wasn't Pln() or Ln().
// Regardless of whether Pln or Ln is called on cp2,
// the indented lines terminate with newlineDelim before
// the next unindented line.
func (cp *CodePress) I(block func(cp2 *CodePress)) *CodePress {
	if len(cp.lines) > 0 {
		lastLine := cp.lines[len(cp.lines)-1]
		if lastLine.value != "" {
			panic("cannot indent after nonempty line")
		}
		if lastLine.indentStr != cp.indentPrefix {
			panic("unexpected indent string in last line")
		}
		// remove last empty line
		cp.lines = cp.lines[:len(cp.lines)-1]
	}
	cp2 := cp.SubCodePress()
	cp2.indentPrefix = cp.indentPrefix + cp.indentDelim
	block(cp2)
	ilines := cp2.Lines()
	// remove last empty line from cp2
	ilines = withoutFinalNewline(ilines)
	cp.lines = append(cp.lines, ilines...)
	// (re)introduce last line with original indent
	cp.lines = append(cp.lines, newLine(cp.indentPrefix, ""))
	return cp
}

// Prints the final representation of the contents.
func (cp *CodePress) Print() string {
	lines := []string{}
	for _, line := range cp.lines {
		lines = append(lines, line.String())
	}
	return strings.Join(lines, cp.newlineStr)
}

// Returns the lines.
// This may be useful for adding additional indentation to each line for code blocks.
func (cp *CodePress) Lines() (lines []line) {
	return cp.lines
}

// Convenience
func (cp *CodePress) RandID(prefix string) string {
	return prefix + "_" + cp.RandStr(8)
}

// Convenience
func (cp *CodePress) RandStr(length int) string {
	const strChars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz" // 62 characters
	chars := []byte{}
MAIN_LOOP:
	for {
		val := cp.rnd.Int63()
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

// SubCodePress creates a blank CodePress suitable for inlining code.
// It starts with no indentation, zero lines,
// a derived rand from the original, but the same indent and nl strings..
func (cp *CodePress) SubCodePress() *CodePress {
	cp2 := NewCodePress()
	cp2.rnd = detrand.DeriveRand(cp.rnd)
	cp2.indentPrefix = ""
	cp2.indentDelim = cp.indentDelim
	cp2.newlineStr = cp.newlineStr
	cp2.lines = nil
	return cp2
}

// ref: the reference to the value being encoded.
type EncoderPressFunc func(cp *CodePress, ref string) (code string)

//----------------------------------------

// If the final line is a line with no value, remove it
func withoutFinalNewline(lines []line) []line {
	if len(lines) > 0 && lines[len(lines)-1].value == "" {
		return lines[:len(lines)-1]
	} else {
		return lines
	}
}
