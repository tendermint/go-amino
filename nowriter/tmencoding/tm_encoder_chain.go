package tmencoding

import "bytes"
import "time"

// assert adaptor works at compile time to fulfill TMEncoderBuilderIntr
var _ TMEncoderBuilderIntr = (*TMEncoderChain)(nil)

// wrap a BytesOut encoder with a standard stateful bytes.Buffer
// to provide the TMEncoderBuilderIntr
type TMEncoderChain struct {
	buf  bytes.Buffer
	pure TMEncoderPure
}

func NewTMEncoderChain(pure TMEncoderPure) *TMEncoderChain {
	return &TMEncoderChain{bytes.Buffer{}, pure}
}

func (a *TMEncoderChain) Bytes() []byte {
	return a.buf.Bytes()
}

func (a *TMEncoderChain) EncodeBool(b bool) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.EncodeBool(b))
	return a
}

func (a *TMEncoderChain) EncodeFloat32(f float32) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.EncodeFloat32(f))
	return a
}

func (a *TMEncoderChain) EncodeFloat64(f float64) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.EncodeFloat64(f))
	return a
}

func (a *TMEncoderChain) EncodeInt8(i int8) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.EncodeInt8(i))
	return a
}

func (a *TMEncoderChain) EncodeInt16(i int16) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.EncodeInt16(i))
	return a
}

func (a *TMEncoderChain) EncodeInt32(i int32) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.EncodeInt32(i))
	return a
}

func (a *TMEncoderChain) EncodeInt64(i int64) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.EncodeInt64(i))
	return a
}

func (a *TMEncoderChain) EncodeOctet(b byte) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.EncodeOctet(b))
	return a
}

func (a *TMEncoderChain) EncodeOctets(b []byte) TMEncoderBuilderIntr {
	a.buf.Write(b)
	return a
}

func (a *TMEncoderChain) EncodeTime(t time.Time) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.EncodeTime(t))
	return a
}

func (a *TMEncoderChain) EncodeUint8(i uint8) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.EncodeUint8(i))
	return a
}

func (a *TMEncoderChain) EncodeUint16s(iz []uint16) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.EncodeUint16s(iz))
	return a
}

func (a *TMEncoderChain) EncodeUint32(i uint32) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.EncodeUint32(i))
	return a
}

func (a *TMEncoderChain) EncodeUint64(i uint64) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.EncodeUint64(i))
	return a
}

func (a *TMEncoderChain) EncodeUvarint(i uint) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.EncodeUvarint(i))
	return a
}

func (a *TMEncoderChain) EncodeVarint(i int) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.EncodeVarint(i))
	return a
}
