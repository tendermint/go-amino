package tmarrayencoder

import (
	"time"
)

type TMArrayEncoderLengthPure struct {
}

var _ TMArrayEncoder = TMArrayEncoderLengthPure{}
var simpleAE TMArrayEncoder = TMArrayEncoderUnlengthPure{}

// encode length prefix
func encodeLP(i int) []byte {
	return elementEncoder.EncodeUvarint(uint(i))
}

func (a TMArrayEncoderLengthPure) EncodeBoolArray(b []bool) []byte {
	return append(encodeLP(len(b)), simpleAE.EncodeBoolArray(b)...)
}

func (a TMArrayEncoderLengthPure) EncodeFloat32Array(f []float32) []byte {
	return append(encodeLP(len(f)), simpleAE.EncodeFloat32Array(f)...)
}

func (a TMArrayEncoderLengthPure) EncodeFloat64Array(f []float64) []byte {
	return append(encodeLP(len(f)), simpleAE.EncodeFloat64Array(f)...)
}

func (a TMArrayEncoderLengthPure) EncodeInt8Array(i []int8) []byte {
	return append(encodeLP(len(i)), simpleAE.EncodeInt8Array(i)...)
}

func (a TMArrayEncoderLengthPure) EncodeInt16Array(i []int16) []byte {
	return append(encodeLP(len(i)), simpleAE.EncodeInt16Array(i)...)
}

func (a TMArrayEncoderLengthPure) EncodeInt32Array(i []int32) []byte {
	return append(encodeLP(len(i)), simpleAE.EncodeInt32Array(i)...)
}

func (a TMArrayEncoderLengthPure) EncodeInt64Array(i []int64) []byte {
	return append(encodeLP(len(i)), simpleAE.EncodeInt64Array(i)...)
}

func (a TMArrayEncoderLengthPure) EncodeOctetArray(b []byte) []byte {
	return append(encodeLP(len(b)), simpleAE.EncodeOctetArray(b)...)
}

func (a TMArrayEncoderLengthPure) EncodeTimeArray(t []time.Time) []byte {
	return append(encodeLP(len(t)), simpleAE.EncodeTimeArray(t)...)
}

func (a TMArrayEncoderLengthPure) EncodeUint8Array(i []uint8) []byte {
	return append(encodeLP(len(i)), simpleAE.EncodeUint8Array(i)...)
}

func (a TMArrayEncoderLengthPure) EncodeUint16Array(i []uint16) []byte {
	return append(encodeLP(len(i)), simpleAE.EncodeUint16Array(i)...)
}

func (a TMArrayEncoderLengthPure) EncodeUint32Array(i []uint32) []byte {
	return append(encodeLP(len(i)), simpleAE.EncodeUint32Array(i)...)
}

func (a TMArrayEncoderLengthPure) EncodeUint64Array(i []uint64) []byte {
	return append(encodeLP(len(i)), simpleAE.EncodeUint64Array(i)...)
}

func (a TMArrayEncoderLengthPure) EncodeUvarintArray(i []uint) []byte {
	return append(encodeLP(len(i)), simpleAE.EncodeUvarintArray(i)...)
}

func (a TMArrayEncoderLengthPure) EncodeVarintArray(i []int) []byte {
	return append(encodeLP(len(i)), simpleAE.EncodeVarintArray(i)...)
}

func (a TMArrayEncoderLengthPure) PrefixStatus(TMArrayEncoderLength) {
}
