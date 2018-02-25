package wire

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"reflect"
	"time"

	"github.com/davecgh/go-spew/spew"
)

//----------------------------------------
// cdc.decodeReflectBinary

// This is the main entrypoint for decoding all types from binary form.  This
// function calls decodeReflectBinary*, and generally those functions should
// only call this one, for the prefix bytes are consumed here when present.
// CONTRACT: rv.CanAddr() is true.
func (cdc *Codec) decodeReflectBinary(br *bufio.Reader, info *TypeInfo, rv reflect.Value, opts FieldOptions) (n int, err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}
	if info.Type.Kind() == reflect.Interface && rv.Kind() == reflect.Ptr {
		panic("should not happen")
	}

	if printLog {
		spew.Printf("(d) decodeReflectBinary(info: %v, rv: %#v (%v), opts: %v)\n",
			info, rv.Interface(), rv.Type(), opts)
		defer func() {
			fmt.Printf("(d) -> n: %v, err: %v\n", n, err)
		}()
	}

	// TODO Read the disamb bytes here if necessary.
	// e.g. rv isn't an interface, and
	// info.ConcreteType.AlwaysDisambiguate.  But we don't support
	// this yet.

	if !info.Registered {
		// No need for disambiguation, decode as is.
		n, err = cdc._decodeReflectBinary(br, info, rv, opts)
		return
	}

	var _n int
	var bz []byte
	_n, err = peekConsumeDiscard(br, PrefixBytesLen, func(bzz []byte) (int, error) {
		bz = bzz
		return len(bzz), nil
	})
	n += _n
	if err != nil {
		return
	}
	if len(bz) < PrefixBytesLen {
		err = errors.New("EOF skipping prefix bytes.")
		return
	}
	if !info.Prefix.EqualBytes(bz) {
		panic("should not happen")
	}

	_n, err = cdc._decodeReflectBinary(br, info, rv, opts)
	n += _n
	return
}

