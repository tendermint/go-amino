package tmbuilder

import (
	"bufio"
	"bytes"
	"encoding/binary"
	cmn "github.com/tendermint/tmlibs/common"
	"math"
	"time"
)

// Implementation of the TMBuilderBytesOut interface.
type TMBuilderPure struct {
}

var _ TMBuilderBytesOut = TMBuilderPure{}
var legacy TMBuilderLegacy

func (e TMBuilderPure) EncodeBool(b bool) []byte {
	var bb byte
	if b {
		bb = 0x01
	} else {
		bb = 0x00
	}
	return []byte{bb}
}

func (e TMBuilderPure) EncodeFloat32(f float32) []byte {
	return e.EncodeUint32(math.Float32bits(f))
}

func (e TMBuilderPure) EncodeFloat64(f float64) []byte {
	return e.EncodeUint64(math.Float64bits(f))
}

func (e TMBuilderPure) EncodeInt8(i int8) []byte {
	return e.EncodeOctet(byte(i))
}

func (e TMBuilderPure) EncodeInt16(i int16) []byte {
	var buf [2]byte
	binary.BigEndian.PutUint16(buf[:], uint16(i))
	return buf[:]
}

func (e TMBuilderPure) EncodeInt32(i int32) []byte {
	var buf [2]byte
	binary.BigEndian.PutUint16(buf[:], uint16(i))
	return buf[:]
}

func (e TMBuilderPure) EncodeInt64(i int64) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(i))
	return buf[:]
}

func (e TMBuilderPure) EncodeOctet(b byte) []byte {
	return []byte{b}
}

// for orthogonality only
func (e TMBuilderPure) EncodeOctets(b []byte) []byte {
	arr := make([]byte, len(b))
	copy(arr, b)
	return arr
}

func (e TMBuilderPure) EncodeTime(t time.Time) []byte {
	nanosecs := t.UnixNano()
	millisecs := nanosecs / 1000000
	if nanosecs < 0 {
		cmn.PanicSanity("can't encode times below 1970")
	}
	return e.EncodeInt64(millisecs * 1000000)
}

func (e TMBuilderPure) EncodeUint8(i uint8) []byte {
	return e.EncodeOctet(byte(i))
}

func (e TMBuilderPure) EncodeUint16(i uint16) []byte {
	var buf [2]byte
	binary.BigEndian.PutUint16(buf[:], uint16(i))
	return buf[:]
}

func (e TMBuilderPure) EncodeUint16s(iz []uint16) []byte {
	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	var inst_n int
	n := &inst_n
	var inst_err error
	err := &inst_err

	legacy.EncodeUint32(uint32(len(iz)), w, n, err)
	for _, i := range iz {
		legacy.EncodeUint16(i, w, n, err)
		if *err != nil {
			return nil
		}
	}

	return b.Bytes()
}

func (e TMBuilderPure) EncodeUint32(i uint32) []byte {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], uint32(i))
	return buf[:]
}

func (e TMBuilderPure) EncodeUint64(i uint64) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(i))
	return buf[:]
}

func (e TMBuilderPure) EncodeUvarint(i uint) []byte {
	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	var inst_n int
	n := &inst_n
	var inst_err error
	err := &inst_err

	var size = uvarintSize(uint64(i))
	legacy.EncodeUint8(uint8(size), w, n, err)
	if size > 0 {
		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], uint64(i))
		legacy.EncodeTo(buf[(8-size):], w, n, err)
	}

	return b.Bytes()
}

func (e TMBuilderPure) EncodeVarint(i int) []byte {
	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	var inst_n int
	n := &inst_n
	var inst_err error
	err := &inst_err

	var negate = false
	if i < 0 {
		negate = true
		i = -i
	}
	var size = uvarintSize(uint64(i))
	if negate {
		// e.g. 0xF1 for a single negative byte
		legacy.EncodeUint8(uint8(size+0xF0), w, n, err)
	} else {
		legacy.EncodeUint8(uint8(size), w, n, err)
	}
	if size > 0 {
		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], uint64(i))
		legacy.EncodeTo(buf[(8-size):], w, n, err)
	}

	return b.Bytes()
}
