package tmencoding

import "bytes"
import "time"

// assert adaptor works at compile time to fulfill TMEncoderBuilderIntr
var _ TMEncoderBuilderIntr = (*TMEncoderBytesOutBuilderAdapter)(nil)

// wrap a BytesOut encoder with a standard stateful bytes.Buffer
// to provide the TMEncoderBuilderIntr
type TMEncoderBytesOutBuilderAdapter struct {
	buf  bytes.Buffer
	pure TMEncoderBytesOutIntr
}

func NewTMEncoderBytesOutBuilderAdapter(pure TMEncoderBytesOutIntr) *TMEncoderBytesOutBuilderAdapter {
	return &TMEncoderBytesOutBuilderAdapter{bytes.Buffer{}, pure}
}

func (a *TMEncoderBytesOutBuilderAdapter) Bytes() []byte {
	return a.buf.Bytes()
}

func (a *TMEncoderBytesOutBuilderAdapter) WriteBool(b bool) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.WriteBool(b))
	return a
}

func (a *TMEncoderBytesOutBuilderAdapter) WriteFloat32(f float32) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.WriteFloat32(f))
	return a
}

func (a *TMEncoderBytesOutBuilderAdapter) WriteFloat64(f float64) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.WriteFloat64(f))
	return a
}

func (a *TMEncoderBytesOutBuilderAdapter) WriteInt8(i int8) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.WriteInt8(i))
	return a
}

func (a *TMEncoderBytesOutBuilderAdapter) WriteInt16(i int16) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.WriteInt16(i))
	return a
}

func (a *TMEncoderBytesOutBuilderAdapter) WriteInt32(i int32) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.WriteInt32(i))
	return a
}

func (a *TMEncoderBytesOutBuilderAdapter) WriteInt64(i int64) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.WriteInt64(i))
	return a
}

func (a *TMEncoderBytesOutBuilderAdapter) WriteOctet(b byte) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.WriteOctet(b))
	return a
}

func (a *TMEncoderBytesOutBuilderAdapter) WriteOctets(b []byte) TMEncoderBuilderIntr {
	a.buf.Write(b)
	return a
}

func (a *TMEncoderBytesOutBuilderAdapter) WriteTime(t time.Time) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.WriteTime(t))
	return a
}

func (a *TMEncoderBytesOutBuilderAdapter) WriteUint8(i uint8) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.WriteUint8(i))
	return a
}

func (a *TMEncoderBytesOutBuilderAdapter) WriteUint16s(iz []uint16) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.WriteUint16s(iz))
	return a
}

func (a *TMEncoderBytesOutBuilderAdapter) WriteUint32(i uint32) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.WriteUint32(i))
	return a
}

func (a *TMEncoderBytesOutBuilderAdapter) WriteUint64(i uint64) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.WriteUint64(i))
	return a
}

func (a *TMEncoderBytesOutBuilderAdapter) WriteUvarint(i uint) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.WriteUvarint(i))
	return a
}

func (a *TMEncoderBytesOutBuilderAdapter) WriteVarint(i int) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.WriteVarint(i))
	return a
}
