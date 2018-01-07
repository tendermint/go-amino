package wire

import (
	"math/rand"
	"reflect"
	"runtime/debug"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"

	fuzz "github.com/google/gofuzz"
)

//----------------------------------------
// Struct types

type PrimitivesStruct struct {
	Int8    int8
	Int16   int16
	Int32   int32
	Int64   int64
	Varint  int64 `binary:"varint"`
	Int     int
	Byte    byte
	Uint8   uint8
	Uint16  uint16
	Uint32  uint32
	Uint64  uint64
	Uvarint uint64 `binary:"varint"`
	Uint    uint
	String  string
	Bytes   []byte
	Time    time.Time
}

type ShortArraysStruct struct {
	TimeAr [0]time.Time
}

type ArraysStruct struct {
	Int8Ar    [4]int8
	Int16Ar   [4]int16
	Int32Ar   [4]int32
	Int64Ar   [4]int64
	VarintAr  [4]int64 `binary:"varint"`
	IntAr     [4]int
	ByteAr    [4]byte
	Uint8Ar   [4]uint8
	Uint16Ar  [4]uint16
	Uint32Ar  [4]uint32
	Uint64Ar  [4]uint64
	UvarintAr [4]uint64 `binary:"varint"`
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
	VarintSl  []int64 `binary:"varint"`
	IntSl     []int
	ByteSl    []byte
	Uint8Sl   []uint8
	Uint16Sl  []uint16
	Uint32Sl  []uint32
	Uint64Sl  []uint64
	UvarintSl []uint64 `binary:"varint"`
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
	VarintPt  *int64 `binary:"varint"`
	IntPt     *int
	BytePt    *byte
	Uint8Pt   *uint8
	Uint16Pt  *uint16
	Uint32Pt  *uint32
	Uint64Pt  *uint64
	UvarintPt *uint64 `binary:"varint"`
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
	//(*PrimitivesStruct)(nil),
	(*ShortArraysStruct)(nil),
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
	for _, ptr := range structTypes {
		rt := getTypeFromPointer(ptr)
		name := rt.Name()
		t.Run(name, func(t *testing.T) { _testCodecBinary(t, rt) })
	}
}

func TestCodecBinaryDef(t *testing.T) {
	for _, ptr := range defTypes {
		rt := getTypeFromPointer(ptr)
		name := rt.Name()
		t.Run(name, func(t *testing.T) { _testCodecBinary(t, rt) })
	}
}

func _testCodecBinary(t *testing.T, rt reflect.Type) {

	err := error(nil)
	bz := []byte{}
	cdc := NewCodec()
	f := fuzz.New()
	rv := reflect.New(rt)
	rv2 := reflect.New(rt)
	ptr := rv.Interface()
	ptr2 := rv2.Interface()
	rnd := rand.New(rand.NewSource(10))
	f.RandSource(rnd)
	f.Funcs(
		func(bz *[]byte, c fuzz.Continue) {
			// Prefer nil instead of empty, for deep equality.
			// (go-wire decoder will always prefer nil).
			c.Fuzz(bz)
			if len(*bz) == 0 {
				*bz = nil
			}
		},
		func(tyme *time.Time, c fuzz.Continue) {
			// Set time.Unix(_,_) to wipe .wal
			switch 0 { // c.Intn(4) {
			case 0:
				ns := c.Int63n(10)
				*tyme = time.Unix(0, ns)
			case 1:
				ns := c.Int63n(10)
				*tyme = time.Unix(0, ns)
				break
				/*
					ns := c.Int63n(1e10)
					*tyme = time.Unix(0, ns)
				*/
			case 2:
				ns := c.Int63n(10)
				*tyme = time.Unix(0, ns)
				break
				/*
					const maxSeconds = 4611686018 // (1<<63 - 1) / 1e9
					s := c.Int63n(maxSeconds)
					ns := c.Int63n(1e10)
					*tyme = time.Unix(s, ns)
				*/
			case 3:
				ns := c.Int63n(10)
				*tyme = time.Unix(0, ns)
				break
				/*
					s := c.Int63n(10)
					ns := c.Int63n(1e10)
					*tyme = time.Unix(s, ns)
				*/
			}
			// Strip timezone and monotonic for deep equality.
			*tyme = tyme.UTC().Truncate(time.Millisecond)
		},
	)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic'd:\nreason: %v\n%s\nerr: %v\nbz: %X\nrv: %#v\nrv2: %#v\nptr: %v\nptr2: %v\n",
				r, debug.Stack(), err, bz, rv, rv2, spw(ptr), spw(ptr2),
			)
		}
	}()

	for i := 0; i < 1e4; i++ {
		f.Fuzz(ptr)

		// Reset, which makes debugging decoding easier.
		rv2 = reflect.New(rt)
		ptr2 = rv2.Interface()

		bz, err = cdc.MarshalBinary(ptr)
		if err != nil {
			t.Fatalf("failed to marshal %v to bytes: %v\n", spw(ptr), err)
		}

		err = cdc.UnmarshalBinary(bz, ptr2)
		if err != nil {
			t.Fatalf("failed to unmarshal bytes %X: %v\nptr: %v\n", bz, err, spw(ptr))
		}

		if !reflect.DeepEqual(ptr, ptr2) {
			t.Fatalf("end to end failed.\nstart: %v\nend: %v\nbytes: %X\n",
				spw(ptr), spw(ptr2), bz)
		}
	}
}

func TestCodecBinaryInterface(t *testing.T) {
	// XXX
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

func spw(o interface{}) string {
	return spew.Sprintf("%#v", o)
}
