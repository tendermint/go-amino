package wire

import (
	"testing"
	"time"
)

//----------------------------------------
// Struct types

type PrimitivesStruct struct {
	Int8    int8
	Int16   int16
	Int32   int32
	Int64   int64
	Varint  int64 `wire:binvarint`
	Int     int
	Byte    byte
	Uint8   uint8
	Uint16  uint16
	Uint32  uint32
	Uint64  uint64
	Uvarint uint64 `wire:binvaruint`
	Uint    uint
	String  string
	Bytes   []byte
	Time    time.Time
}

type ArraysStruct struct {
	Int8Ar    [4]int8
	Int16Ar   [4]int16
	Int32Ar   [4]int32
	Int64Ar   [4]int64
	VarintAr  [4]int64 `wire:binvarint`
	IntAr     [4]int
	ByteAr    [4]byte
	Uint8Ar   [4]uint8
	Uint16Ar  [4]uint16
	Uint32Ar  [4]uint32
	Uint64Ar  [4]uint64
	UvarintAr [4]uint64 `wire:binvaruint`
	UintAr    [4]int
	StringAr  [4]string
	BytesAr   [4][]byte
	TimeAr    [4]time.Time
}

type SlicesStruct struct {
	Int8Sl    []int8
	Int16Sl   []int16
	Int32Sl   []int32
	Int64Sl   []int64
	VarintSl  []int64 `wire:binvarint`
	IntSl     []int
	ByteSl    []byte
	Uint8Sl   []uint8
	Uint16Sl  []uint16
	Uint32Sl  []uint32
	Uint64Sl  []uint64
	UvarintSl []uint64 `wire:binvaruint`
	UintSl    []int
	StringSl  []string
	BytesSl   [][]byte
	TimeSl    []time.Time
}

type PointersStruct struct {
	Int8Pt    *int8
	Int16Pt   *int16
	Int32Pt   *int32
	Int64Pt   *int64
	VarintPt  *int64 `wire:binvarint`
	IntPt     *int
	BytePt    *byte
	Uint8Pt   *uint8
	Uint16Pt  *uint16
	Uint32Pt  *uint32
	Uint64Pt  *uint64
	UvarintPt *uint64 `wire:binvaruint`
	UintPt    *int
	StringPt  *string
	BytesPt   *[]byte
	TimePt    *time.Time
}

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

var structTypes = []interface{}{
	(*PrimitivesStruct)(nil),
	(*ArraysStruct)(nil),
	(*SlicesStruct)(nil),
	(*PointersStruct)(nil),
	(*ComplexSt)(nil),
	(*EmbeddedSt1)(nil),
	(*EmbeddedSt2)(nil),
	(*EmbeddedSt3)(nil),
	(*EmbeddedSt4)(nil),
	(*EmbeddedSt5)(nil),
}

//----------------------------------------
// Type definition types

type IntDef int

type IntAr [4]int

type IntSl []int

type ByteAr [4]byte

type ByteSl []byte

type PrimitivesStructSl []PrimitivesStruct

type PrimitivesStructDef PrimitivesStruct

var defTypes = []interface{}{
	(*IntDef)(nil),
	(*IntAr)(nil),
	(*IntSl)(nil),
	(*ByteAr)(nil),
	(*ByteSl)(nil),
	(*PrimitivesStructSl)(nil),
	(*PrimitivesStructDef)(nil),
}

//----------------------------------------
// Interface+Concrete types

type SimpleInterface interface {
	AssureSimpleInterface()
}

type SimpleConcrete1 struct {
	// empty
}

func (_ SimpleConcrete1) AssureSimpleInterface() {}

type SimpleConcrete2 struct {
	PrimitivesStruct
}

func (_ SimpleConcrete2) AssureSimpleInterface() {}

type SimpleConcrete3 struct {
	PrimitivesStruct
}

func (_ *SimpleConcrete3) AssureSimpleInterface() {}

type SimpleConcrete4 struct {
	*PrimitivesStruct
}

func (_ *SimpleConcrete4) AssureSimpleInterface() {}

type SimpleConcrete5 struct {
	// empty
}

func (_ *SimpleConcrete5) AssureSimpleInterface() {}

//-------------------------------------

func TestCodecBinaryStruct(t *testing.T) {
	for _, stPtr := range structTypes {
		stRv := getTypeFromPointer(stPtr)
		stName := stRv.Name()
		t.Run(stName, _testCodecBinary)
	}
}

func TestCodecBinaryDef(t *testing.T) {
	for _, stPtr := range defTypes {
		stRv := getTypeFromPointer(stPtr)
		stName := stRv.Name()
		t.Run(stName, _testCodecBinary)
	}
}

func _testCodecBinary(t *testing.T) {
	t.Log("woot")
}

func TestCodecBinaryInterface(t *testing.T) {
	t.Log("woot2")
}

/*
func TestBinary(t *testing.T) {

	for i, testCase := range testCases {

		t.Log(fmt.Sprintf("Running test case %v", i))

		// Construct an object
		o := testCase.Constructor()

		// Write the object
		data, err := wire.MarshalBinary(o)
		assert.Nil(t, err)
		t.Logf("Binary: %X", data)

		instance, instancePtr := testCase.Instantiator()

		// Read onto a struct
		err = wire.UnmarshalBinary(data, instance)
		if err != nil {
			t.Fatalf("Failed to read into instance: %v", err)
		}

		// Validate object
		testCase.Validator(instance, t)

		// Read onto a pointer
		err = wire.UnmarshalBinary(data, instancePtr)
		if err != nil {
			t.Fatalf("Failed to read into instance: %v", err)
		}
		if instance != instancePtr {
			t.Errorf("Expected pointer to pass through")
		}

		// Validate object
		testCase.Validator(reflect.ValueOf(instance).Elem().Interface(), t)

		// Read with len(data)-1 limit should fail.
		// TODO
		/*
			instance, _ = testCase.Instantiator()
			err = wire.UnmarshalBinary(data, instance)
			if err != wire.ErrBinaryReadOverflow {
				t.Fatalf("Expected ErrBinaryReadOverflow")
			}

			// Read with len(data) limit should succeed.
			instance, _ = testCase.Instantiator()
			err = wire.UnmarshalBinary(data, instance)
			if err != nil {
				t.Fatalf("Failed to read instance with sufficient limit: %v n: %v len(data): %v type: %v",
					(err).Error(), _, len(data), reflect.TypeOf(instance))
			}
		* /
	}

}
*/
