package ebuffer

import (
	"fmt"
)

type reservation struct {
	pos  int
	len  int
	used int
}

func (res *reservation) Copy() *reservation {
	return &reservation{
		pos:  res.pos,
		len:  res.len,
		used: res.used,
	}
}

func (res *reservation) MarkUsed(num int) {
	left := res.len - res.used
	if left < num {
		panic(fmt.Sprintf("Cannot mark %v reserved bytes as used, only %v left", num, left))
	} else {
		res.used += num
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

func (eb *EBuffer) Copy() *EBuffer {
	buf := make([]byte, len(eb.buf))
	copy(buf, eb.buf)
	resz := make([]*reservation, len(eb.resz))
	for i, res := range eb.resz {
		resz[i] = res.Copy()
	}
	return &EBuffer{
		buf:  buf,
		resz: resz,
		len:  eb.len,
	}
}

func (eb *EBuffer) Len() int {
	return eb.len
}

func (eb *EBuffer) Truncate(newlen int) {
	if newlen > eb.len {
		panic("Cannot truncate to a greater length")
	} else if newlen == eb.len {
		return
	} else if newlen < 0 {
		panic("Cannot truncate to a negative length")
	}
	delete := eb.len - newlen
	for i := len(eb.resz) - 1; i >= 0; i-- {
		res := eb.resz[i]
		rest := len(eb.buf) - (res.pos + res.len) // after res
		if delete <= rest {
			// NOTE: Leave trailing unused reservations.
			break
		} else if (delete - rest) < res.used {
			// NOTE: No reason to modify the unused res bytes
			eb.buf = eb.buf[:len(eb.buf)-rest]
			res.used -= (delete - rest)
			eb.len -= delete
			return
		} else {
			// NOTE: If we are deleting a used reservation
			// fully, also discard it.
			eb.buf = eb.buf[:len(eb.buf)-(rest+res.len)]
			eb.len -= (rest + res.used)
			delete -= (rest + res.used)
			eb.resz = eb.resz[:i]
			res.used = -1 // discard
			continue
		}
	}
	eb.buf = eb.buf[:len(eb.buf)-delete]
	eb.len -= delete
	return
}

func (eb *EBuffer) Append(bz []byte) {
	eb.buf = append(eb.buf, bz...)
	eb.len += len(bz)
}

// CONTRACT: res is in eb.resz.
func (eb *EBuffer) Edit(res *reservation, bz []byte) {
	if res.used != 0 {
		// NOTE: Do not change this behavior, it's very
		// tricky to get this right w/ truncation, and even
		// if done right, usage would be difficult to
		// understand.
		// NOTE: discarded reservations have used=-1
		panic("EBuffer reservation already edited")
	}
	if len(bz) > res.len {
		panic(fmt.Sprintf("EBuffer edit (%v) exceeds reservation length %v",
			len(bz), res.len))
	}
	copy(eb.buf[res.pos:], bz) // Left adjusted.
	res.MarkUsed(len(bz))
	eb.len += len(bz)
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
		newbuf = append(newbuf, eb.buf[cur:res.pos+res.used]...) // Left adjusted
		cur = res.pos + res.len                                  // Left adjusted
	}
	newbuf = append(newbuf, eb.buf[cur:]...)

	return newbuf
}