// CONTRACT: any immediate disamb/prefix bytes have been consumed/stripped.
// CONTRACT: rv.CanAddr() is true.
func (cdc *Codec) _decodeReflectBinary(br *bufio.Reader, info *TypeInfo, rv reflect.Value, opts FieldOptions) (n int, err error) {

	// TODO consider the binary equivalent of json.Unmarshaller.

	// If a pointer, handle pointer byte.
	// 0x00 means nil, 0x01 means not nil.
	if rv.Kind() == reflect.Ptr {
		b0, rErr := br.ReadByte()
		if rErr != nil {
			err = fmt.Errorf("reading pointer type: %v", rErr)
			return
		}
		switch b0 {
		case 0x00:
			n += 1
			rv.Set(reflect.Zero(rv.Type()))
			return
		case 0x01:
			n += 1
			// so continue...
		default:
			err = fmt.Errorf("unexpected pointer byte %X", b0)
			return
		}
	}

	// Dereference-and-construct pointers all the way.
	// This works for pointer-pointers.
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			newPtr := reflect.New(rv.Type().Elem())
			rv.Set(newPtr)
		}
		rv = rv.Elem()
	}

	var _n int
	switch info.Type.Kind() {

	//----------------------------------------
	// Complex

	case reflect.Interface:
		_n, err = cdc.decodeReflectBinaryInterface(br, info, rv, opts)
		n += _n
		return

	case reflect.Array:
		_n, err = cdc.decodeReflectBinaryArray(br, info, rv, opts)
		n += _n
		return

	case reflect.Slice:
		_n, err = cdc.decodeReflectBinarySlice(br, info, rv, opts)
		n += _n
		return

	case reflect.Struct:
		_n, err = cdc.decodeReflectBinaryStruct(br, info, rv, opts)
		n += _n
		return

	//----------------------------------------
	// Signed

	case reflect.Int64:
		if opts.BinVarint {
			_n, err = peekConsumeDiscard(br, binary.MaxVarintLen32, func(bz []byte) (int, error) {
				num, _n, err := DecodeVarint(bz)
				if err != nil {
					return 0, err
				}
				rv.SetInt(num)
				return _n, nil
			})
			n += _n
			if err != nil {
				return
			}
		} else {
			var num int64
			num, _n, err = peekConsumeDiscardInt64(br)
			if err == nil {
				rv.SetInt(num)
				n += _n
			}
			return
		}

		// End of reflect.Int64
		return

	case reflect.Int32:
		_n, err = peekConsumeDiscard(br, binary.MaxVarintLen32, func(bz []byte) (int, error) {
			num, _n, err := DecodeInt32(bz)
			if err != nil {
				return 0, err
			}
			rv.SetInt(int64(num))
			return _n, nil
		})
		n += _n

		// End of reflect.Int32
		return

	case reflect.Int16:
		_n, err = peekConsumeDiscard(br, binary.MaxVarintLen16, func(bz []byte) (int, error) {
			num, _n, err := DecodeInt16(bz)
			if err != nil {
				return 0, err
			}
			rv.SetInt(int64(num))
			return _n, nil
		})
		n += _n

		// End of reflect.Int16
		return

	case reflect.Int8:
		_n, err = peekConsumeDiscard(br, 1, func(bz []byte) (int, error) {
			num, _n, err := DecodeInt8(bz)
			if err != nil {
				return 0, err
			}
			rv.SetInt(int64(num))
			return _n, nil
		})
		n += _n

		// End of reflect.Int8
		return

	case reflect.Int:
		_n, err = peekConsumeDiscard(br, binary.MaxVarintLen32, func(bz []byte) (int, error) {
			num, _n, err := DecodeVarint(bz)
			if err != nil {
				return 0, err
			}
			rv.SetInt(num)
			return _n, nil
		})
		n += _n

		// End of reflect.Int
		return

	//----------------------------------------
	// Unsigned

	case reflect.Uint64:
		if opts.BinVarint {
			_n, err = peekConsumeDiscard(br, binary.MaxVarintLen32, func(bz []byte) (int, error) {
				num, _n, err := DecodeUvarint(bz)
				if err != nil {
					return 0, err
				}
				rv.SetUint(num)
				return _n, nil
			})
			n += _n

		} else {
			_n, err = peekConsumeDiscard(br, 8, func(bz []byte) (int, error) {
				num, _n, err := DecodeUint64(bz)
				if err != nil {
					return 0, err
				}
				rv.SetUint(num)
				return _n, nil
			})
			n += _n

		}

		// End of reflect.Uint64
		return

	case reflect.Uint32:
		_n, err = peekConsumeDiscard(br, 4, func(bz []byte) (int, error) {
			num, _n, err := DecodeUint32(bz)
			if err != nil {
				return 0, err
			}
			rv.SetUint(uint64(num))
			return _n, nil
		})
		n += _n

		// End of reflect.Uint32
		return

	case reflect.Uint16:
		_n, err = peekConsumeDiscard(br, binary.MaxVarintLen32, func(bz []byte) (int, error) {
			num, _n, err := DecodeUint16(bz)
			if err != nil {
				return 0, err
			}
			rv.SetUint(uint64(num))
			return _n, nil
		})
		n += _n

		// End of reflect.Uint16
		return

	case reflect.Uint8:
		_n, err = peekConsumeDiscard(br, binary.MaxVarintLen32, func(bz []byte) (int, error) {
			num, _n, err := DecodeUint8(bz)
			if err != nil {
				return 0, err
			}
			rv.SetUint(uint64(num))
			return _n, nil
		})
		n += _n

		// End of reflect.Uint8
		return

	case reflect.Uint:
		_n, err = peekConsumeDiscard(br, binary.MaxVarintLen32, func(bz []byte) (int, error) {
			num, _n, err := DecodeUvarint(bz)
			if err != nil {
				return 0, err
			}
			rv.SetUint(num)
			return _n, nil
		})
		n += _n

		// End of reflect.Uint
		return

	//----------------------------------------
	// Misc.

	case reflect.Bool:
		_n, err = peekConsumeDiscard(br, 1, func(bz []byte) (int, error) {
			b, _n, err := DecodeBool(bz)
			if err != nil {
				return 0, err
			}
			rv.SetBool(b)
			return _n, nil
		})
		n += _n
		// End of reflect.Bool
		return

	case reflect.Float64:
		var f float64
		if !opts.Unsafe {
			err = errors.New("Float support requires `wire:\"unsafe\"`.")
			return
		}
		_n, err = peekConsumeDiscard(br, 8, func(bz []byte) (int, error) {
			f, _n, err = DecodeFloat64(bz)
			if err != nil {
				return 0, err
			}
			rv.SetFloat(f)
			return _n, nil
		})
		n += _n

		// End of reflect.Float64
		return

	case reflect.Float32:
		var f float32
		if !opts.Unsafe {
			err = errors.New("Float support requires `wire:\"unsafe\"`.")
			return
		}
		_n, err = peekConsumeDiscard(br, 4, func(bz []byte) (n int, err error) {
			f, n, err = DecodeFloat32(bz)
			if err != nil {
				return 0, err
			}
			rv.SetFloat(float64(f))
			return
		})
		n += _n

		// End of reflect.Float32
		return

	case reflect.String:
		var str string
		str, _n, err = peekConsumeDiscardString(br)
		if err == nil {
			rv.SetString(str)
		}
		n += _n

		// End of reflect.String
		return

	default:
		panic(fmt.Sprintf("unknown field type %v", info.Type.Kind()))
	}

}

