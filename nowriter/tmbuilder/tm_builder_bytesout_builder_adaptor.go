package tmbuilder

import "bytes"
import "time"

// assert adaptor works at compile time to fulfill TMBuilder
var _ TMBuilder = (*TMBuilderBytesOutAdaptor)(nil)

// wrap a BytesOut builder with a standard stateful bytes.Buffer
// to provide the TMBuilder
type TMBuilderBytesOutAdaptor struct {
	buf  bytes.Buffer
	pure TMBuilderBytesOut
}

func NewTMBuilderBytesOutAdaptor(pure TMBuilderBytesOut) *TMBuilderBytesOutAdaptor {
	return &TMBuilderBytesOutAdaptor{bytes.Buffer{}, pure}
}

func (a *TMBuilderBytesOutAdaptor) Bytes() []byte {
	return a.buf.Bytes()
}

func (a *TMBuilderBytesOutAdaptor) WriteBool(b bool) TMBuilder {
	a.buf.Write(a.pure.WriteBool(b))
	return a
}

func (a *TMBuilderBytesOutAdaptor) WriteFloat32(f float32) TMBuilder {
	a.buf.Write(a.pure.WriteFloat32(f))
	return a
}

func (a *TMBuilderBytesOutAdaptor) WriteFloat64(f float64) TMBuilder {
	a.buf.Write(a.pure.WriteFloat64(f))
	return a
}

func (a *TMBuilderBytesOutAdaptor) WriteInt8(i int8) TMBuilder {
	a.buf.Write(a.pure.WriteInt8(i))
	return a
}

func (a *TMBuilderBytesOutAdaptor) WriteInt16(i int16) TMBuilder {
	a.buf.Write(a.pure.WriteInt16(i))
	return a
}

func (a *TMBuilderBytesOutAdaptor) WriteInt32(i int32) TMBuilder {
	a.buf.Write(a.pure.WriteInt32(i))
	return a
}

func (a *TMBuilderBytesOutAdaptor) WriteInt64(i int64) TMBuilder {
	a.buf.Write(a.pure.WriteInt64(i))
	return a
}

func (a *TMBuilderBytesOutAdaptor) WriteOctet(b byte) TMBuilder {
	a.buf.Write(a.pure.WriteOctet(b))
	return a
}

func (a *TMBuilderBytesOutAdaptor) WriteOctets(b []byte) TMBuilder {
	a.buf.Write(b)
	return a
}

func (a *TMBuilderBytesOutAdaptor) WriteTime(t time.Time) TMBuilder {
	a.buf.Write(a.pure.WriteTime(t))
	return a
}

func (a *TMBuilderBytesOutAdaptor) WriteUint8(i uint8) TMBuilder {
	a.buf.Write(a.pure.WriteUint8(i))
	return a
}

func (a *TMBuilderBytesOutAdaptor) WriteUint16s(iz []uint16) TMBuilder {
	a.buf.Write(a.pure.WriteUint16s(iz))
	return a
}

func (a *TMBuilderBytesOutAdaptor) WriteUint32(i uint32) TMBuilder {
	a.buf.Write(a.pure.WriteUint32(i))
	return a
}

func (a *TMBuilderBytesOutAdaptor) WriteUint64(i uint64) TMBuilder {
	a.buf.Write(a.pure.WriteUint64(i))
	return a
}

func (a *TMBuilderBytesOutAdaptor) WriteUvarint(i uint) TMBuilder {
	a.buf.Write(a.pure.WriteUvarint(i))
	return a
}

func (a *TMBuilderBytesOutAdaptor) WriteVarint(i int) TMBuilder {
	a.buf.Write(a.pure.WriteVarint(i))
	return a
}
