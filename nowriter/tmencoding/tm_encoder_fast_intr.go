package tmencoding

import "io"

type TMEncoderFastIOWriterIntr interface {
	WriteBool(b bool, w io.Writer, n *int, err *error)
	WriteByte(b byte, w io.Writer, n *int, err *error)
	WriteInt8(i int8, w io.Writer, n *int, err *error)
	WriteInt16(i int16, w io.Writer, n *int, err *error)
	WriteInt32(i int32, w io.Writer, n *int, err *error)
	WriteInt64(i int64, w io.Writer, n *int, err *error)
	WriteTo(bz []byte, w io.Writer, n *int, err *error)
	WriteUint8(i uint8, w io.Writer, n *int, err *error)
	WriteUint16(i uint16, w io.Writer, n *int, err *error)
	WriteUint16s(iz []uint16, w io.Writer, n *int, err *error)
	WriteUint32(i uint32, w io.Writer, n *int, err *error)
	WriteUint64(i uint64, w io.Writer, n *int, err *error)
	WriteUvarint(i uint, w io.Writer, n *int, err *error)
	WriteVarint(i int, w io.Writer, n *int, err *error)
}
