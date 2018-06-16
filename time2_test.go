package amino

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestDecodeTime(t *testing.T) {

	type T struct {
		Time time.Time
	}

	tm, err := time.Parse("2006-01-02 15:04:05 +0000 UTC", "1970-01-01 00:00:00 +0000 UTC")
	assert.NoError(t, err)
	cdc := NewCodec()
	b, err := cdc.MarshalBinary(T{Time: tm})
	fmt.Println(b)
	assert.NoError(t, err)
	var ti T
	err = cdc.UnmarshalBinary(b, &ti)
	assert.NoError(t, err)
	assert.Equal(t, T{Time: tm}, ti)
	// Time:(time.Time)1970-01-01 00:00:01.978131102 +0000 UTC}
	// 1970-01-01 00:00:01.978131102 +0000 UTC
}
