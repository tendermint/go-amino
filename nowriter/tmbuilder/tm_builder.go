package tmbuilder

import "time"

// Allow chaining builder pattern syntactic sugar
// teb.EncodeInt8(42).EncodeInt32(my_integer).Bytes()
type TMBuilder interface {
	Bytes() []byte
	EncodeBool(b bool) TMBuilder
	EncodeFloat32(f float32) TMBuilder
	EncodeFloat64(f float64) TMBuilder
	EncodeInt8(i int8) TMBuilder
	EncodeInt16(i int16) TMBuilder
	EncodeInt32(i int32) TMBuilder
	EncodeInt64(i int64) TMBuilder
	EncodeOctet(b byte) TMBuilder
	EncodeOctets(b []byte) TMBuilder
	EncodeTime(t time.Time) TMBuilder
	EncodeUint8(i uint8) TMBuilder
	EncodeUint16s(iz []uint16) TMBuilder
	EncodeUint32(i uint32) TMBuilder
	EncodeUint64(i uint64) TMBuilder
	EncodeUvarint(i uint) TMBuilder
	EncodeVarint(i int) TMBuilder
}

// Ensure chaining syntax functions properly with compile-time assertion.
func sugar_assertion_do_not_call(t TMBuilder) int {
	var confirm_syntax []byte
	confirm_syntax = t.EncodeBool(false).EncodeUint64(17).Bytes()
	return len(confirm_syntax)
}
