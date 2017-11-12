package tmencoding

// Allow chaining builder pattern syntactic sugar variation of Facade
// teb.WriteInt8(42).WriteInt32(my_integer).Bytes()
type TMEncoderBuilderIntr interface {
	Bytes() []byte
	WriteBool(b bool) TMEncoderBuilderIntr
	WriteFloat32(f float32) TMEncoderBuilderIntr
	WriteFloat64(f float64) TMEncoderBuilderIntr
	WriteInt8(i int8) TMEncoderBuilderIntr
	WriteInt16(i int16) TMEncoderBuilderIntr
	WriteInt32(i int32) TMEncoderBuilderIntr
	WriteInt64(i int64) TMEncoderBuilderIntr
	WriteOctet(b byte) TMEncoderBuilderIntr
	WriteOctets(b []byte) TMEncoderBuilderIntr
	WriteUint8(i uint8) TMEncoderBuilderIntr
	WriteUint16s(iz []uint16) TMEncoderBuilderIntr
	WriteUint32(i uint32) TMEncoderBuilderIntr
	WriteUint64(i uint64) TMEncoderBuilderIntr
	WriteUvarint(i uint) TMEncoderBuilderIntr
	WriteVarint(i int) TMEncoderBuilderIntr
}

// Ensure chaining syntax functions properly with compile-time assertion.
func sugar_assertion_do_not_call(t TMEncoderBuilderIntr) int {
	var confirm_syntax []byte
	confirm_syntax = t.WriteBool(false).WriteUint64(17).Bytes()
	return len(confirm_syntax)
}
