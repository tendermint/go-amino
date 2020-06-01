package genproto

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// NOTE: actual first.
func assertEquals(t *testing.T, actual interface{}, expected interface{}) {
	assert.Equal(t, expected, actual)
}

func TestPrintP3Types(t *testing.T) {
	doc := P3Doc{
		Comment: "doc comment",
		Syntax:  "some_syntax",
		Messages: []P3Message{
			P3Message{
				Comment: "message comment",
				Name:    "message_name",
				Fields: []P3Field{
					P3Field{
						Comment:  "field_comment",
						Type:     P3FieldTypeString,
						Name:     "field_name",
						Number:   1,
						Repeated: false,
					},
					P3Field{
						Comment:  "field_comment",
						Type:     P3FieldTypeUInt64,
						Name:     "field_name",
						Number:   2,
						Repeated: true,
					},
				},
			},
			P3Message{
				Comment: "message comment 2",
				Name:    "message_name_2",
				Fields:  []P3Field{},
			},
		},
	}

	proto3Schema := doc.Print()
	assertEquals(t, proto3Schema, `// Auto-generated Proto3 schema file, generatedy by go-amino
// doc comment

syntax = some_syntax

// message comment
message message_name {
	// field_comment
	string field_name = 1;
	// field_comment
	repeated uint64 field_name = 2;
}

// message comment 2
message message_name_2 {
}
`)
}
