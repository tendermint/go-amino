package tmarrayencoder

import (
	"github.com/tendermint/go-wire/nowriter/tmencoding"
	"time"
)

type TMArrayEncoder interface {
	EncodeBoolArray(b []bool) []byte
	EncodeFloat32Array(f []float32) []byte
	EncodeFloat64Array(f []float64) []byte
	EncodeInt8Array(i []int8) []byte
	EncodeInt16Array(i []int16) []byte
	EncodeInt32Array(i []int32) []byte
	EncodeInt64Array(i []int64) []byte
	EncodeOctetArray(b []byte) []byte
	EncodeTimeArray(t []time.Time) []byte
	EncodeUint8Array(i []uint8) []byte
	EncodeUint16Array(i []uint16) []byte
	EncodeUint32Array(i []uint32) []byte
	EncodeUint64Array(i []uint64) []byte
	EncodeUvarintArray(i []uint) []byte
	EncodeVarintArray(i []int) []byte
}

var elementEncoder tmencoding.TMEncoder = tmencoding.TMEncoderPure{}
