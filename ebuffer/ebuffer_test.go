package ebuffer

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEBufferBasic(t *testing.T) {
	// NOTE: the max cap must be sufficiently large.
	for cap := 0; cap < 20; cap++ {
		t.Run(fmt.Sprintf("cap-%v", cap), func(t *testing.T) {

			ebuf := NewEBuffer(0)
			assert.Equal(t, ebuf.Len(), 0)
			ebuf.Append([]byte(""))
			assert.Equal(t, ebuf.Len(), 0)
			ebuf.Append([]byte(nil))
			assert.Equal(t, ebuf.Len(), 0)
			ebuf.Append([]byte("abc"))
			assert.Equal(t, ebuf.Len(), 3)
			res := ebuf.Reserve(10)
			assert.Equal(t, ebuf.Len(), 3)
			ebuf.Append([]byte("ghi"))
			assert.Equal(t, ebuf.Len(), 6)
			err := ebuf.Edit(res, []byte("def"))
			assert.Nil(t, err)
			assert.Equal(t, ebuf.Len(), 9)

			buf := ebuf.Compact()
			assert.Equal(t, []byte("abcdefghi"), buf)
		})
	}
}

func TestEBufferRes(t *testing.T) {
	// NOTE: the max cap must be sufficiently large.
	for cap := 0; cap < 20; cap++ {
		t.Run(fmt.Sprintf("cap-%v", cap), func(t *testing.T) {
			ebuf := NewEBuffer(0)
			_ = ebuf.Reserve(10) // not used
			_ = ebuf.Reserve(10) // not used
			ebuf.Append([]byte(""))
			_ = ebuf.Reserve(10) // not used
			_ = ebuf.Reserve(10) // not used
			ebuf.Append([]byte(nil))
			_ = ebuf.Reserve(10) // not used
			_ = ebuf.Reserve(10) // not used
			ebuf.Append([]byte("abc"))
			res := ebuf.Reserve(10)
			assert.Equal(t, ebuf.Len(), 3)
			ebuf.Append([]byte("ghi"))
			assert.Equal(t, ebuf.Len(), 6)
			err := ebuf.Edit(res, []byte("def0123456789"))
			assert.NotNil(t, err) // exceeded

			assert.Equal(t, ebuf.Len(), 6)
			buf := ebuf.Compact()
			assert.Equal(t, []byte("abcghi"), buf)

			err = ebuf.Edit(res, []byte("def"))
			assert.Nil(t, err) // good
			assert.Equal(t, ebuf.Len(), 9)

			err = ebuf.Edit(res, []byte("DEF"))
			assert.NotNil(t, err) // already used
			assert.Equal(t, ebuf.Len(), 9)

			buf = ebuf.Compact()
			assert.Equal(t, []byte("abcdefghi"), buf)
		})
	}
}