func peekConsumeDiscard(br *bufio.Reader, maxPeek int, consumeBytes func([]byte) (int, error)) (int, error) {
	bz, err := br.Peek(maxPeek)
	if err != nil {
		return 0, err
	}

	n, err := consumeBytes(bz)
	if err != nil {
		return 0, err
	}

	// Now discard exactly the peeked number of bytes.
	dn, err := br.Discard(n)
	if err != nil {
		return 0, err
	}
	if dn != n {
		return 0, fmt.Errorf("peekConsumeDiscard %d wanted %d", dn, n)
	}
	return dn, nil
}

func peekConsumeDiscardString(br *bufio.Reader) (string, int, error) {
	bz, n, err := peekConsumeDiscardByteSlice(br)
	if err != nil {
		return "", 0, err
	}
	// TODO: (@odeke-em) figure out how to send over a string
	// using the slice's underlying header perhaps, to avoid
	// an extraneous string<-->[]byte allocation.
	return string(bz), n, nil
}

func peekConsumeDiscardByteSlice(br *bufio.Reader) (bz []byte, n int, err error) {
	// 1. Firstly read out the byte slice length
	var length int64
	// The length is encoded as a Varint
	n, err = peekConsumeDiscard(br, binary.MaxVarintLen32, func(bz []byte) (nn int, err error) {
		length, nn, err = DecodeVarint(bz)
		return
	})
	if err != nil {
		return
	}

	// 2 Validate the length
	// 2.1: Check if negative.
	if g, w := length, int64(-1); g <= w {
		err = fmt.Errorf("possible underflow trying to make []byte got = %d want > %d", g, w)
		return
	}

	// 2.2 Validate against extraneous string allocations i.e. length >= maxInt32.
	// See https://github.com/tendermint/go-wire/pull/38
	if g, w := length, int64(math.MaxInt32); g > w {
		err = fmt.Errorf("possible overflow trying to make []byte got = %d want <= %d", g, w)
		return
	}

	var _n int
	_n, err = peekConsumeDiscard(br, int(length), func(bzz []byte) (int, error) {
		bz = bzz
		return len(bzz), nil
	})
	n += _n
	return
}

