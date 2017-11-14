package tmbuilder

import "io"
import "encoding/binary"
import "math"
import "time"
import cmn "github.com/tendermint/tmlibs/common"

// Implementation of the legacy (`TMBuilderFastIOEncoder`) interface
type TMBuilderLegacy struct {
}

var Legacy *TMBuilderLegacy = &TMBuilderLegacy{}           // convenience
var _ TMBuilderFastIOEncoder = (*TMBuilderLegacy)(nil) // complete

// Does not use builder pattern to encourage migration away from this struct
func (e *TMBuilderLegacy) EncodeBool(b bool, w io.Writer, n *int, err *error) {
	var bb byte
	if b {
		bb = 0x01
	} else {
		bb = 0x00
	}
	e.EncodeTo([]byte{bb}, w, n, err)
}

func (e *TMBuilderLegacy) EncodeFloat32(f float32, w io.Writer, n *int, err *error) {
	e.EncodeUint32(math.Float32bits(f), w, n, err)
}

func (e *TMBuilderLegacy) EncodeFloat64(f float64, w io.Writer, n *int, err *error) {
	e.EncodeUint64(math.Float64bits(f), w, n, err)
}

func (e *TMBuilderLegacy) EncodeInt8(i int8, w io.Writer, n *int, err *error) {
	e.EncodeOctet(byte(i), w, n, err)
}

func (e *TMBuilderLegacy) EncodeInt16(i int16, w io.Writer, n *int, err *error) {
	var buf [2]byte
	binary.BigEndian.PutUint16(buf[:], uint16(i))
	*n += 2
	e.EncodeTo(buf[:], w, n, err)
}

func (e *TMBuilderLegacy) EncodeInt32(i int32, w io.Writer, n *int, err *error) {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], uint32(i))
	*n += 4
	e.EncodeTo(buf[:], w, n, err)
}
func (e *TMBuilderLegacy) EncodeInt64(i int64, w io.Writer, n *int, err *error) {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(i))
	*n += 8
	e.EncodeTo(buf[:], w, n, err)
}

func (e *TMBuilderLegacy) EncodeOctet(b byte, w io.Writer, n *int, err *error) {
	e.EncodeTo([]byte{b}, w, n, err)
}

func (e *TMBuilderLegacy) EncodeTime(t time.Time, w io.Writer, n *int, err *error) {
	nanosecs := t.UnixNano()
	millisecs := nanosecs / 1000000
	if nanosecs < 0 {
		cmn.PanicSanity("can't encode times below 1970")
	} else {
		e.EncodeInt64(millisecs*1000000, w, n, err)
	}
}

func (e *TMBuilderLegacy) EncodeUint8(i uint8, w io.Writer, n *int, err *error) {
	e.EncodeOctet(byte(i), w, n, err)
}

func (e *TMBuilderLegacy) EncodeUint16(i uint16, w io.Writer, n *int, err *error) {
	var buf [2]byte
	binary.BigEndian.PutUint16(buf[:], uint16(i))
	*n += 2
	e.EncodeTo(buf[:], w, n, err)
}

func (e *TMBuilderLegacy) EncodeUint16s(iz []uint16, w io.Writer, n *int, err *error) {
	e.EncodeUint32(uint32(len(iz)), w, n, err)
	for _, i := range iz {
		e.EncodeUint16(i, w, n, err)
		if *err != nil {
			return
		}
	}
}
func (e *TMBuilderLegacy) EncodeUint32(i uint32, w io.Writer, n *int, err *error) {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], uint32(i))
	*n += 4
	e.EncodeTo(buf[:], w, n, err)
}

func (e *TMBuilderLegacy) EncodeUint64(i uint64, w io.Writer, n *int, err *error) {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(i))
	*n += 8
	e.EncodeTo(buf[:], w, n, err)
}

func (e *TMBuilderLegacy) EncodeUvarint(i uint, w io.Writer, n *int, err *error) {
	var size = uvarintSize(uint64(i))
	e.EncodeUint8(uint8(size), w, n, err)
	if size > 0 {
		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], uint64(i))
		e.EncodeTo(buf[(8-size):], w, n, err)
	}
}

func (e *TMBuilderLegacy) EncodeVarint(i int, w io.Writer, n *int, err *error) {
	var negate = false
	if i < 0 {
		negate = true
		i = -i
	}
	var size = uvarintSize(uint64(i))
	if negate {
		// e.g. 0xF1 for a single negative byte
		e.EncodeUint8(uint8(size+0xF0), w, n, err)
	} else {
		e.EncodeUint8(uint8(size), w, n, err)
	}
	if size > 0 {
		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], uint64(i))
		e.EncodeTo(buf[(8-size):], w, n, err)
	}
}

// Encode all of bz to w
// Increment n and set err accordingly.
func (e *TMBuilderLegacy) EncodeTo(bz []byte, w io.Writer, n *int, err *error) {
	if *err != nil {
		return
	}
	n_, err_ := w.Write(bz)
	*n += n_
	*err = err_
}

func uvarintSize(i uint64) int {
	if i == 0 {
		return 0
	}
	if i < 1<<8 {
		return 1
	}
	if i < 1<<16 {
		return 2
	}
	if i < 1<<24 {
		return 3
	}
	if i < 1<<32 {
		return 4
	}
	if i < 1<<40 {
		return 5
	}
	if i < 1<<48 {
		return 6
	}
	if i < 1<<56 {
		return 7
	}
	return 8
}
