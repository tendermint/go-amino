package main

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"os"

	"github.com/tendermint/go-amino"
	cmn "github.com/tendermint/tmlibs/common"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println(`Usage: aminoscan <STRUCT HEXBYTES>`) // TODO support more options, including support for framing.
		return
	}
	bz := hexDecode(os.Args[1])             // Read input hex bytes.
	s, n, err := scanStruct(bz, "")         // Assume that it's  struct.
	s += cmn.Red(fmt.Sprintf("%X", bz[n:])) // Bytes remaining are red.
	fmt.Println(s, n, err)                  // Print color-encoded bytes s.
}

func scanAny(typ amino.Typ3, bz []byte, indent string) (term bool, s string, n int, err error) {
	switch typ {
	case amino.Typ3_Varint:
		s, n, err = scanVarint(bz, indent)
	case amino.Typ3_8Byte:
		s, n, err = scan8Byte(bz, indent)
	case amino.Typ3_ByteLength:
		s, n, err = scanByteLength(bz, indent)
	case amino.Typ3_Struct:
		s, n, err = scanStruct(bz, indent)
	case amino.Typ3_StructTerm:
		term = true
	case amino.Typ3_4Byte:
		s, n, err = scan4Byte(bz, indent)
	case amino.Typ3_List:
		s, n, err = scanList(bz, indent)
	case amino.Typ3_Interface:
		s, n, err = scanInterface(bz, indent)
	default:
		panic("should not happen")
	}
	return
}

func scanVarint(bz []byte, indent string) (s string, n int, err error) {
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
	fmt.Printf("%s%s (", indent, s)
	if okI64 {
		fmt.Printf("i64:%v ", i64)
	}
	if okU64 {
		fmt.Printf("u64:%v", u64)
	}
	fmt.Print(")\n")
	return
}

func scan8Byte(bz []byte, indent string) (s string, n int, err error) {
	if len(bz) < 8 {
		err = errors.New("EOF reading 8byte field.")
		return
	}
	n = 8
	s = cmn.Blue(fmt.Sprintf("%X", bz[:8]))
	fmt.Printf("%s%s\n", indent, s)
	return
}

func scanByteLength(bz []byte, indent string) (s string, n int, err error) {
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
	fmt.Printf("%s%s (%v bytes)\n", indent, s, length)
	return
}

func scanStruct(bz []byte, indent string) (s string, n int, err error) {
	var _s, _n, typ = string(""), int(0), amino.Typ3(0x00)
FOR_LOOP:
	for {
		_s, typ, _n, err = scanFieldKey(bz, indent+"  ")
		if slide(&bz, &n, _n) && concat(&s, _s) && err != nil {
			return
		}
		var term bool
		term, _s, _n, err = scanAny(typ, bz, indent+"  ")
		if slide(&bz, &n, _n) && concat(&s, _s) && err != nil {
			return
		}
		if term {
			break FOR_LOOP
		}
	}
	return
}

func scanFieldKey(bz []byte, indent string) (s string, typ amino.Typ3, n int, err error) {
	var u64 uint64
	u64, n = binary.Uvarint(bz)
	if n < 0 {
		n = 0
		err = errors.New("error decoding uvarint")
		return
	}
	typ = amino.Typ3(u64 & 0x07)
	var number uint32 = uint32(u64 >> 3)
	s = fmt.Sprintf("%X", bz[:n])
	fmt.Printf("%s%s @%v %v\n", indent, s, number, typ)
	return
}

func scan4Byte(bz []byte, indent string) (s string, n int, err error) {
	if len(bz) < 4 {
		err = errors.New("EOF reading 4byte field.")
		return
	}
	n = 4
	s = cmn.Blue(fmt.Sprintf("%X", bz[:4]))
	fmt.Printf("%s%s\n", indent, s)
	return
}

func scanList(bz []byte, indent string) (s string, n int, err error) {
	// Read element Typ4.
	if len(bz) < 1 {
		err = errors.New("EOF reading list element typ4.")
		return
	}
	var typ = amino.Typ4(bz[0])
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
	fmt.Printf("%s%s of %v with %v items\n", indent, s, typ, num)
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
				fmt.Printf("%s00 (not nil)\n", indent)
			case 0x01:
				s += "01" // Is nil (NOTE: reverse logic)
				fmt.Printf("%s01 (is nil)\n", indent)
				continue
			default:
				err = fmt.Errorf("Unexpected nil pointer byte %X", nb)
				return
			}
		}
		// Read element.
		_, _s, _n, err = scanAny(typ.Typ3(), bz, indent+"  ")
		if slide(&bz, &n, _n) && concat(&s, _s) && err != nil {
			return
		}
	}
	return
}

func scanInterface(bz []byte, indent string) (s string, n int, err error) {
	db, hasDb, pb, typ, _, isNil, _n, err := amino.DecodeDisambPrefixBytes(bz)
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
		fmt.Printf("%s%s (nil interface)\n", indent, s)
	} else if hasDb {
		fmt.Printf("%s%s (disamb: %X, prefix: %X, typ: %v)\n",
			indent, s, db.Bytes(), pb.Bytes(), typ)
	} else {
		fmt.Printf("%s%s (prefix: %X, typ: %v)\n",
			indent, s, pb.Bytes(), typ)
	}
	_, _s, _n, err := scanAny(typ, bz, indent)
	if slide(&bz, &n, _n) && concat(&s, _s) && err != nil {
		return
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
