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
	bz := hexDecode(os.Args[1])             // Read input hex bytes.
	s, n, err := scanStruct(bz)             // Assume that it's  struct.
	s += cmn.Red(fmt.Sprintf("%X", bz[n:])) // Bytes remaining are red.
	fmt.Println(s, n, err)                  // Print color-encoded bytes s.
}

func scanAny(typ wire.Typ3, bz []byte) (stop bool, s string, n int, err error) {
	switch typ {
	case wire.Typ3_Varint:
		s, n, err = scanVarint(bz)
	case wire.Typ3_8Byte:
		s, n, err = scan8Byte(bz)
	case wire.Typ3_ByteLength:
		s, n, err = scanByteLength(bz)
	case wire.Typ3_Struct:
		s, n, err = scanStruct(bz)
	case wire.Typ3_StructTerm:
		stop = true
	case wire.Typ3_4Byte:
		s, n, err = scan4Byte(bz)
	case wire.Typ3_List:
		s, n, err = scanList(bz)
	case wire.Typ3_Interface:
		s, n, err = scanInterface(bz)
	default:
		panic("should not happen")
	}
	return
}

func scanVarint(bz []byte) (s string, n int, err error) {
	if len(bz) == 0 {
		err = fmt.Errorf("EOF reading Varint")
	}
	// First try Varint.
	var i64, okI64 = int64(0), true
	i64, n = binary.Varint(bz)
	if n <= 0 {
		n = 0
		okI64 = false
	}
	// Then try Uvarint.
	var u64, okU64, _n = uint64(0), true, int(0)
	u64, _n = binary.Uvarint(bz)
	if n != _n {
		n = 0
		okU64 = false
	}
	// If neither work, return error.
	if !okI64 && !okU64 {
		err = fmt.Errorf("Invalid (u)varint")
		return
	}
	// s is the same either way.
	s = cmn.Cyan(fmt.Sprintf("%X", bz[:n]))
	fmt.Printf("%v (", s)
	if okI64 {
		fmt.Printf("i64:%v ", i64)
	}
	if okU64 {
		fmt.Printf("u64:%v", u64)
	}
	fmt.Print(")\n")
	return
}

func scan8Byte(bz []byte) (s string, n int, err error) {
	if len(bz) < 8 {
		err = errors.New("EOF reading 8byte field.")
		return
	}
	n = 8
	s = cmn.Blue(fmt.Sprintf("%X", bz[:8]))
	fmt.Printf("%v\n", s)
	return
}

func scanByteLength(bz []byte) (s string, n int, err error) {
	// Read the length.
	var length, l64, _n = int(0), uint64(0), int(0)
	l64, _n = binary.Uvarint(bz)
	if n < 0 {
		n = 0
		err = errors.New("error decoding uvarint")
		return
	}
	length = int(l64)
	if length >= len(bz) {
		err = errors.New("EOF reading byte-length delimited.")
		return
	}
	s = cmn.Cyan(fmt.Sprintf("%X", bz[:_n]))
	slide(&bz, &n, _n)
	// Read the remaining bytes.
	s += cmn.Green(fmt.Sprintf("%X", bz[:length]))
	slide(&bz, &n, length)
	fmt.Printf("%v (%v bytes)\n", s, length)
	return
}

func scanStruct(bz []byte) (s string, n int, err error) {
	var _s, _n, typ = string(""), int(0), wire.Typ3(0x00)
FOR_LOOP:
	for {
		_s, typ, _n, err = scanFieldKey(bz)
		if slide(&bz, &n, _n) && concat(&s, _s) && err != nil {
			return
		}
		var stop bool
		stop, _s, _n, err = scanAny(typ, bz)
		if slide(&bz, &n, _n) && concat(&s, _s) && err != nil {
			return
		}
		if stop {
			break FOR_LOOP
		}
	}
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
	fmt.Printf("%v @%v %v\n", s, number, typ)
	return
}

func scan4Byte(bz []byte) (s string, n int, err error) {
	if len(bz) < 4 {
		err = errors.New("EOF reading 4byte field.")
		return
	}
	n = 4
	s = cmn.Blue(fmt.Sprintf("%X", bz[:4]))
	fmt.Printf("%v\n", s)
	return
}

func scanList(bz []byte) (s string, n int, err error) {
	// Read element Typ4.
	if len(bz) < 1 {
		err = errors.New("EOF reading list element typ4.")
		return
	}
	var typ = wire.Typ4(bz[0])
	if typ&0xF0 > 0 {
		err = errors.New("Invalid list element typ4 byte")
	}
	s = fmt.Sprintf("%X", bz[:1])
	if slide(&bz, &n, 1) && err != nil {
		return
	}
	// Read number of elements.
	var num, _n = uint64(0), int(0)
	num, _n = binary.Uvarint(bz)
	if _n < 0 {
		_n = 0
		err = errors.New("error decoding list length (uvarint)")
	}
	s += cmn.Cyan(fmt.Sprintf("%X", bz[:_n]))
	if slide(&bz, &n, _n) && err != nil {
		return
	}
	fmt.Printf("%v of %v with %v items\n", s, typ, num)
	// Read elements.
	var _s string
	for i := 0; i < int(num); i++ {
		// Maybe read nil byte.
		if typ&0x08 != 0 {
			if len(bz) == 0 {
				err = errors.New("EOF reading list nil byte")
				return
			}
			var nb = bz[0]
			slide(&bz, &n, 1)
			switch nb {
			case 0x00:
				s += "00"
				fmt.Printf("00 (not nil)\n")
			case 0x01:
				s += "01" // Is nil (NOTE: reverse logic)
				fmt.Printf("01 (is nil)\n")
				continue
			default:
				err = fmt.Errorf("Unexpected nil pointer byte %X", nb)
				return
			}
		}
		// Read element.
		_, _s, _n, err = scanAny(typ.Typ3(), bz)
		if slide(&bz, &n, _n) && concat(&s, _s) && err != nil {
			return
		}
	}
	return
}

func scanInterface(bz []byte) (s string, n int, err error) {
	db, hasDb, pb, typ, _, isNil, _n, err := wire.DecodeDisambPrefixBytes(bz)
	if slide(&bz, &n, _n) && err != nil {
		return
	}
	pb3 := pb.WithTyp3(typ)
	if isNil {
		s = cmn.Magenta("0000")
	} else if hasDb {
		s = cmn.Magenta(fmt.Sprintf("%X%X", db.Bytes(), pb3.Bytes()))
	} else {
		s = cmn.Magenta(fmt.Sprintf("%X", pb3.Bytes()))
	}
	if isNil {
		fmt.Printf("%v (nil interface)\n", s)
	} else if hasDb {
		fmt.Printf("%v (disamb: %X, prefix: %X, typ: %v)\n",
			s, db.Bytes(), pb.Bytes(), typ)
	} else {
		fmt.Printf("%v (prefix: %X, typ: %v)\n",
			s, pb.Bytes(), typ)
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
