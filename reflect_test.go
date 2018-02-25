package wire

import (
	"math/rand"
	"reflect"
	"runtime/debug"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	fuzz "github.com/google/gofuzz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/go-wire/tests"
)

//-------------------------------------
// Non-interface Google fuzz tests

func TestCodecStruct(t *testing.T) {
	for _, ptr := range tests.StructTypes {
		rt := getTypeFromPointer(ptr)
		name := rt.Name()
		t.Run(name+":binary", func(t *testing.T) { _testCodec(t, rt, "binary") })
		t.Run(name+":json", func(t *testing.T) { _testCodec(t, rt, "json") })
	}
}

func TestCodecDef(t *testing.T) {
	for _, ptr := range tests.DefTypes {
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
	//cdc.RegisterInterface((*tests.Interface1)(nil), nil)
	cdc.RegisterConcrete((*tests.Concrete1)(nil), "Concrete1", nil)

	bz, err := cdc.MarshalBinary(struct{ tests.Interface1 }{tests.Concrete1{}})
	assert.NotNil(t, err, "unregistered interface")
	assert.Empty(t, bz)
}

func TestCodecBinaryRegister2(t *testing.T) {
	cdc := NewCodec()
	cdc.RegisterInterface((*tests.Interface1)(nil), nil)
	cdc.RegisterConcrete((*tests.Concrete1)(nil), "Concrete1", nil)

	bz, err := cdc.MarshalBinary(struct{ tests.Interface1 }{tests.Concrete1{}})
	assert.Nil(t, err, "correctly registered")
	assert.Equal(t, []byte{0xe3, 0xda, 0xb8, 0x33}, bz,
		"prefix bytes did not match")
}

func TestCodecBinaryRegister3(t *testing.T) {
	cdc := NewCodec()
	cdc.RegisterConcrete((*tests.Concrete1)(nil), "Concrete1", nil)
	cdc.RegisterInterface((*tests.Interface1)(nil), nil)

	bz, err := cdc.MarshalBinary(struct{ tests.Interface1 }{tests.Concrete1{}})
	assert.Nil(t, err, "correctly registered")
	assert.Equal(t, []byte{0xe3, 0xda, 0xb8, 0x33}, bz,
		"prefix bytes did not match")
}

func TestCodecBinaryRegister4(t *testing.T) {
	cdc := NewCodec()
	cdc.RegisterConcrete((*tests.Concrete1)(nil), "Concrete1", nil)
	cdc.RegisterInterface((*tests.Interface1)(nil), &InterfaceOptions{
		AlwaysDisambiguate: true,
	})

	bz, err := cdc.MarshalBinary(struct{ tests.Interface1 }{tests.Concrete1{}})
	assert.Nil(t, err, "correctly registered")
	assert.Equal(t, []byte{0x0, 0x12, 0xb5, 0x86, 0xe3, 0xda, 0xb8, 0x33}, bz,
		"prefix bytes did not match")
}

func TestCodecBinaryRegister5(t *testing.T) {
	cdc := NewCodec()
	//cdc.RegisterConcrete((*tests.Concrete1)(nil), "Concrete1", nil)
	cdc.RegisterInterface((*tests.Interface1)(nil), nil)

	bz, err := cdc.MarshalBinary(struct{ tests.Interface1 }{tests.Concrete1{}})
	assert.NotNil(t, err, "concrete type not registered")
	assert.Empty(t, bz)
}

func TestCodecBinaryRegister6(t *testing.T) {
	cdc := NewCodec()
	cdc.RegisterInterface((*tests.Interface1)(nil), nil)
	cdc.RegisterConcrete((*tests.Concrete1)(nil), "Concrete1", nil)

	assert.Panics(t, func() {
		cdc.RegisterConcrete((*tests.Concrete2)(nil), "Concrete1", nil)
	}, "duplicate concrete name")
}

func TestCodecBinaryRegister7(t *testing.T) {
	cdc := NewCodec()
	cdc.RegisterInterface((*tests.Interface1)(nil), nil)
	cdc.RegisterConcrete((*tests.Concrete1)(nil), "Concrete1", nil)
	cdc.RegisterConcrete((*tests.Concrete2)(nil), "Concrete2", nil)

	{ // test tests.Concrete1, no conflict.
		bz, err := cdc.MarshalBinary(struct{ tests.Interface1 }{tests.Concrete1{}})
		assert.Nil(t, err, "correctly registered")
		assert.Equal(t, []byte{0xe3, 0xda, 0xb8, 0x33}, bz,
			"disfix bytes did not match")
	}

	{ // test tests.Concrete2, no conflict
		bz, err := cdc.MarshalBinary(struct{ tests.Interface1 }{tests.Concrete2{}})
		assert.Nil(t, err, "correctly registered")
		assert.Equal(t, []byte{0x6a, 0x9, 0xca, 0x1}, bz,
			"disfix bytes did not match")
	}
}

func TestCodecBinaryRegister8(t *testing.T) {
	cdc := NewCodec()
	cdc.RegisterInterface((*tests.Interface1)(nil), nil)
	cdc.RegisterConcrete(tests.Concrete3{}, "Concrete3", nil)

	assert.Panics(t, func() {
		cdc.RegisterConcrete(tests.Concrete2{}, "Concrete3", nil)
	}, "duplicate concrete name")

	var c3 tests.Concrete3
	copy(c3[:], []byte("0123"))

	bz, err := cdc.MarshalBinary(struct{ tests.Interface1 }{c3})
	assert.Nil(t, err)
	assert.Equal(t, []byte{0x53, 0x37, 0x21, 0x01, 0x30, 0x31, 0x32, 0x33}, bz,
		"Concrete3 incorrectly serialized")

	var i1 tests.Interface1
	err = cdc.UnmarshalBinary(bz, &i1)
	assert.Nil(t, err)
	assert.Equal(t, c3, i1)
}

func TestCodecJSONRegister8(t *testing.T) {
	cdc := NewCodec()
	cdc.RegisterInterface((*tests.Interface1)(nil), nil)
	cdc.RegisterConcrete(tests.Concrete3{}, "Concrete3", nil)

	assert.Panics(t, func() {
		cdc.RegisterConcrete(tests.Concrete2{}, "Concrete3", nil)
	}, "duplicate concrete name")

	var c3 tests.Concrete3
	copy(c3[:], []byte("0123"))

	// NOTE: We don't wrap c3...
	// But that's OK, JSON still writes the disfix bytes by default.
	bz, err := cdc.MarshalJSON(c3)
	assert.Nil(t, err)
	assert.Equal(t, []byte(`{"_df":"43FAF453372101","_v":"MDEyMw=="}`),
		bz, "Concrete3 incorrectly serialized")

	var i1 tests.Interface1
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
