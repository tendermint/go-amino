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

			ebuf := NewEBuffer(cap)
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
			ebuf.Edit(res, []byte("def"))
			assert.Equal(t, ebuf.Len(), 9)
			assert.Panics(t, func() { ebuf.Edit(res, []byte("DEF")) })

			buf := ebuf.Compact()
			assert.Equal(t, []byte("abcdefghi"), buf)

			testEBufferTruncateCompact(t, ebuf)
		})
	}
}

func TestEBufferRes(t *testing.T) {
	// NOTE: the max cap must be sufficiently large.
	for cap := 0; cap < 20; cap++ {
		t.Run(fmt.Sprintf("cap-%v", cap), func(t *testing.T) {
			ebuf := NewEBuffer(cap)
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
			assert.Panics(t, func() { ebuf.Edit(res, []byte("def0123456789")) }) // exceeded

			assert.Equal(t, ebuf.Len(), 6)
			buf := ebuf.Compact()
			assert.Equal(t, []byte("abcghi"), buf)

			ebuf.Edit(res, []byte("def"))
			assert.Equal(t, ebuf.Len(), 9)
			assert.Panics(t, func() { ebuf.Edit(res, []byte("DEF")) }) // already used

			buf = ebuf.Compact()
			assert.Equal(t, []byte("abcdefghi"), buf)

			testEBufferTruncateCompact(t, ebuf)
		})
	}
}

func TestEBufferTruncate1(t *testing.T) {
	ebuf := NewEBuffer(0)
	ebuf.Append([]byte("abc"))
	_ = ebuf.Reserve(10)
	ebuf.Append([]byte("def"))
	assert.Equal(t, ebuf.Len(), 6)
	assert.Equal(t, string(ebuf.Compact()), "abcdef")

	testEBufferTruncateCompact(t, ebuf)
}

func TestEBufferTruncate2(t *testing.T) {
	ebuf := NewEBuffer(0)
	ebuf.Append([]byte("abc"))
	res1 := ebuf.Reserve(10)
	ebuf.Append([]byte("def"))
	assert.Equal(t, ebuf.Len(), 6)
	assert.Equal(t, string(ebuf.Compact()), "abcdef")
	ebuf.Truncate(6)
	assert.Equal(t, ebuf.Len(), 6)
	assert.Equal(t, string(ebuf.Compact()), "abcdef")
	ebuf.Edit(res1, []byte("00"))
	assert.Equal(t, ebuf.Len(), 8)
	assert.Equal(t, string(ebuf.Compact()), "abc00def")

	testEBufferTruncateCompact(t, ebuf)
}

func TestEBufferTruncate3(t *testing.T) {
	ebuf := NewEBuffer(0)
	ebuf.Append([]byte("abc"))
	res1 := ebuf.Reserve(10)
	ebuf.Append([]byte("def"))
	assert.Equal(t, ebuf.Len(), 6)
	assert.Equal(t, string(ebuf.Compact()), "abcdef")
	ebuf.Truncate(6)
	assert.Equal(t, ebuf.Len(), 6)
	assert.Equal(t, string(ebuf.Compact()), "abcdef")
	ebuf.Edit(res1, []byte("00"))
	assert.Equal(t, ebuf.Len(), 8)
	assert.Equal(t, string(ebuf.Compact()), "abc00def")

	testEBufferTruncateCompact(t, ebuf)
}

func TestEBufferTruncate4(t *testing.T) {
	ebuf := NewEBuffer(0)
	ebuf.Append([]byte("abc"))
	_ = ebuf.Reserve(10) // not used
	res1 := ebuf.Reserve(10)
	_ = ebuf.Reserve(10) // not used
	ebuf.Append([]byte("def"))
	assert.Equal(t, ebuf.Len(), 6)
	assert.Equal(t, string(ebuf.Compact()), "abcdef")
	ebuf.Truncate(6)
	assert.Equal(t, ebuf.Len(), 6)
	assert.Equal(t, string(ebuf.Compact()), "abcdef")
	ebuf.Edit(res1, []byte("00"))
	assert.Equal(t, ebuf.Len(), 8)
	assert.Equal(t, string(ebuf.Compact()), "abc00def")

	testEBufferTruncateCompact(t, ebuf)
}

func testEBufferTruncateCompact(t *testing.T, ebuf *EBuffer) {
	length := ebuf.Len()
	val := string(ebuf.Compact())
	assert.Equal(t, length, len(val))

	// Try compacting to all lengths
	for i := 0; i < length; i++ {
		ebuf2 := ebuf.Copy()
		ebuf2.Truncate(i)
		val2 := string(ebuf2.Compact())
		assert.Equal(t, val2, val[:i])
	}

	// Try compacting to bad lengths
	ebuf3 := ebuf.Copy()
	assert.Panics(t, func() { ebuf3.Truncate(length + 1) })
	ebuf4 := ebuf.Copy()
	assert.Panics(t, func() { ebuf4.Truncate(-1) })
}
