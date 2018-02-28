package main

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"os"

	"github.com/tendermint/go-wire"
	cmn "github.com/tendermint/tmlibs/common"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("Usage: wirescan <HEXBYTES>")
		return
	}
	bz := hexDecode(os.Args[1])
	s, n, err := scanStruct(bz)
	s += cmn.Red(fmt.Sprintf("%X", bz[n:]))
	fmt.Println(s, n, err)
}

func scanVarint(bz []byte) (s string, n int, err error) {
	// First try Varint.
	var i64 int64
	i64, n = binary.Varint(bz)
	if n < 0 {
		n = 0
		err = errors.New("error decoding varint")
		return
	}
	// s is the same either way.
	s = cmn.Green(fmt.Sprintf("%X", bz[:n]))
	// Also print Uvarint.
	var u64, _n = uint64(0), int(0)
	u64, _n = binary.Uvarint(bz)
	if n != _n {
		n = 0
		err = errors.New("error decoding varint")
		return
	}
	fmt.Printf("%v i64:%v u64:%v\n", s, i64, u64)
	return
}

func scan8Byte(bz []byte) (s string, n int, err error) {
	if len(bz) < 8 {
		err = errors.New("EOF reading 8byte field.")
		return
	}
	n = 8
	s = cmn.Cyan(fmt.Sprintf("%X", bz[:8]))
	fmt.Printf("%v\n", s)
	return
}

func scanByteLength(bz []byte) (s string, n int, err error) {
	// Read the length
	var length, l64, _n = int(0), uint64(0), int(0)
	l64, _n = binary.Uvarint(bz)
	if n != _n {
		n = 0
		err = errors.New("error decoding varint")
		return
	}
	slide(&bz, &n, _n)
	length = int(l64)
	if length >= len(bz) {
		err = errors.New("EOF reading byte-length delimited.")
		return
	}
	// Read the remaining bytes
	s = cmn.Yellow(fmt.Sprintf("%X", bz[:length]))
	fmt.Printf("%v (%v bytes)\n", s, length)
	return
}

func scanStruct(bz []byte) (s string, n int, err error) {
	fmt.Println("Start Struct")
	var _s, _n, typ = string(""), int(0), wire.Typ3(0x00)
FOR_LOOP:
	for {
		_s, typ, _n, err = scanFieldKey(bz)
		if slide(&bz, &n, _n) && concat(&s, _s) && err != nil {
			return
		}
		switch typ {
		case wire.Typ3_Varint:
			_s, _n, err = scanVarint(bz)
		case wire.Typ3_8Byte:
			_s, _n, err = scan8Byte(bz)
		case wire.Typ3_ByteLength:
			_s, _n, err = scanByteLength(bz)
		case wire.Typ3_Struct:
			_s, _n, err = scanStruct(bz)
		case wire.Typ3_StructTerm:
			break FOR_LOOP
		case wire.Typ3_4Byte:
			_s, _n, err = scan4Byte(bz)
		case wire.Typ3_List:
			_s, _n, err = scanList(bz)
		case wire.Typ3_Interface:
			_s, _n, err = scanInterface(bz)
		default:
			panic("should not happen")
		}
		if slide(&bz, &n, _n) && concat(&s, _s) && err != nil {
			return
		}
	}
	fmt.Println("End Struct")
	return
}

func scanFieldKey(bz []byte) (s string, typ wire.Typ3, n int, err error) {
	var u64 uint64
	u64, n = binary.Uvarint(bz)
	if n < 0 {
		n = 0
		err = errors.New("error decoding uvarint")
		return
	}
	typ = wire.Typ3(u64 & 0x07)
	var number uint32 = uint32(u64 >> 3)
	s = fmt.Sprintf("%X", bz[:n])
	fmt.Printf("%v @%v #%X\n", s, number, typ)
	return
}

func scan4Byte(bz []byte) (s string, n int, err error) {
	if len(bz) < 4 {
		err = errors.New("EOF reading 4byte field.")
		return
	}
	n = 4
	s = cmn.Cyan(fmt.Sprintf("%X", bz[:8]))
	fmt.Printf("%v\n", s)
	return
}

func scanList(bz []byte) (s string, n int, err error) {
	// Read element Typ4
	if len(bz) < 1 {
		err = errors.New("EOF reading 4byte field.")
		return
	}
	var typ = wire.Typ4(bz[0])
	if typ&0xF0 > 0 {
		err = errors.New("Invalid list element typ4 byte")
	}
	if slide(&bz, &n, 1) && err != nil {
		return
	}
	// Read number elements
	var num, _n = uint64(0), int(0)
	num, _n = binary.Uvarint(bz)
	if _n < 0 {
		_n = 0
		err = errors.New("error decoding list length (uvarint)")
	}
	if slide(&bz, &n, _n) && err != nil {
		return
	}
	s = cmn.Yellow(fmt.Sprintf("%X", bz[:n]))
	fmt.Printf("%v (%v #%X)\n", s, num, typ)
	return
}

func scanInterface(bz []byte) (s string, n int, err error) {
	df, hasDb, pb, _, isNil, _n, err := wire.DecodeDisambPrefixBytes(bz)
	if slide(&bz, &n, _n) && err != nil {
		return
	}
	s = cmn.Magenta(fmt.Sprintf("%X%X", df.Bytes(), pb.Bytes()))
	if isNil {
		fmt.Printf("%v (nil interface)\n", s)
	} else if hasDb {
		fmt.Printf("%v (disamb: %X, prefix: %X, typ: #%X)\n",
			s, df.Bytes(), pb.Bytes(), pb.Typ3())
	} else {
		fmt.Printf("%v (prefix: %X, typ: #%X)\n",
			s, pb.Bytes(), pb.Typ3())
	}
	return
}

//----------------------------------------
// Misc.

func slide(bzPtr *[]byte, n *int, _n int) bool {
	if len(*bzPtr) < _n {
		panic("eof")
	}
	*bzPtr = (*bzPtr)[_n:]
	*n += _n
	return true
}

func concat(sPtr *string, _s string) bool {
	*sPtr += _s
	return true
}

func hexDecode(s string) []byte {
	bz, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return bz
}
