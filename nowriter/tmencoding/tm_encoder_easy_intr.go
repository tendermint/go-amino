package tmencoding

import "time"

type TMEncoderEasyIntr interface {
	WriteBool(b bool)
	WriteFloat32(f float32)
	WriteFloat64(f float64)
	WriteInt8(i int8)
	WriteInt16(i int16)
	WriteInt32(i int32)
	WriteInt64(i int64)
	WriteOctet(b byte)
	WriteOctets(b []byte)
	WriteTime(t time.Time)
	WriteUint8(i uint8)
	WriteUint16s(iz []uint16)
	WriteUint32(i uint32)
	WriteUint64(i uint64)
	WriteUvarint(i uint)
	WriteVarint(i int)
}
