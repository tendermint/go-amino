package amino

import (
	"encoding/binary"
	"io"
	"math"
	"time"
)

//----------------------------------------
// Signed

func EncodeInt8(w io.Writer, i int8) (err error) {
	return EncodeVarint(w, uint64(i))
}

func EncodeInt16(w io.Writer, i int16) (err error) {
	return EncodeVarint(w, uint64(i))
}

func EncodeInt32(w io.Writer, i int32) (err error) {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], uint32(i))
	_, err = w.Write(buf[:])
	return
}

func EncodeInt64(w io.Writer, i int64) (err error) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(i))
	_, err = w.Write(buf[:])
	return err
}

func EncodeVarint(w io.Writer, v uint64) (err error) {
	// TODO(ismail): this is copy & pasted (slightly modified) from:
	// https://github.com/golang/protobuf/blob/3a3da3a4e26776cc22a79ef46d5d58477532dede/proto/table_marshal.go#L1285-L1366
	// find out why it is inlined like this, if we could use the binary package instead, or, clarify copyright
	// if we want to keep it like this:
	var b []byte
	// TODO: make 1-byte (maybe 2-byte) case inline-able, once we
	// have non-leaf inliner.
	switch {
	case v < 1<<7:
		b = append(b, byte(v))
	case v < 1<<14:
		b = append(b,
			byte(v&0x7f|0x80),
			byte(v>>7))
	case v < 1<<21:
		b = append(b,
			byte(v&0x7f|0x80),
			byte((v>>7)&0x7f|0x80),
			byte(v>>14))
	case v < 1<<28:
		b = append(b,
			byte(v&0x7f|0x80),
			byte((v>>7)&0x7f|0x80),
			byte((v>>14)&0x7f|0x80),
			byte(v>>21))
	case v < 1<<35:
		b = append(b,
			byte(v&0x7f|0x80),
			byte((v>>7)&0x7f|0x80),
			byte((v>>14)&0x7f|0x80),
			byte((v>>21)&0x7f|0x80),
			byte(v>>28))
	case v < 1<<42:
		b = append(b,
			byte(v&0x7f|0x80),
			byte((v>>7)&0x7f|0x80),
			byte((v>>14)&0x7f|0x80),
			byte((v>>21)&0x7f|0x80),
			byte((v>>28)&0x7f|0x80),
			byte(v>>35))
	case v < 1<<49:
		b = append(b,
			byte(v&0x7f|0x80),
			byte((v>>7)&0x7f|0x80),
			byte((v>>14)&0x7f|0x80),
			byte((v>>21)&0x7f|0x80),
			byte((v>>28)&0x7f|0x80),
			byte((v>>35)&0x7f|0x80),
			byte(v>>42))
	case v < 1<<56:
		b = append(b,
			byte(v&0x7f|0x80),
			byte((v>>7)&0x7f|0x80),
			byte((v>>14)&0x7f|0x80),
			byte((v>>21)&0x7f|0x80),
			byte((v>>28)&0x7f|0x80),
			byte((v>>35)&0x7f|0x80),
			byte((v>>42)&0x7f|0x80),
			byte(v>>49))
	case v < 1<<63:
		b = append(b,
			byte(v&0x7f|0x80),
			byte((v>>7)&0x7f|0x80),
			byte((v>>14)&0x7f|0x80),
			byte((v>>21)&0x7f|0x80),
			byte((v>>28)&0x7f|0x80),
			byte((v>>35)&0x7f|0x80),
			byte((v>>42)&0x7f|0x80),
			byte((v>>49)&0x7f|0x80),
			byte(v>>56))
	default:
		b = append(b,
			byte(v&0x7f|0x80),
			byte((v>>7)&0x7f|0x80),
			byte((v>>14)&0x7f|0x80),
			byte((v>>21)&0x7f|0x80),
			byte((v>>28)&0x7f|0x80),
			byte((v>>35)&0x7f|0x80),
			byte((v>>42)&0x7f|0x80),
			byte((v>>49)&0x7f|0x80),
			byte((v>>56)&0x7f|0x80),
			1)
	}

	_, err = w.Write(b)
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
	return EncodeUvarint(w, uint64(b))
}

func EncodeUint8(w io.Writer, u uint8) (err error) {
	return EncodeUvarint(w, uint64(u))
}

func EncodeUint16(w io.Writer, u uint16) (err error) {
	return EncodeUvarint(w, uint64(u))
}

func EncodeUint32(w io.Writer, u uint32) (err error) {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], u)
	_, err = w.Write(buf[:])
	return
}

func EncodeUint64(w io.Writer, u uint64) (err error) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], u)
	_, err = w.Write(buf[:])
	return
}

func EncodeUvarint(w io.Writer, u uint64) (err error) {
	var buf [10]byte
	n := binary.PutUvarint(buf[:], u)
	_, err = w.Write(buf[0:n])
	return
}

func UvarintSize(u uint64) int {
	var buf [10]byte
	n := binary.PutUvarint(buf[:], u)
	return n
}

//----------------------------------------
// Other

func EncodeBool(w io.Writer, b bool) (err error) {
	if b {
		err = EncodeUint8(w, 1) // same as EncodeUvarint(w, 1).
	} else {
		err = EncodeUint8(w, 0) // same as EncodeUvarint(w, 0).
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
	var s = t.Unix()
	var ns = int32(t.Nanosecond()) // this int64 -> int32 is safe.

	// TODO: We are hand-encoding a struct until MarshalAmino/UnmarshalAmino is supported.

	err = encodeFieldNumberAndTyp3(w, 1, Typ3_8Byte)
	if err != nil {
		return
	}
	err = EncodeInt64(w, s)
	if err != nil {
		return
	}

	err = encodeFieldNumberAndTyp3(w, 2, Typ3_4Byte)
	if err != nil {
		return
	}
	err = EncodeInt32(w, ns)
	if err != nil {
		return
	}

	return
}

func EncodeByteSlice(w io.Writer, bz []byte) (err error) {
	err = EncodeUvarint(w, uint64(len(bz)))
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
