package wire

import (
	"encoding/binary"
	"errors"
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

//----------------------------------------
// (u)varints

func DecodeUvarint(bz []byte) (i uint64, n int, err error) {
	i, n = binary.Uvarint(bz)
	if n == 0 {
		err = fmt.Errorf("eof decoding uvarint")
	}
	return
}

func DecodeVarint(bz []byte) (i int, n int, err error) {
	i, n = binary.Varint(bz)
	if n == 0 {
		err = fmt.Errorf("eof decoding varint")
	}
	return
}

//----------------------------------------
// Misc.

func DecodeBool(bz []byte) (b bool, n int, err error) {
	const size int = 1
	if len(bz) < size {
		return
	}
	switch bz[0] {
	case 0:
		n = size
	case 1:
		n = size
		b = true
	default:
		err = fmt.Errorf("invalid bool")
	}
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

func DecodeTime(bz []byte) (t time.Time, n int, err error) {
	i, n, err := DecodeInt64(bz)
	if i < 0 {
		err = fmt.Errorf("DecodeTime: negative time")
		return
	}
	if i%1000000 != 0 {
		err = fmt.Errorf("submillisecond precision not supported")
		return
	}
	return time.Unix(0, t)
}

func DecodeByteSlice(bz []byte) (bz2 []byte, n int, err error) {
	var count int64
	var _n int
	count, _n, err = DecodeVarint(bz)
	incrSlice(bz, &bz, &n, _n)
	if err != nil {
		return
	}
	if len(bz) < count {
		err = fmt.Errorf("insufficient bytes decoding []byte of length %v", count)
		return
	}
	bz2 = make([]byte, count)
	copy(bz2, bz[0:count])
	n += count
	return
}

//----------------------------------------

// CONTRACT: by the time this is called, len(bz) >= _n
// Returns true so you can write one-liners.
func incrSlice(bz []byte, bz2 *[]byte, n *int, _n int) bool {
	*bz2 = bz[_n:]
	*n += _n
	return true
}
