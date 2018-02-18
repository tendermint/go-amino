package wire

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"time"
)

//----------------------------------------
// Signed

func DecodeInt8(bz []byte) (i int8, n int, err error) {
	const size int = 1
	if len(bz) < size {
		err = errors.New("EOF decoding int8")
		return
	}
	i = int8(bz[0])
	n = size
	return
}

func DecodeInt16(bz []byte) (i int16, n int, err error) {
	const size int = 2
	if len(bz) < size {
		err = errors.New("EOF decoding int16")
		return
	}
	i = int16(binary.BigEndian.Uint16(bz[:size]))
	n = size
	return
}

func DecodeInt32(bz []byte) (i int32, n int, err error) {
	const size int = 4
	if len(bz) < size {
		err = errors.New("EOF decoding int32")
		return
	}
	i = int32(binary.BigEndian.Uint32(bz[:size]))
	n = size
	return
}

func DecodeInt64(bz []byte) (i int64, n int, err error) {
	const size int = 8
	if len(bz) < size {
		err = errors.New("EOF decoding int64")
		return
	}
	i = int64(binary.BigEndian.Uint64(bz[:size]))
	n = size
	return
}

func DecodeVarint(bz []byte) (i int64, n int, err error) {
	i, n = binary.Varint(bz)
	if n < 0 {
		n = 0
		err = errors.New("EOF decoding varint")
	}
	return
}

//----------------------------------------
// Unsigned

func DecodeByte(bz []byte) (b byte, n int, err error) {
	const size int = 1
	if len(bz) < size {
		err = errors.New("EOF decoding byte")
		return
	}
	b = bz[0]
	n = size
	return
}

func DecodeUint8(bz []byte) (i uint8, n int, err error) {
	const size int = 1
	if len(bz) < size {
		err = errors.New("EOF decoding uint8")
		return
	}
	i = bz[0]
	n = size
	return
}
func DecodeUint16(bz []byte) (i uint16, n int, err error) {
	const size int = 2
	if len(bz) < size {
		err = errors.New("EOF decoding uint16")
		return
	}
	i = binary.BigEndian.Uint16(bz[:size])
	n = size
	return
}

func DecodeUint32(bz []byte) (i uint32, n int, err error) {
	const size int = 4
	if len(bz) < size {
		err = errors.New("EOF decoding uint32")
		return
	}
	i = binary.BigEndian.Uint32(bz[:size])
	n = size
	return
}

func DecodeUint64(bz []byte) (i uint64, n int, err error) {
	const size int = 8
	if len(bz) < size {
		err = errors.New("EOF decoding uint64")
		return
	}
	i = binary.BigEndian.Uint64(bz[:size])
	n = size
	return
}

func DecodeUvarint(bz []byte) (i uint64, n int, err error) {
	i, n = binary.Uvarint(bz)
	if n <= 0 {
		n = 0
		err = errors.New("EOF decoding uvarint")
	}
	return
}

//----------------------------------------
// Other

func DecodeBool(bz []byte) (b bool, n int, err error) {
	const size int = 1
	if len(bz) < size {
		err = errors.New("EOF decoding bool")
		return
	}
	switch bz[0] {
	case 0:
		b = false
	case 1:
		b = true
	default:
		err = errors.New("invalid bool")
	}
	n = size
	return
}

// NOTE: UNSAFE
func DecodeFloat32(bz []byte) (f float32, n int, err error) {
	const size int = 4
	if len(bz) < size {
		err = errors.New("EOF decoding float32")
		return
	}
	i := binary.BigEndian.Uint32(bz[:size])
	f = math.Float32frombits(i)
	n = size
	return
}

// NOTE: UNSAFE
func DecodeFloat64(bz []byte) (f float64, n int, err error) {
	const size int = 8
	if len(bz) < size {
		err = errors.New("EOF decoding float64")
		return
	}
	i := binary.BigEndian.Uint64(bz[:size])
	f = math.Float64frombits(i)
	n = size
	return
}

// DecodeTime decodes seconds (int64) and nanoseconds (int32) since January 1,
// 1970 UTC, and returns the corresponding time.  If nanoseconds is not in the
// range [0, 999999999], or if seconds is too large, the behavior is
// undefined.
// TODO return errro if behavior is undefined.
func DecodeTime(bz []byte) (t time.Time, n int, err error) {
	s, _n, err := DecodeInt64(bz)
	if slide(&bz, &n, _n) && err != nil {
		return
	}
	ns, _n, err := DecodeInt32(bz)
	if slide(&bz, &n, _n) && err != nil {
		return
	}
	if ns < 0 || 999999999 < ns {
		err = fmt.Errorf("Invalid time, nanoseconds out of bounds %v", ns)
		return
	}
	t = time.Unix(s, int64(ns))
	// strip timezone and monotonic for deep equality
	t = t.UTC().Truncate(0)
	return
}

func DecodeByteSlice(bz []byte) (bz2 []byte, n int, err error) {
	var count int64
	var _n int
	count, _n, err = DecodeVarint(bz)
	if slide(&bz, &n, _n) && err != nil {
		return
	}
	if int(count) < 0 {
		err = fmt.Errorf("invalid negative length %v decoding []byte", count)
		return
	}
	if len(bz) < int(count) {
		err = fmt.Errorf("insufficient bytes decoding []byte of length %v", count)
		return
	}
	bz2 = make([]byte, count)
	copy(bz2, bz[0:count])
	n += int(count)
	return
}

func DecodeString(bz []byte) (s string, n int, err error) {
	var bz2 []byte
	bz2, n, err = DecodeByteSlice(bz)
	s = string(bz2)
	return
}
