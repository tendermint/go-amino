package tmbuilder

import "io"
import "time"

// Legacy interface following old code as closely as possible.
// Changed `EncodeBytes` to `EncodeOctets` to solve `go vet` issue.
// The explicit declaration of this interface provides a verbal
// discussion tool as well as a code marking work and it allows
// us to migrate away from using the global namespace for the Encode*
// methods in the old 'wire.go' as we refactor.  This class may
// disappear once the refactoring is complete.
type TMBuilderFastIOEncoder interface {
	EncodeBool(b bool, w io.Writer, n *int, err *error)
	EncodeFloat32(f float32, w io.Writer, n *int, err *error)
	EncodeFloat64(f float64, w io.Writer, n *int, err *error)
	EncodeInt8(i int8, w io.Writer, n *int, err *error)
	EncodeInt16(i int16, w io.Writer, n *int, err *error)
	EncodeInt32(i int32, w io.Writer, n *int, err *error)
	EncodeInt64(i int64, w io.Writer, n *int, err *error)
	EncodeOctet(b byte, w io.Writer, n *int, err *error)
	EncodeTime(t time.Time, w io.Writer, n *int, err *error)
	EncodeTo(bz []byte, w io.Writer, n *int, err *error)
	EncodeUint8(i uint8, w io.Writer, n *int, err *error)
	EncodeUint16(i uint16, w io.Writer, n *int, err *error)
	EncodeUint16s(iz []uint16, w io.Writer, n *int, err *error)
	EncodeUint32(i uint32, w io.Writer, n *int, err *error)
	EncodeUint64(i uint64, w io.Writer, n *int, err *error)
	EncodeUvarint(i uint, w io.Writer, n *int, err *error)
	EncodeVarint(i int, w io.Writer, n *int, err *error)
}
