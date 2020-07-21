package tests

import "time"

//----------------------------------------
// Struct types

type EmptyStruct struct {
}

type PrimitivesStruct struct {
	Int8        int8
	Int16       int16
	Int32       int32
	Int32Fixed  int32 `binary:"fixed32"`
	Int64       int64
	Int64Fixed  int64 `binary:"fixed64"`
	Int         int
	Byte        byte
	Uint8       uint8
	Uint16      uint16
	Uint32      uint32
	Uint32Fixed uint32 `binary:"fixed32"`
	Uint64      uint64
	Uint64Fixed uint64 `binary:"fixed64"`
	Uint        uint
	Str         string
	Bytes       []byte
	Time        time.Time
	Empty       EmptyStruct
}

type ShortArraysStruct struct {
	TimeAr [0]time.Time
}

type ArraysStruct struct {
	Int8Ar        [4]int8
	Int16Ar       [4]int16
	Int32Ar       [4]int32
	Int32FixedAr  [4]int32 `binary:"fixed32"`
	Int64Ar       [4]int64
	Int64FixedAr  [4]int64 `binary:"fixed64"`
	IntAr         [4]int
	ByteAr        [4]byte
	Uint8Ar       [4]uint8
	Uint16Ar      [4]uint16
	Uint32Ar      [4]uint32
	Uint32FixedAr [4]uint32 `binary:"fixed32"`
	Uint64Ar      [4]uint64
	Uint64FixedAr [4]uint64 `binary:"fixed64"`
	UintAr        [4]uint
	StrAr         [4]string
	BytesAr       [4][]byte
	TimeAr        [4]time.Time
	EmptyAr       [4]EmptyStruct
}

type ArraysArraysStruct struct {
	Int8ArAr        [2][2]int8
	Int16ArAr       [2][2]int16
	Int32ArAr       [2][2]int32
	Int32FixedArAr  [2][2]int32 `binary:"fixed32"`
	Int64ArAr       [2][2]int64
	Int64FixedArAr  [2][2]int64 `binary:"fixed64"`
	IntArAr         [2][2]int
	ByteArAr        [2][2]byte
	Uint8ArAr       [2][2]uint8
	Uint16ArAr      [2][2]uint16
	Uint32ArAr      [2][2]uint32
	Uint32FixedArAr [2][2]uint32 `binary:"fixed32"`
	Uint64ArAr      [2][2]uint64
	Uint64FixedArAr [2][2]uint64 `binary:"fixed64"`
	UintArAr        [2][2]uint
	StrArAr         [2][2]string
	BytesArAr       [2][2][]byte
	TimeArAr        [2][2]time.Time
	EmptyArAr       [2][2]EmptyStruct
}

type SlicesStruct struct {
	Int8Sl        []int8
	Int16Sl       []int16
	Int32Sl       []int32
	Int32FixedSl  []int32 `binary:"fixed32"`
	Int64Sl       []int64
	Int64FixedSl  []int64 `binary:"fixed64"`
	IntSl         []int
	ByteSl        []byte
	Uint8Sl       []uint8
	Uint16Sl      []uint16
	Uint32Sl      []uint32
	Uint32FixedSl []uint32 `binary:"fixed32"`
	Uint64Sl      []uint64
	Uint64FixedSl []uint64 `binary:"fixed64"`
	UintSl        []uint
	StrSl         []string
	BytesSl       [][]byte
	TimeSl        []time.Time
	EmptySl       []EmptyStruct
}

type SlicesSlicesStruct struct {
	Int8SlSl        [][]int8
	Int16SlSl       [][]int16
	Int32SlSl       [][]int32
	Int32FixedSlSl  [][]int32 `binary:"fixed32"`
	Int64SlSl       [][]int64
	Int64FixedSlSl  [][]int64 `binary:"fixed64"`
	IntSlSl         [][]int
	ByteSlSl        [][]byte
	Uint8SlSl       [][]uint8
	Uint16SlSl      [][]uint16
	Uint32SlSl      [][]uint32
	Uint32FixedSlSl [][]uint32 `binary:"fixed32"`
	Uint64SlSl      [][]uint64
	Uint64FixedSlSl [][]uint64 `binary:"fixed64"`
	UintSlSl        [][]uint
	StrSlSl         [][]string
	BytesSlSl       [][][]byte
	TimeSlSl        [][]time.Time
	EmptySlSl       [][]EmptyStruct
}

type PointersStruct struct {
	Int8Pt        *int8
	Int16Pt       *int16
	Int32Pt       *int32
	Int32FixedPt  *int32 `binary:"fixed32"`
	Int64Pt       *int64
	Int64FixedPt  *int64 `binary:"fixed64"`
	IntPt         *int
	BytePt        *byte
	Uint8Pt       *uint8
	Uint16Pt      *uint16
	Uint32Pt      *uint32
	Uint32FixedPt *uint32 `binary:"fixed32"`
	Uint64Pt      *uint64
	Uint64FixedPt *uint64 `binary:"fixed64"`
	UintPt        *uint
	StrPt         *string
	BytesPt       *[]byte
	TimePt        *time.Time
	EmptyPt       *EmptyStruct
}

