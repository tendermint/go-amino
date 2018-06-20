package amino

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDecodeSkippedFieldsInTime(t *testing.T) {
	type testTime struct {
		Time time.Time
	}
	cdc := NewCodec()

	tm, err := time.Parse("2006-01-02 15:04:05 +0000 UTC", "1970-01-01 00:00:00 +0000 UTC")
	assert.NoError(t, err)

	b, err := cdc.MarshalBinary(testTime{Time: tm})
	assert.NoError(t, err)
	var ti testTime
	err = cdc.UnmarshalBinary(b, &ti)
	assert.NoError(t, err)
	assert.Equal(t, testTime{Time: tm}, ti)

	tm2, err := time.Parse("2006-01-02 15:04:05 +0000 UTC", "1970-01-01 00:00:01.978131102 +0000 UTC")
	assert.NoError(t, err)

	b, err = cdc.MarshalBinary(testTime{Time: tm2})
	assert.NoError(t, err)
	err = cdc.UnmarshalBinary(b, &ti)
	assert.NoError(t, err)
	assert.Equal(t, testTime{Time: tm2}, ti)
}