// CONTRACT: rv.CanAddr() is true.
func (cdc *Codec) decodeReflectBinaryInterface(br *bufio.Reader, iinfo *TypeInfo, rv reflect.Value, opts FieldOptions) (n int, err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}
	if !rv.IsNil() {
		// JAE: Heed this note, this is very tricky.
		err = errors.New("Decoding to a non-nil interface is not supported yet")
		return
	}

	// Peek disambiguation / prefix info.
	disfix, hasDisamb, prefix, hasPrefix, isNil, _n, err := decodeDisambPrefixBytesWithReader(br)
	if err != nil {
		return
	}

	// Special case for nil.
	if isNil {
		n += 1 + DisambBytesLen // Consume 0x{00 00 00 00}
		rv.Set(iinfo.ZeroValue)
		return
	}

	if hasDisamb {
		n += 1 + DisfixBytesLen
	} else {
		n += PrefixBytesLen
	}

	// Get concrete type info.
	var cinfo *TypeInfo
	if hasDisamb {
		cinfo, err = cdc.getTypeInfoFromDisfix_rlock(disfix)
	} else if hasPrefix {
		cinfo, err = cdc.getTypeInfoFromPrefix_rlock(iinfo, prefix)
	} else {
		err = errors.New("Expected disambiguation or prefix bytes.")
	}
	if err != nil {
		return
	}

	// Construct the concrete type.
	var crv, irvSet = constructConcreteType(cinfo)

	// Decode into the concrete type.
	_n, err = cdc._decodeReflectBinary(br, cinfo, crv, opts)
	if err != nil {
		rv.Set(irvSet) // Helps with debugging
		return
	}
	n += _n

	// We need to set here, for when !PointerPreferred and the type
	// is say, an array of bytes (e.g. [32]byte), then we must call
	// rv.Set() *after* the value was acquired.
	// NOTE: rv.Set() should succeed because it was validated
	// already during Register[Interface/Concrete].
	rv.Set(irvSet)
	return
}

// CONTRACT: rv.CanAddr() is true.
func (cdc *Codec) decodeReflectBinaryArray(br *bufio.Reader, info *TypeInfo, rv reflect.Value, opts FieldOptions) (n int, err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}
	ert := info.Type.Elem()
	length := info.Type.Len()
	_n := 0

	switch ert.Kind() {

	case reflect.Uint8: // Special case: byte array
		var bz []byte
		_n, err = peekConsumeDiscard(br, length, func(bzz []byte) (int, error) {
			bz = bzz
			return len(bzz), nil
		})
		n += _n
		if err != nil {
			return
		}

		if len(bz) < length {
			return 0, fmt.Errorf("Insufficient bytes to decode [%v]byte.", length)
		}
		reflect.Copy(rv, reflect.ValueOf(bz[0:length]))
		return

	default: // General case.
		var einfo *TypeInfo
		einfo, err = cdc.getTypeInfo_wlock(ert)
		if err != nil {
			return
		}
		for i := 0; i < length; i++ {
			erv := rv.Index(i)
			_n, err = cdc.decodeReflectBinary(br, einfo, erv, opts)
			if err != nil {
				return
			}
			n += _n
		}
		return
	}
}

// CONTRACT: rv.CanAddr() is true.
func (cdc *Codec) decodeReflectBinarySlice(br *bufio.Reader, info *TypeInfo, rv reflect.Value, opts FieldOptions) (n int, err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}
	ert := info.Type.Elem()
	_n := 0 // nolint: ineffassign

	switch ert.Kind() {

	case reflect.Uint8: // Special case: byte slice
		var byteslice []byte
		byteslice, _n, err = peekConsumeDiscardByteSlice(br)
		if err != nil {
			return
		}
		n += _n
		if len(byteslice) == 0 {
			// Special case when length is 0.
			// NOTE: We prefer nil slices.
			rv.Set(info.ZeroValue)
		} else {
			rv.Set(reflect.ValueOf(byteslice))
		}
		return

	default: // General case.
		var einfo *TypeInfo
		einfo, err = cdc.getTypeInfo_wlock(ert)
		if err != nil {
			return
		}

		// Read length.
		var length64 int64
		length64, _n, err = peekConsumeDiscardVarInt(br)
		if err != nil {
			return
		}

		if length64 < 0 {
			err = errors.New("Invalid negative slice length")
			return
		}
		n += _n

		// Special case when length is 0.
		// NOTE: We prefer nil slices.
		if length64 == 0 {
			rv.Set(info.ZeroValue)
			return
		}

		length := int(length64)
		// Read into a new slice.
		var esrt = reflect.SliceOf(ert) // TODO could be optimized.
		var srv = reflect.MakeSlice(esrt, length, length)
		for i := 0; i < length; i++ {
			erv := srv.Index(i)
			_n, err = cdc.decodeReflectBinary(br, einfo, erv, opts)
			if err != nil {
				return
			}
			n += _n
		}

		// TODO do we need this extra step?
		rv.Set(srv)
		return
	}
}

