package wire

// Add tests for internal utils and helpers here.

import (
	"bytes"
	"testing"
)

func TestTrimNullPrefixBytes(t *testing.T) {
	tests := []struct {
		in   []byte
		want []byte
	}{
		{in: []byte{0x00, 0x00, 0x00, 0x00}, want: []byte{}},
		{in: []byte{0x00, 0x00, 0x01, 0x00}, want: []byte{0x01, 0x00}},
		{in: []byte{0x10, 0x00, 0x01, 0x00}, want: []byte{0x10, 0x00, 0x01, 0x00}},
	}

	for i, tt := range tests {
		if got, want := trimNullPrefixBytes(tt.in), tt.want; !bytes.Equal(got, want) {
			t.Errorf("#%d: got=(% X) want=(% X)", i, got, want)
		}
	}
}
