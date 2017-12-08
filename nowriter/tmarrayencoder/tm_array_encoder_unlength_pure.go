package tmarrayencoder

import (
	"time"
)

type TMArrayEncoderUnlengthPure struct {
}

var _ TMArrayEncoderUnlength = TMArrayEncoderUnlengthPure{}

func (a TMArrayEncoderUnlengthPure) EncodeBoolArray(b []bool) (ary []byte) {
	for _, e := range b {
		ary = append(ary, elementEncoder.EncodeBool(e)...)
	}
	return
}

func (a TMArrayEncoderUnlengthPure) EncodeFloat32Array(f []float32) (ary []byte) {
	for _, e := range f {
		ary = append(ary, elementEncoder.EncodeFloat32(e)...)
	}
	return
}

func (a TMArrayEncoderUnlengthPure) EncodeFloat64Array(f []float64) (ary []byte) {
	for _, e := range f {
		ary = append(ary, elementEncoder.EncodeFloat64(e)...)
	}
	return
}

func (a TMArrayEncoderUnlengthPure) EncodeInt8Array(i []int8) (ary []byte) {
	for _, e := range i {
		ary = append(ary, elementEncoder.EncodeInt8(e)...)
	}
	return
}

func (a TMArrayEncoderUnlengthPure) EncodeInt16Array(i []int16) (ary []byte) {
	for _, e := range i {
		ary = append(ary, elementEncoder.EncodeInt16(e)...)
	}
	return
}

func (a TMArrayEncoderUnlengthPure) EncodeInt32Array(i []int32) (ary []byte) {
	for _, e := range i {
		ary = append(ary, elementEncoder.EncodeInt32(e)...)
	}
	return
}

func (a TMArrayEncoderUnlengthPure) EncodeInt64Array(i []int64) (ary []byte) {
	for _, e := range i {
		ary = append(ary, elementEncoder.EncodeInt64(e)...)
	}
	return
}

func (a TMArrayEncoderUnlengthPure) EncodeOctetArray(b []byte) (ary []byte) {
	for _, e := range b {
		ary = append(ary, elementEncoder.EncodeOctet(e)...)
	}
	return
}

func (a TMArrayEncoderUnlengthPure) EncodeTimeArray(t []time.Time) (ary []byte) {
	for _, e := range t {
		ary = append(ary, elementEncoder.EncodeTime(e)...)
	}
	return
}

func (a TMArrayEncoderUnlengthPure) EncodeUint8Array(i []uint8) (ary []byte) {
	for _, e := range i {
		ary = append(ary, elementEncoder.EncodeUint8(e)...)
	}
	return
}

func (a TMArrayEncoderUnlengthPure) EncodeUint16Array(i []uint16) (ary []byte) {
	for _, e := range i {
		ary = append(ary, elementEncoder.EncodeUint16(e)...)
	}
	return
}

func (a TMArrayEncoderUnlengthPure) EncodeUint32Array(i []uint32) (ary []byte) {
	for _, e := range i {
		ary = append(ary, elementEncoder.EncodeUint32(e)...)
	}
	return
}

func (a TMArrayEncoderUnlengthPure) EncodeUint64Array(i []uint64) (ary []byte) {
	for _, e := range i {
		ary = append(ary, elementEncoder.EncodeUint64(e)...)
	}
	return
}

func (a TMArrayEncoderUnlengthPure) EncodeUvarintArray(i []uint) (ary []byte) {
	for _, e := range i {
		ary = append(ary, elementEncoder.EncodeUvarint(e)...)
	}
	return
}

func (a TMArrayEncoderUnlengthPure) EncodeVarintArray(i []int) (ary []byte) {
	for _, e := range i {
		ary = append(ary, elementEncoder.EncodeVarint(e)...)
	}
	return
}

func (a TMArrayEncoderUnlengthPure) PrefixStatus(TMArrayEncoderUnlength) {
}
