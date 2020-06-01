package genproto

import (
	"testing"
)

func TestBasic(t *testing.T) {
	GenerateProtoForPatterns("./example")
}
