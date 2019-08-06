package amino

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUvarintSize(t *testing.T) {
	tests := []struct {
		name string
		u    uint64
		want int
	}{
		{"0 bit", 0, 1},
		{"1 bit", 1 << 0, 1},
		{"6 bits", 1 << 5, 1},
		{"7 bits", 1 << 6, 1},
		{"8 bits", 1 << 7, 2},
		{"62 bits", 1 << 61, 9},
		{"63 bits", 1 << 62, 9},
		{"64 bits", 1 << 63, 10},
	}
	for i, testcase := range tests {
		t.Run(testcase.name, func(t *testing.T) {
			require.Equal(t, testcase.want, UvarintSize(testcase.u), "failed on tc %d", i) // nolint:scopelint
		})
	}
}
