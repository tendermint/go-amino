package tmencoding

import (
	"bufio"
	"bytes"
	"encoding/binary"
	cmn "github.com/tendermint/tmlibs/common"
	"math"
	"time"
)

type TMEncoderPure struct {
}

var _ TMEncoderBytesOutIntr = TMEncoderPure{}
var legacy TMEncoderLegacy

func (e TMEncoderPure) WriteBool(b bool) []byte {
	var bb byte
	if b {
		bb = 0x01
	} else {
		bb = 0x00
	}
	return []byte{bb}
}

func (e TMEncoderPure) WriteFloat32(f float32) []byte {
	return e.WriteUint32(math.Float32bits(f))
}

func (e TMEncoderPure) WriteFloat64(f float64) []byte {
	return e.WriteUint64(math.Float64bits(f))
}

func (e TMEncoderPure) WriteInt8(i int8) []byte {
	return e.WriteOctet(byte(i))
}

func (e TMEncoderPure) WriteInt16(i int16) []byte {
	var buf [2]byte
	binary.BigEndian.PutUint16(buf[:], uint16(i))
	return buf[:]
}

func (e TMEncoderPure) WriteInt32(i int32) []byte {
	var buf [2]byte
	binary.BigEndian.PutUint16(buf[:], uint16(i))
	return buf[:]
}

func (e TMEncoderPure) WriteInt64(i int64) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(i))
	return buf[:]
}

func (e TMEncoderPure) WriteOctet(b byte) []byte {
	return []byte{b}
}

// for orthogonality only
func (e TMEncoderPure) WriteOctets(b []byte) []byte {
	arr := make([]byte, len(b))
	copy(arr, b)
	return arr
}

func (e TMEncoderPure) WriteTime(t time.Time) []byte {
	nanosecs := t.UnixNano()
	millisecs := nanosecs / 1000000
	if nanosecs < 0 {
		cmn.PanicSanity("can't encode times below 1970")
	}
	return e.WriteInt64(millisecs * 1000000)
}

func (e TMEncoderPure) WriteUint8(i uint8) []byte {
	return e.WriteOctet(byte(i))
}

func (e TMEncoderPure) WriteUint16(i uint16) []byte {
	var buf [2]byte
	binary.BigEndian.PutUint16(buf[:], uint16(i))
	return buf[:]
}

func (e TMEncoderPure) WriteUint16s(iz []uint16) []byte {
	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	var inst_n int
	n := &inst_n
	var inst_err error
	err := &inst_err

	legacy.WriteUint32(uint32(len(iz)), w, n, err)
	for _, i := range iz {
		legacy.WriteUint16(i, w, n, err)
		if *err != nil {
			return nil
		}
	}

	return b.Bytes()
}

func (e TMEncoderPure) WriteUint32(i uint32) []byte {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], uint32(i))
	return buf[:]
}

func (e TMEncoderPure) WriteUint64(i uint64) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(i))
	return buf[:]
}

func (e TMEncoderPure) WriteUvarint(i uint) []byte {
	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	var inst_n int
	n := &inst_n
	var inst_err error
	err := &inst_err

	var size = uvarintSize(uint64(i))
	legacy.WriteUint8(uint8(size), w, n, err)
	if size > 0 {
		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], uint64(i))
		legacy.WriteTo(buf[(8-size):], w, n, err)
	}

	return b.Bytes()
}

func (e TMEncoderPure) WriteVarint(i int) []byte {
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
		legacy.WriteUint8(uint8(size+0xF0), w, n, err)
	} else {
		legacy.WriteUint8(uint8(size), w, n, err)
	}
	if size > 0 {
		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], uint64(i))
		legacy.WriteTo(buf[(8-size):], w, n, err)
	}

	return b.Bytes()
}
