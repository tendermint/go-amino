package wire

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMarshal(t *testing.T) {
	s := SimpleStruct{
		String: "hello",
		Bytes:  []byte("goodbye"),
		Time:   time.Now().Truncate(time.Millisecond),
	}

	b, err := Marshal(s)
	assert.Nil(t, err)

	var s2 SimpleStruct
	err = Unmarshal(b, &s2)
	assert.Nil(t, err)
	assert.Equal(t, s, s2)

}
