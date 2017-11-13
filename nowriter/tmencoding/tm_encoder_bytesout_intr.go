package tmencoding

import "time"

// Simplest pure interface for encoding and generally preferred.
type TMEncoderBytesOutIntr interface {
	WriteBool(b bool) []byte
	WriteFloat32(f float32) []byte
	WriteFloat64(f float64) []byte
	WriteInt8(i int8) []byte
	WriteInt16(i int16) []byte
	WriteInt32(i int32) []byte
	WriteInt64(i int64) []byte
	WriteOctet(b byte) []byte
	WriteOctets(b []byte) []byte
	WriteTime(t time.Time) []byte
	WriteUint8(i uint8) []byte
	WriteUint16(i uint16) []byte
	WriteUint16s(iz []uint16) []byte
	WriteUint32(i uint32) []byte
	WriteUint64(i uint64) []byte
	WriteUvarint(i uint) []byte
	WriteVarint(i int) []byte
}
