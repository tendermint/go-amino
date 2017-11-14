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

func (a *TMBuilderBytesOutAdaptor) EncodeBool(b bool) TMBuilder {
	a.buf.Write(a.pure.EncodeBool(b))
	return a
}

func (a *TMBuilderBytesOutAdaptor) EncodeFloat32(f float32) TMBuilder {
	a.buf.Write(a.pure.EncodeFloat32(f))
	return a
}

func (a *TMBuilderBytesOutAdaptor) EncodeFloat64(f float64) TMBuilder {
	a.buf.Write(a.pure.EncodeFloat64(f))
	return a
}

func (a *TMBuilderBytesOutAdaptor) EncodeInt8(i int8) TMBuilder {
	a.buf.Write(a.pure.EncodeInt8(i))
	return a
}

func (a *TMBuilderBytesOutAdaptor) EncodeInt16(i int16) TMBuilder {
	a.buf.Write(a.pure.EncodeInt16(i))
	return a
}

func (a *TMBuilderBytesOutAdaptor) EncodeInt32(i int32) TMBuilder {
	a.buf.Write(a.pure.EncodeInt32(i))
	return a
}

func (a *TMBuilderBytesOutAdaptor) EncodeInt64(i int64) TMBuilder {
	a.buf.Write(a.pure.EncodeInt64(i))
	return a
}

func (a *TMBuilderBytesOutAdaptor) EncodeOctet(b byte) TMBuilder {
	a.buf.Write(a.pure.EncodeOctet(b))
	return a
}

func (a *TMBuilderBytesOutAdaptor) EncodeOctets(b []byte) TMBuilder {
	a.buf.Write(b)
	return a
}

func (a *TMBuilderBytesOutAdaptor) EncodeTime(t time.Time) TMBuilder {
	a.buf.Write(a.pure.EncodeTime(t))
	return a
}

func (a *TMBuilderBytesOutAdaptor) EncodeUint8(i uint8) TMBuilder {
	a.buf.Write(a.pure.EncodeUint8(i))
	return a
}

func (a *TMBuilderBytesOutAdaptor) EncodeUint16s(iz []uint16) TMBuilder {
	a.buf.Write(a.pure.EncodeUint16s(iz))
	return a
}

func (a *TMBuilderBytesOutAdaptor) EncodeUint32(i uint32) TMBuilder {
	a.buf.Write(a.pure.EncodeUint32(i))
	return a
}

func (a *TMBuilderBytesOutAdaptor) EncodeUint64(i uint64) TMBuilder {
	a.buf.Write(a.pure.EncodeUint64(i))
	return a
}

func (a *TMBuilderBytesOutAdaptor) EncodeUvarint(i uint) TMBuilder {
	a.buf.Write(a.pure.EncodeUvarint(i))
	return a
}

func (a *TMBuilderBytesOutAdaptor) EncodeVarint(i int) TMBuilder {
	a.buf.Write(a.pure.EncodeVarint(i))
	return a
}
