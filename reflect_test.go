package wire

import (
	"math/rand"
	"reflect"
	"runtime/debug"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

// NOTE: See registered fuzz funcs for *byte, **byte, and ***byte.
type NestedPointersStruct struct {
	Ptr1 *byte
	Ptr2 **byte
	Ptr3 ***byte
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
	(*ShortArraysStruct)(nil),
	(*ArraysStruct)(nil),
	(*SlicesStruct)(nil),
	(*PointersStruct)(nil),
	(*NestedPointersStruct)(nil),
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
// Register types

type Interface1 interface {
	AssertInterface1()
}

type Interface2 interface {
	AssertInterface2()
}

type Concrete1 struct{}

func (_ Concrete1) AssertInterface1() {}
func (_ Concrete1) AssertInterface2() {}

type Concrete2 struct{}

func (_ Concrete2) AssertInterface1() {}
func (_ Concrete2) AssertInterface2() {}

type Concrete3 [4]byte

func (_ Concrete3) AssertInterface1() {}

//-------------------------------------
// Non-interface tests

func TestCodecStruct(t *testing.T) {
	for _, ptr := range structTypes {
		rt := getTypeFromPointer(ptr)
		name := rt.Name()
		t.Run(name+":binary", func(t *testing.T) { _testCodec(t, rt, "binary") })
		t.Run(name+":json", func(t *testing.T) { _testCodec(t, rt, "json") })
	}
}

func TestCodecDef(t *testing.T) {
	for _, ptr := range defTypes {
		rt := getTypeFromPointer(ptr)
		name := rt.Name()
		t.Run(name+":binary", func(t *testing.T) { _testCodec(t, rt, "binary") })
		t.Run(name+":json", func(t *testing.T) { _testCodec(t, rt, "json") })
	}
}

func _testCodec(t *testing.T, rt reflect.Type, codecType string) {

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
	f.Funcs(fuzzFuncs...)

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

		switch codecType {
		case "binary":
			bz, err = cdc.MarshalBinary(ptr)
		case "json":
			bz, err = cdc.MarshalJSON(ptr)
		default:
			panic("should not happen")
		}
		require.Nil(t, err,
			"failed to marshal %v to bytes: %v\n",
			spw(ptr), err)

		switch codecType {
		case "binary":
			err = cdc.UnmarshalBinary(bz, ptr2)
		case "json":
			err = cdc.UnmarshalJSON(bz, ptr2)
		default:
			panic("should not happen")
		}
		require.Nil(t, err,
			"failed to unmarshal bytes %X: %v\nptr: %v\n",
			bz, err, spw(ptr))
		require.Equal(t, ptr, ptr2,
			"end to end failed.\nstart: %v\nend: %v\nbytes: %X\nstring(bytes): %s\n",
			spw(ptr), spw(ptr2), bz, bz)
	}

}

//----------------------------------------
// Register tests

func TestCodecBinaryRegister1(t *testing.T) {
	cdc := NewCodec()
	//cdc.RegisterInterface((*Interface1)(nil), nil)
	cdc.RegisterConcrete((*Concrete1)(nil), "Concrete1", nil)

	bz, err := cdc.MarshalBinary(struct{ Interface1 }{Concrete1{}})
	assert.NotNil(t, err, "unregistered interface")
	assert.Empty(t, bz)
}

func TestCodecBinaryRegister2(t *testing.T) {
	cdc := NewCodec()
	cdc.RegisterInterface((*Interface1)(nil), nil)
	cdc.RegisterConcrete((*Concrete1)(nil), "Concrete1", nil)

	bz, err := cdc.MarshalBinary(struct{ Interface1 }{Concrete1{}})
	assert.Nil(t, err, "correctly registered")
	assert.Equal(t, []byte{0xe3, 0xda, 0xb8, 0x33}, bz,
		"prefix bytes did not match")
}

func TestCodecBinaryRegister3(t *testing.T) {
	cdc := NewCodec()
	cdc.RegisterConcrete((*Concrete1)(nil), "Concrete1", nil)
	cdc.RegisterInterface((*Interface1)(nil), nil)

	bz, err := cdc.MarshalBinary(struct{ Interface1 }{Concrete1{}})
	assert.Nil(t, err, "correctly registered")
	assert.Equal(t, []byte{0xe3, 0xda, 0xb8, 0x33}, bz,
		"prefix bytes did not match")
}

func TestCodecBinaryRegister4(t *testing.T) {
	cdc := NewCodec()
	cdc.RegisterConcrete((*Concrete1)(nil), "Concrete1", nil)
	cdc.RegisterInterface((*Interface1)(nil), &InterfaceOptions{
		AlwaysDisambiguate: true,
	})

	bz, err := cdc.MarshalBinary(struct{ Interface1 }{Concrete1{}})
	assert.Nil(t, err, "correctly registered")
	assert.Equal(t, []byte{0x0, 0x12, 0xb5, 0x86, 0xe3, 0xda, 0xb8, 0x33}, bz,
		"prefix bytes did not match")
}

func TestCodecBinaryRegister5(t *testing.T) {
	cdc := NewCodec()
	//cdc.RegisterConcrete((*Concrete1)(nil), "Concrete1", nil)
	cdc.RegisterInterface((*Interface1)(nil), nil)

	bz, err := cdc.MarshalBinary(struct{ Interface1 }{Concrete1{}})
	assert.NotNil(t, err, "concrete type not registered")
	assert.Empty(t, bz)
}

func TestCodecBinaryRegister6(t *testing.T) {
	cdc := NewCodec()
	cdc.RegisterInterface((*Interface1)(nil), nil)
	cdc.RegisterConcrete((*Concrete1)(nil), "Concrete1", nil)

	assert.Panics(t, func() {
		cdc.RegisterConcrete((*Concrete2)(nil), "Concrete1", nil)
	}, "duplicate concrete name")
}

func TestCodecBinaryRegister7(t *testing.T) {
	cdc := NewCodec()
	cdc.RegisterInterface((*Interface1)(nil), nil)
	cdc.RegisterConcrete((*Concrete1)(nil), "Concrete1", nil)
	cdc.RegisterConcrete((*Concrete2)(nil), "Concrete2", nil)

	{ // test Concrete1, no conflict.
		bz, err := cdc.MarshalBinary(struct{ Interface1 }{Concrete1{}})
		assert.Nil(t, err, "correctly registered")
		assert.Equal(t, []byte{0xe3, 0xda, 0xb8, 0x33}, bz,
			"disfix bytes did not match")
	}

	{ // test Concrete2, no conflict
		bz, err := cdc.MarshalBinary(struct{ Interface1 }{Concrete2{}})
		assert.Nil(t, err, "correctly registered")
		assert.Equal(t, []byte{0x6a, 0x9, 0xca, 0x1}, bz,
			"disfix bytes did not match")
	}
}

func TestCodecBinaryRegister8(t *testing.T) {
	cdc := NewCodec()
	cdc.RegisterInterface((*Interface1)(nil), nil)
	cdc.RegisterConcrete(Concrete3{}, "Concrete3", nil)

	assert.Panics(t, func() {
		cdc.RegisterConcrete(Concrete2{}, "Concrete3", nil)
	}, "duplicate concrete name")

	var c3 Concrete3
	copy(c3[:], []byte("0123"))

	bz, err := cdc.MarshalBinary(struct{ Interface1 }{c3})
	assert.Nil(t, err)
	assert.Equal(t, []byte{0x53, 0x37, 0x21, 0x01, 0x30, 0x31, 0x32, 0x33}, bz,
		"Concrete3 incorrectly serialized")

	var i1 Interface1
	err = cdc.UnmarshalBinary(bz, &i1)
	assert.Nil(t, err)
	assert.Equal(t, c3, i1)
}

func TestCodecJSONRegister8(t *testing.T) {
	cdc := NewCodec()
	cdc.RegisterInterface((*Interface1)(nil), nil)
	cdc.RegisterConcrete(Concrete3{}, "Concrete3", nil)

	assert.Panics(t, func() {
		cdc.RegisterConcrete(Concrete2{}, "Concrete3", nil)
	}, "duplicate concrete name")

	var c3 Concrete3
	copy(c3[:], []byte("0123"))

	// NOTE: We don't wrap c3...
	// But that's OK, JSON still writes the disfix bytes by default.
	bz, err := cdc.MarshalJSON(c3)
	assert.Nil(t, err)
	assert.Equal(t, []byte(`{"_df":"43FAF453372101","_v":"MDEyMw=="}`),
		bz, "Concrete3 incorrectly serialized")

	var i1 Interface1
	err = cdc.UnmarshalJSON(bz, &i1)
	assert.Nil(t, err)
	assert.Equal(t, c3, i1)
}

//----------------------------------------
// Misc.

func spw(o interface{}) string {
	return spew.Sprintf("%#v", o)
}

var fuzzFuncs = []interface{}{
	func(bz *[]byte, c fuzz.Continue) {
		// Prefer nil instead of empty, for deep equality.
		// (go-wire decoder will always prefer nil).
		c.Fuzz(bz)
		if len(*bz) == 0 {
			*bz = nil
		}
	},
	func(bz **[]byte, c fuzz.Continue) {
		// Prefer nil instead of empty, for deep equality.
		// (go-wire decoder will always prefer nil).
		c.Fuzz(bz)
		if *bz == nil {
			return
		}
		if len(**bz) == 0 {
			*bz = nil
		}
		return
	},
	func(tyme *time.Time, c fuzz.Continue) {
		// Set time.Unix(_,_) to wipe .wal
		switch c.Intn(4) {
		case 0:
			ns := c.Int63n(10)
			*tyme = time.Unix(0, ns)
		case 1:
			ns := c.Int63n(1e10)
			*tyme = time.Unix(0, ns)
		case 2:
			const maxSeconds = 4611686018 // (1<<63 - 1) / 1e9
			s := c.Int63n(maxSeconds)
			ns := c.Int63n(1e10)
			*tyme = time.Unix(s, ns)
		case 3:
			s := c.Int63n(10)
			ns := c.Int63n(1e10)
			*tyme = time.Unix(s, ns)
		}
		// Strip timezone and monotonic for deep equality.
		*tyme = tyme.UTC().Truncate(time.Millisecond)
	},

	// For testing nested pointers...
	func(ptr **byte, c fuzz.Continue) {
		if c.Intn(5) == 0 {
			*ptr = nil
			return
		}
		*ptr = new(byte)
	},
	func(ptr ***byte, c fuzz.Continue) {
		if c.Intn(5) == 0 {
			*ptr = nil
			return
		}
		*ptr = new(*byte)
		**ptr = new(byte)
	},
	func(ptr ****byte, c fuzz.Continue) {
		if c.Intn(5) == 0 {
			*ptr = nil
			return
		}
		*ptr = new(**byte)
		**ptr = new(*byte)
		***ptr = new(byte)
	},
}
