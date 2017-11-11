package tmencoding

type TMEncoderBytesOutIntr interface {
	WriteBool(b bool) []byte
	WriteByte(b byte) []byte
	WriteInt8(i int8) []byte
	WriteInt16(i int16) []byte
	WriteInt32(i int32) []byte
	WriteInt64(i int64) []byte
	WriteUint8(i uint8) []byte
	WriteUint16(i uint16) []byte
	WriteUint16s(iz []uint16) []byte
	WriteUint32(i uint32) []byte
	WriteUint64(i uint64) []byte
	WriteUvarint(i uint) []byte
	WriteVarint(i int) []byte
}
