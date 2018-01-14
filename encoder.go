package wire

import (
	"encoding/binary"
	"io"
	"math"
	"time"
)

//----------------------------------------
// Signed

func EncodeInt8(w io.Writer, i int8) (err error) {
	return EncodeByte(w, byte(i))
}

func EncodeInt16(w io.Writer, i int16) (err error) {
	var buf [2]byte
	binary.BigEndian.PutUint16(buf[:], uint16(i))
	_, err = w.Write(buf[:])
	return
}

func EncodeInt32(w io.Writer, i int32) (err error) {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], uint32(i))
	_, err = w.Write(buf[:])
	return
}

func EncodeInt64(w io.Writer, i int64) (err error) {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(i))
	_, err = w.Write(buf[:])
	return err
}

func EncodeVarint(w io.Writer, i int64) (err error) {
	var buf [10]byte
	n := binary.PutVarint(buf[:], i)
	_, err = w.Write(buf[0:n])
	return
}

func VarintSize(i int64) int {
	var buf [10]byte
	n := binary.PutVarint(buf[:], i)
	return n
}

//----------------------------------------
// Unsigned

func EncodeByte(w io.Writer, b byte) (err error) {
	_, err = w.Write([]byte{b})
	return
}

func EncodeUint8(w io.Writer, i uint8) (err error) {
	return EncodeByte(w, i)
}

func EncodeUint16(w io.Writer, i uint16) (err error) {
	var buf [2]byte
	binary.BigEndian.PutUint16(buf[:], i)
	_, err = w.Write(buf[:])
	return
}

func EncodeUint32(w io.Writer, i uint32) (err error) {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], i)
	_, err = w.Write(buf[:])
	return
}

func EncodeUint64(w io.Writer, i uint64) (err error) {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], i)
	_, err = w.Write(buf[:])
	return
}

func EncodeUvarint(w io.Writer, i uint64) (err error) {
	var buf [10]byte
	n := binary.PutUvarint(buf[:], i)
	_, err = w.Write(buf[0:n])
	return
}

func UvarintSize(i uint64) int {
	var buf [10]byte
	n := binary.PutUvarint(buf[:], i)
	return n
}

//----------------------------------------
// Other

func EncodeBool(w io.Writer, b bool) (err error) {
	if b {
		err = EncodeUint8(w, uint8(1))
	} else {
		err = EncodeUint8(w, uint8(0))
	}
	return
}

// NOTE: UNSAFE
func EncodeFloat32(w io.Writer, f float32) (err error) {
	return EncodeUint32(w, math.Float32bits(f))
}

// NOTE: UNSAFE
func EncodeFloat64(w io.Writer, f float64) (err error) {
	return EncodeUint64(w, math.Float64bits(f))
}

// EncodeTime writes the number of seconds (int64) and nanoseconds (int32),
// with millisecond resolution since January 1, 1970 UTC to the Writer as an
// Int64.
// Milliseconds are used to ease compatibility with Javascript,
// which does not support finer resolution.
func EncodeTime(w io.Writer, t time.Time) (err error) {
	s := t.Unix()
	ns := int32(t.Nanosecond()) // this int64 -> int32 is safe.
	err = EncodeInt64(w, s)
	if err != nil {
		return err
	}
	err = EncodeInt32(w, ns)
	if err != nil {
		return err
	}
	return
}

func EncodeByteSlice(w io.Writer, bz []byte) (err error) {
	err = EncodeVarint(w, int64(len(bz)))
	if err != nil {
		return
	}
	_, err = w.Write(bz)
	return
}

func ByteSliceSize(bz []byte) int {
	return UvarintSize(uint64(len(bz))) + len(bz)
}

func EncodeString(w io.Writer, s string) (err error) {
	return EncodeByteSlice(w, []byte(s))
}
