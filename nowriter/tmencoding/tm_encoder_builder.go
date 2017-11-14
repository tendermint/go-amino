package tmencoding

import "time"

// Allow chaining builder pattern syntactic sugar variation of Facade
// teb.EncodeInt8(42).EncodeInt32(my_integer).Bytes()
type TMEncoderBuilderIntr interface {
	Bytes() []byte
	EncodeBool(b bool) TMEncoderBuilderIntr
	EncodeFloat32(f float32) TMEncoderBuilderIntr
	EncodeFloat64(f float64) TMEncoderBuilderIntr
	EncodeInt8(i int8) TMEncoderBuilderIntr
	EncodeInt16(i int16) TMEncoderBuilderIntr
	EncodeInt32(i int32) TMEncoderBuilderIntr
	EncodeInt64(i int64) TMEncoderBuilderIntr
	EncodeOctet(b byte) TMEncoderBuilderIntr
	EncodeOctets(b []byte) TMEncoderBuilderIntr
	EncodeTime(t time.Time) TMEncoderBuilderIntr
	EncodeUint8(i uint8) TMEncoderBuilderIntr
	EncodeUint16s(iz []uint16) TMEncoderBuilderIntr
	EncodeUint32(i uint32) TMEncoderBuilderIntr
	EncodeUint64(i uint64) TMEncoderBuilderIntr
	EncodeUvarint(i uint) TMEncoderBuilderIntr
	EncodeVarint(i int) TMEncoderBuilderIntr
}

// Ensure chaining syntax functions properly with compile-time assertion.
func sugar_assertion_do_not_call(t TMEncoderBuilderIntr) int {
	var confirm_syntax []byte
	confirm_syntax = t.EncodeBool(false).EncodeUint64(17).Bytes()
	return len(confirm_syntax)
}
