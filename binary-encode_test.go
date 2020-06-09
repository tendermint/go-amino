package amino

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeFieldNumberAndTyp3(t *testing.T) {
	buf := new(bytes.Buffer)
	err := encodeFieldNumberAndTyp3(buf, 1, Typ3ByteLength)
	assert.Nil(t, err)
	assert.Equal(t, []byte{0x00}, buf.Bytes()) // XXX This should fail
}
