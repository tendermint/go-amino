package tmbuilder

import "time"

// Allow chaining builder pattern syntactic sugar
// teb.WriteInt8(42).WriteInt32(my_integer).Bytes()
type TMBuilder interface {
	Bytes() []byte
	WriteBool(b bool) TMBuilder
	WriteFloat32(f float32) TMBuilder
	WriteFloat64(f float64) TMBuilder
	WriteInt8(i int8) TMBuilder
	WriteInt16(i int16) TMBuilder
	WriteInt32(i int32) TMBuilder
	WriteInt64(i int64) TMBuilder
	WriteOctet(b byte) TMBuilder
	WriteOctets(b []byte) TMBuilder
	WriteTime(t time.Time) TMBuilder
	WriteUint8(i uint8) TMBuilder
	WriteUint16s(iz []uint16) TMBuilder
	WriteUint32(i uint32) TMBuilder
	WriteUint64(i uint64) TMBuilder
	WriteUvarint(i uint) TMBuilder
	WriteVarint(i int) TMBuilder
}

// Ensure chaining syntax functions properly with compile-time assertion.
func sugar_assertion_do_not_call(t TMBuilder) int {
	var confirm_syntax []byte
	confirm_syntax = t.WriteBool(false).WriteUint64(17).Bytes()
	return len(confirm_syntax)
}
