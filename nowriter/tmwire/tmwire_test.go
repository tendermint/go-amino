package tmwire

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/go-wire/nowriter/tmdecoding"
	"github.com/tendermint/go-wire/nowriter/tmencoding"
	"github.com/tendermint/go-wire/nowriter/tmlegacy"
	"testing"
)

var legacy = tmlegacy.TMEncoderLegacy{}
var pure = tmencoding.TMEncoderPure{}
var dec = tmdecoding.TMDecoderPure{}

func TestByte(t *testing.T) {
	for i := 0; i < 256; i += 1 {
		x0 := byte(i)
		buf1 := new(bytes.Buffer)
		n1, err1 := new(int), new(error)
		legacy.WriteOctet(x0, buf1, n1, err1)
		b1 := buf1.Bytes()
		b2 := pure.EncodeOctet(x0)
		if !bytes.Equal(b1, b2) {
			t.Fatalf("Bytes do not match for %#v and %#v", b1, b2)
		}
		b3, n3, err3 := dec.DecodeOctet(b1)
		if b3 != x0 {
			t.Fatalf("Decoded bytes do not match for %#v and %#v", b3, x0)
		}
		if n3 != 1 {
			t.Fatalf("Decoded byte count is not 1")
		}
		assert.Nil(t, err3)
	}
}