func peekConsumeDiscardTime(br *bufio.Reader) (*time.Time, int, error) {
	var n int
	s, _n, err := peekConsumeDiscardInt64(br)
	if err != nil {
		return nil, 0, err
	}
	n += _n
	ns, _n, err := peekConsumeDiscardInt32(br)
	if err != nil {
		return nil, n, err
	}
	n += _n
	if ns < 0 || 999999999 < ns {
		return nil, n, fmt.Errorf("Invalid time, nanoseconds out of bounds %v", ns)
	}
	t := time.Unix(s, int64(ns))
	// strip timezone and monotonic for deep equality
	t = t.UTC().Truncate(0)
	return &t, n, nil
}

func peekConsumeDiscardInt64(br *bufio.Reader) (i64 int64, nRead int, err error) {
	nRead, err = peekConsumeDiscard(br, binary.MaxVarintLen64, func(bz []byte) (n int, err error) {
		i64, n, err = DecodeInt64(bz)
		return
	})
	return
}

func peekConsumeDiscardVarInt(br *bufio.Reader) (i32 int64, nRead int, err error) {
	nRead, err = peekConsumeDiscard(br, binary.MaxVarintLen32, func(bz []byte) (n int, err error) {
		i32, n, err = DecodeVarint(bz)
		return
	})
	return
}

const int32Size = 4

func peekConsumeDiscardInt32(br *bufio.Reader) (i32 int32, nRead int, err error) {
	nRead, err = peekConsumeDiscard(br, int32Size, func(bz []byte) (int, error) {
		if len(bz) < int32Size {
			return 0, errors.New("EOF decoding int32")
		}
		i32 = int32(binary.BigEndian.Uint32(bz[:int32Size]))
		return len(bz), nil
	})
	return
}

// CONTRACT: rv.CanAddr() is true.
func (cdc *Codec) decodeReflectBinaryStruct(br *bufio.Reader, info *TypeInfo, rv reflect.Value, _ FieldOptions) (n int, err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}
	_n := 0 // nolint: ineffassign

	switch info.Type {

	case timeType: // Special case: time.Time
		var t *time.Time
		t, _n, err = peekConsumeDiscardTime(br)
		if err != nil {
			return
		}
		n += _n
		rv.Set(reflect.ValueOf(*t))
		return

	default:
		for _, field := range info.Fields {

			// Get field rv and info.
			var frv = rv.Field(field.Index)
			var finfo *TypeInfo
			finfo, err = cdc.getTypeInfo_wlock(field.Type)
			if err != nil {
				return
			}

			// Decode into field rv.
			_n, err = cdc.decodeReflectBinary(br, finfo, frv, field.FieldOptions)
			n += _n
			if err != nil {
				return
			}
		}
		return
	}
}

func decodeDisambPrefixBytesWithReader(br *bufio.Reader) (df DisfixBytes, hasDb bool, pb PrefixBytes, hasPb bool, isNil bool, n int, err error) {
	var bz []byte
	var _n int
	_n, err = peekConsumeDiscard(br, 4, func(bzz []byte) (int, error) {
		bz = bzz
		return len(bzz), nil
	})
	if err != nil {
		return
	}
	n += _n

	df, hasDb, pb, hasPb, isNil, _n, err = decodeDisambPrefixBytes(bz)
	n += _n
	return
}

//----------------------------------------

func decodeDisambPrefixBytes(bz []byte) (df DisfixBytes, hasDb bool, pb PrefixBytes, hasPb bool, isNil bool, n int, err error) {
	// Validate
	if len(bz) < 4 {
		err = errors.New("EOF reading prefix bytes.")
		return // hasPb = false
	}
	if bz[0] == 0x00 {
		// Special case: nil
		if (DisambBytes{}).EqualBytes(bz[1:4]) {
			isNil = true
			n = 4
			return
		}
		// Validate
		if len(bz) < 8 {
			err = errors.New("EOF reading disamb bytes.")
			return // hasPb = false
		}
		copy(df[0:7], bz[1:8])
		copy(pb[0:4], bz[4:8])
		hasDb = true
		hasPb = true
		n = 8
		return
	} else {
		// General case with no disambiguation
		copy(pb[0:4], bz[0:4])
		hasDb = false
		hasPb = true
		n = 4
		return
	}
}