type PointerSlicesStruct struct {
	Int8PtSl        []*int8
	Int16PtSl       []*int16
	Int32PtSl       []*int32
	Int32FixedPtSl  []*int32 `binary:"fixed32"`
	Int64PtSl       []*int64
	Int64FixedPtSl  []*int64 `binary:"fixed64"`
	IntPtSl         []*int
	BytePtSl        []*byte
	Uint8PtSl       []*uint8
	Uint16PtSl      []*uint16
	Uint32PtSl      []*uint32
	Uint32FixedPtSl []*uint32 `binary:"fixed32"`
	Uint64PtSl      []*uint64
	Uint64FixedPtSl []*uint64 `binary:"fixed64"`
	UintPtSl        []*uint
	StrPtSl         []*string
	BytesPtSl       []*[]byte
	TimePtSl        []*time.Time
	EmptyPtSl       []*EmptyStruct
}

/* amino 1.2 removed nested pointer support.
// NOTE: See registered fuzz funcs for *byte, **byte, and ***byte.
type NestedPointersStruct struct {
	Ptr1 *byte
	Ptr2 **byte
	Ptr3 ***byte
}
*/

type ComplexSt struct {
	PrField PrimitivesStruct
	ArField ArraysStruct
	SlField SlicesStruct
	PtField PointersStruct
}

type EmbeddedSt1 struct {
	PrimitivesStruct
}

type EmbeddedSt2 struct {
	PrimitivesStruct
	ArraysStruct
	SlicesStruct
	PointersStruct
}

type EmbeddedSt3 struct {
	*PrimitivesStruct
	*ArraysStruct
	*SlicesStruct
	*PointersStruct
	*EmptyStruct
}

type EmbeddedSt4 struct {
	Foo1 int
	PrimitivesStruct
	Foo2              string
	ArraysStructField ArraysStruct
	Foo3              []byte
	SlicesStruct
	Foo4                bool
	PointersStructField PointersStruct
	Foo5                uint
}

type EmbeddedSt5 struct {
	Foo1 int
	*PrimitivesStruct
	Foo2              string
	ArraysStructField *ArraysStruct
	Foo3              []byte
	*SlicesStruct
	Foo4                bool
	PointersStructField *PointersStruct
	Foo5                uint
}

var StructTypes = []interface{}{
	(*EmptyStruct)(nil),
	(*PrimitivesStruct)(nil),
	(*ShortArraysStruct)(nil),
	(*ArraysStruct)(nil),
	(*ArraysArraysStruct)(nil),
	(*SlicesStruct)(nil),
	(*SlicesSlicesStruct)(nil),
	(*PointersStruct)(nil),
	(*PointerSlicesStruct)(nil),
	// (*NestedPointersStruct)(nil),
	(*ComplexSt)(nil),
	(*EmbeddedSt1)(nil),
	(*EmbeddedSt2)(nil),
	(*EmbeddedSt3)(nil),
	(*EmbeddedSt4)(nil),
	(*EmbeddedSt5)(nil),
}

//----------------------------------------
// Type definition types

// This will be encoded as
// message SomeName { int64 val = 1; }
type IntDef int

// This will be encoded as
// message SomeName { repeated int val = 1; }
type IntAr [4]int

// This will be encoded as
// message SomeName { repeated int val = 1; }
type IntSl []int

// This will be encoded as
// message SomeName { bytes val = 1; }
type ByteAr [4]byte

// This will be encoded as
// message SomeName { bytes val = 1; }
type ByteSl []byte

type PrimitivesStructDef PrimitivesStruct

// This will be encoded as
// message SomeName { repeated PrimitivesStruct val = 1; }
type PrimitivesStructSl []PrimitivesStruct

// This will be encoded as
// message SomeName { repeated PrimitivesStruct val = 1; }
type PrimitivesStructAr [2]PrimitivesStruct

var DefTypes = []interface{}{
	(*IntDef)(nil),
	(*IntAr)(nil),
	(*IntSl)(nil),
	(*ByteAr)(nil),
	(*ByteSl)(nil),
	(*PrimitivesStructSl)(nil),
	(*PrimitivesStructDef)(nil),
}

//----------------------------------------
// Register/Interface test types

type Interface1 interface {
	AssertInterface1()
}

type Interface2 interface {
	AssertInterface2()
}

type Concrete1 struct{}

func (Concrete1) AssertInterface1() {}
func (Concrete1) AssertInterface2() {}

type Concrete2 struct{}

func (Concrete2) AssertInterface1() {}
func (Concrete2) AssertInterface2() {}

// Special case: this concrete implementation (of Interface1) is a type definition.
type ConcreteTypeDef [4]byte

func (ConcreteTypeDef) AssertInterface1() {}

// Ideally, user's of amino should refrain from using the above
// but wrap actual values in structs; e.g. like:
type ConcreteWrappedBytes struct {
	Value []byte
}

func (ConcreteWrappedBytes) AssertInterface1() {}

// Yet another special case: Field could be a type alias (should not be wrapped).
type InterfaceFieldsStruct struct {
	F1 Interface1
	F2 Interface1
	F3 interface{}
	F4 interface{}
}

func (*InterfaceFieldsStruct) AssertInterface1() {}
