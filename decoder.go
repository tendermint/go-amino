package wire

import (
	"encoding/binary"
	"fmt"
	"math"
	"time"
)

//----------------------------------------
// Signed

func DecodeInt8(bz []byte) (i int8, n int, err error) {
	const size int = 1
	if len(bz) < size {
		return
	}
	i = int8(bz[0])
	n = size
	return
}

func DecodeInt16(bz []byte) (i int16, n int, err error) {
	const size int = 2
	if len(bz) < size {
		return
	}
	i = int16(binary.BigEndian.Uint16(bz[:size]))
	n = size
	return
}

func DecodeInt32(bz []byte) (i int32, n int, err error) {
	const size int = 4
	if len(bz) < size {
		return
	}
	i = int32(binary.BigEndian.Uint32(bz[:size]))
	n = size
	return
}

func DecodeInt64(bz []byte) (i int64, n int, err error) {
	const size int = 8
	if len(bz) < size {
		return
	}
	i = int64(binary.BigEndian.Uint64(bz[:size]))
	n = size
	return
}

func DecodeVarint(bz []byte) (i int64, n int, err error) {
	i, n = binary.Varint(bz)
	if n == 0 {
		err = fmt.Errorf("eof decoding varint")
	}
	return
}

//----------------------------------------
// Unsigned

func DecodeByte(bz []byte) (b byte, n int, err error) {
	const size int = 1
	if len(bz) < size {
		return
	}
	b = bz[0]
	n = size
	return
}

func DecodeUint8(bz []byte) (i uint8, n int, err error) {
	const size int = 1
	if len(bz) < size {
		return
	}
	i = uint8(bz[0])
	n = size
	return
}
func DecodeUint16(bz []byte) (i uint16, n int, err error) {
	const size int = 2
	if len(bz) < size {
		return
	}
	i = binary.BigEndian.Uint16(bz[:size])
	n = size
	return
}

func DecodeUint32(bz []byte) (i uint32, n int, err error) {
	const size int = 4
	if len(bz) < size {
		return
	}
	i = binary.BigEndian.Uint32(bz[:size])
	n = size
	return
}

func DecodeUint64(bz []byte) (i uint64, n int, err error) {
	const size int = 8
	if len(bz) < size {
		return
	}
	i = binary.BigEndian.Uint64(bz[:size])
	n = size
	return
}

func DecodeUvarint(bz []byte) (i uint64, n int, err error) {
	i, n = binary.Uvarint(bz)
	if n == 0 {
		err = fmt.Errorf("eof decoding uvarint")
	}
	return
}

//----------------------------------------
// Other

func DecodeBool(bz []byte) (b bool, n int, err error) {
	const size int = 1
	if len(bz) < size {
		return
	}
	switch bz[0] {
	case 0:
		b = false
	case 1:
		b = true
	default:
		err = fmt.Errorf("invalid bool")
	}
	n = size
	return
}

// NOTE: UNSAFE
func DecodeFloat32(bz []byte) (f float32, n int, err error) {
	const size int = 4
	if len(bz) < size {
		return
	}
	i := uint32(binary.BigEndian.Uint32(bz[:size]))
	f = math.Float32frombits(i)
	n = size
	return
}

// NOTE: UNSAFE
func DecodeFloat64(bz []byte) (f float64, n int, err error) {
	const size int = 8
	if len(bz) < size {
		return
	}
	i := uint64(binary.BigEndian.Uint64(bz[:size]))
	f = math.Float64frombits(i)
	n = size
	return
}

// DecodeTime decodes a Int64 and interprets it as the
// number of nanoseconds since January 1, 1970 UTC, and
// returns the corresponding time. If the Int64 read is
// less than zero, or not a multiple of a million, it sets
// the error and returns the default time.
func DecodeTime(bz []byte) (t time.Time, n int, err error) {
	i, n, err := DecodeInt64(bz)
	if i%1000000 != 0 {
		err = fmt.Errorf("submillisecond precision not supported")
		return
	}
	t = time.Unix(0, i)
	return
}

func DecodeByteSlice(bz []byte) (bz2 []byte, n int, err error) {
	var count int64
	var _n int
	count, _n, err = DecodeVarint(bz)
	if slide(bz, &bz, &n, _n) && err != nil {
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
