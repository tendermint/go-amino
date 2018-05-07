package amino

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
	var i64 = int64(0)
	i64, n, err = DecodeVarint(bz)
	if err != nil {
		return
	}
	if i64 > int64(math.MaxInt8) {
		err = errors.New("EOF decoding int8")
		return
	}
	i = int8(i64)
	return
}

func DecodeInt16(bz []byte) (i int16, n int, err error) {
	var i64 = int64(0)
	i64, n, err = DecodeVarint(bz)
	if err != nil {
		return
	}
	if i64 > int64(math.MaxInt16) {
		err = errors.New("EOF decoding int16")
		return
	}
	i = int16(i64)
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
	return DecodeUint8(bz)
}

func DecodeUint8(bz []byte) (u uint8, n int, err error) {
	var u64 = uint64(0)
	u64, n, err = DecodeUvarint(bz)
	if err != nil {
		return
	}
	if u64 > uint64(math.MaxUint8) {
		err = errors.New("EOF decoding uint8")
		return
	}
	u = uint8(u64)
	return
}
func DecodeUint16(bz []byte) (u uint16, n int, err error) {
	var u64 = uint64(0)
	u64, n, err = DecodeUvarint(bz)
	if err != nil {
		return
	}
	if u64 > uint64(math.MaxUint16) {
		err = errors.New("EOF decoding uint16")
		return
	}
	u = uint16(u64)
	return
}

func DecodeUint32(bz []byte) (u uint32, n int, err error) {
	const size int = 4
	if len(bz) < size {
		err = errors.New("EOF decoding uint32")
		return
	}
	u = binary.BigEndian.Uint32(bz[:size])
	n = size
	return
}

func DecodeUint64(bz []byte) (u uint64, n int, err error) {
	const size int = 8
	if len(bz) < size {
		err = errors.New("EOF decoding uint64")
		return
	}
	u = binary.BigEndian.Uint64(bz[:size])
	n = size
	return
}

func DecodeUvarint(bz []byte) (u uint64, n int, err error) {
	u, n = binary.Uvarint(bz)
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

// DecodeTime decodes milliseconds (int64) since January 1, 1970 UTC,
// and returns the corresponding time.
// TODO return error if behavior is undefined.
func DecodeTime(bz []byte) (t time.Time, n int, err error) {
	var millis, _n = int64(0), int(0)
	millis, _n, err = DecodeInt64(bz)
	if slide(&bz, &n, _n) && err != nil {
		return
	}

	// Construct time.
	var seconds = millis / 1e3
	var nanos = (millis % 1000) * 1e6
	t = time.Unix(seconds, nanos)
	// Strip timezone and monotonic for deep equality.
	t = t.UTC().Truncate(0)
	return
}

func DecodeByteSlice(bz []byte) (bz2 []byte, n int, err error) {
	var count uint64
	var _n int
	count, _n, err = DecodeUvarint(bz)
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
