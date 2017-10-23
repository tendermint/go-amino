package wire

import (
	"io"
	"time"

	cmn "github.com/tendermint/tmlibs/common"
)

/*
Writes nanoseconds since epoch but with millisecond precision.
This is to ease compatibility with Javascript etc.
*/

// WriteTime writes the number of nanoseconds since
// January 1, 1970 UTC to the Writer as an Int64.
// If the given time is less than January 1, 1970 UTC, a -1 is written.
func WriteTime(t time.Time, w io.Writer, n *int, err *error) {
	nanosecs := t.UnixNano()
	millisecs := nanosecs / 1000000
	if nanosecs < 0 {
		WriteInt64(-1, w, n, err)
	} else {
		WriteInt64(millisecs*1000000, w, n, err)
	}
}

// ReadTime reads an Int64 from the Reader, interprets it as
// the number of nanoseconds since January 1, 1970 UTC, and
// returns the corresponding time. If the Int64 read is -1, it returns
// the zero value for time.Time.
func ReadTime(r io.Reader, n *int, err *error) time.Time {
	t := ReadInt64(r, n, err)
	if t == -1 {
		return time.Time{}
	}
	if t%1000000 != 0 {
		cmn.PanicSanity("Time cannot have sub-millisecond precision")
	}
	return time.Unix(0, t)
}
