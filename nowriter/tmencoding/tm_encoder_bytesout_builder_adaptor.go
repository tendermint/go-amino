package tmencoding

import "bytes"

type TMEncoderBytesOutBuilderAdaptor struct {
	TMEncoderBuilderIntr
	buf  bytes.Buffer
	pure TMEncoderBytesOutIntr
}

func (a *TMEncoderBytesOutBuilderAdaptor) Bytes() []byte {
	return a.buf.Bytes()
}

func (a *TMEncoderBytesOutBuilderAdaptor) WriteBool(b bool) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.WriteBool(b))
	return a
}

func (a *TMEncoderBytesOutBuilderAdaptor) WriteFloat32(f float32) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.WriteFloat32(f))
	return a
}

func (a *TMEncoderBytesOutBuilderAdaptor) WriteFloat64(f float64) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.WriteFloat64(f))
	return a
}

func (a *TMEncoderBytesOutBuilderAdaptor) WriteInt8(i int8) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.WriteInt8(i))
	return a
}

func (a *TMEncoderBytesOutBuilderAdaptor) WriteInt16(i int16) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.WriteInt16(i))
	return a
}

func (a *TMEncoderBytesOutBuilderAdaptor) WriteInt32(i int32) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.WriteInt32(i))
	return a
}

func (a *TMEncoderBytesOutBuilderAdaptor) WriteInt64(i int64) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.WriteInt64(i))
	return a
}

func (a *TMEncoderBytesOutBuilderAdaptor) WriteOctet(b byte) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.WriteOctet(b))
	return a
}

func (a *TMEncoderBytesOutBuilderAdaptor) WriteOctets(b []byte) TMEncoderBuilderIntr {
	a.buf.Write(b)
	return a
}

func (a *TMEncoderBytesOutBuilderAdaptor) WriteUint8(i uint8) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.WriteUint8(i))
	return a
}

func (a *TMEncoderBytesOutBuilderAdaptor) WriteUint16s(iz []uint16) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.WriteUint16s(iz))
	return a
}

func (a *TMEncoderBytesOutBuilderAdaptor) WriteUint32(i uint32) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.WriteUint32(i))
	return a
}

func (a *TMEncoderBytesOutBuilderAdaptor) WriteUint64(i uint64) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.WriteUint64(i))
	return a
}

func (a *TMEncoderBytesOutBuilderAdaptor) WriteUvarint(i uint) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.WriteUvarint(i))
	return a
}

func (a *TMEncoderBytesOutBuilderAdaptor) WriteVarint(i int) TMEncoderBuilderIntr {
	a.buf.Write(a.pure.WriteVarint(i))
	return a
}
