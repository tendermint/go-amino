package ebuffer

import (
	"errors"
	"fmt"
)

type reservation struct {
	pos  int
	len  int
	used int
}

func (res *reservation) MarkUsed(num int) error {
	left := res.len - res.used
	if left < num {
		return fmt.Errorf("Cannot mark %v reserved bytes as used, only %v left", num, left)
	} else {
		res.used += num
		return nil
	}
}

type EBuffer struct {
	buf  []byte
	resz []*reservation // byte reservation (holes) ordered by pos. 0 length reservations are valid.
	len  int            // total non-hole bytes written to buf so far.
}

func NewEBuffer(cap int) *EBuffer {
	return &EBuffer{
		buf:  make([]byte, 0, cap),
		resz: make([]*reservation, 0, 10), // TODO make constant
		len:  0,
	}
}

func (eb *EBuffer) Len() int {
	return eb.len
}

func (eb *EBuffer) Append(bz []byte) {
	eb.buf = append(eb.buf, bz...)
	eb.len += len(bz)
}

// CONTRACT: res is in eb.resz.
func (eb *EBuffer) Edit(res *reservation, bz []byte) error {
	if res.used != 0 {
		return errors.New("EBuffer reservation already edited")
	}
	if len(bz) > res.len {
		return fmt.Errorf("EBuffer edit (%v) exceeds reservation length %v",
			len(bz), res.len)
	}
	copy(eb.buf[res.pos+(res.len-len(bz)):], bz) // Right adjusted.
	res.MarkUsed(len(bz))
	eb.len += len(bz)
	return nil
}

func (eb *EBuffer) Reserve(num int) *reservation {
	pos := len(eb.buf)
	res := &reservation{
		pos: pos,
		len: num,
	}
	eb.resz = append(eb.resz, res)
	eb.buf = append(eb.buf, make([]byte, num)...)
	return res
}

func (eb *EBuffer) Compact() []byte {
	holes := 0
	for _, res := range eb.resz {
		holes += (res.len - res.used)
	}

	newbuf := make([]byte, 0, len(eb.buf)-holes) // New buffer to return
	cur := 0                                     // Cursor position on eb.buf

	for _, res := range eb.resz {
		if (res.len - res.used) == 0 {
			continue
		}
		newbuf = append(newbuf, eb.buf[cur:res.pos]...)
		cur = res.pos + (res.len - res.used) // Reserved bytes are written right-adjusted.
	}
	newbuf = append(newbuf, eb.buf[cur:]...)

	return newbuf
}
