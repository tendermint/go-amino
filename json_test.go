package wire_test

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/go-wire"
)

var cdc = wire.NewCodec()

func TestMain(m *testing.M) {
	// Register the concrete types first.
	cdc.RegisterConcrete(&Transport{}, "our/transport", nil)
	cdc.RegisterInterface((*Vehicle)(nil), &wire.InterfaceOptions{AlwaysDisambiguate: true})
	cdc.RegisterInterface((*Asset)(nil), &wire.InterfaceOptions{AlwaysDisambiguate: true})
	cdc.RegisterConcrete(Car(""), "car", nil)
	cdc.RegisterConcrete(insurancePlan(0), "insuranceplan", nil)
	cdc.RegisterConcrete(Boat(""), "boat", nil)
	cdc.RegisterConcrete(Plane{}, "plane", nil)

	os.Exit(m.Run())
}

func TestMarshalJSON(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in      interface{}
		want    string
		wantErr string
	}{
		{&noFields{}, "{}", ""},
		{&noExportedFields{a: 10, b: "foo"}, "{}", ""},
		{nil, "null", ""},
		{&oneExportedField{}, `{"A":""}`, ""},
		{Vehicle(Car("Tesla")), `{"_df":"2B2961A431B23C","_v":"Tesla"}`, ""},
		{Car("Tesla"), `{"_df":"2B2961A431B23C","_v":"Tesla"}`, ""},
		{&oneExportedField{A: "Z"}, `{"A":"Z"}`, ""},
		{[]string{"a", "bc"}, `["a","bc"]`, ""},
		{[]interface{}{"a", "bc", 10, 10.93, 1e3}, ``, "Unregistered"},
		{aPointerField{Foo: new(int), Name: "name"}, `{"Foo":0,"nm":"name"}`, ""},
		{
			aPointerFieldAndEmbeddedField{intPtr(11), "ap", nil, &oneExportedField{A: "foo"}},
			`{"Foo":11,"nm":"ap","bz":{"A":"foo"}}`, "",
		},

		{
			doublyEmbedded{
				Inner: &aPointerFieldAndEmbeddedField{
					intPtr(11), "ap", nil, &oneExportedField{A: "foo"},
				},
			},
			`{"Inner":{"Foo":11,"nm":"ap","bz":{"A":"foo"}},"year":0}`, "",
		},

		{
			struct{}{}, `{}`, "",
		},
		{
			struct{ A int }{A: 10}, `{"A":10}`, "",
		},
		{
			Transport{},
			`{"_df":"AEB127E121A6B2","_v":{"Vehicle":null,"Capacity":0}}`, "",
		},
		{
			Transport{Vehicle: Car("Bugatti")},
			`{"_df":"AEB127E121A6B2","_v":{"Vehicle":{"_df":"2B2961A431B23C","_v":"Bugatti"},"Capacity":0}}`, "",
		},
		{
			BalanceSheet{Assets: []Asset{Car("Corolla"), insurancePlan(1e7)}},
			`{"assets":[{"_df":"2B2961A431B23C","_v":"Corolla"},{"_df":"7DF0BC76182A1F","_v":10000000}]}`, "",
		},
		{
			Transport{Vehicle: Boat("Poseidon"), Capacity: 1789},
			`{"_df":"AEB127E121A6B2","_v":{"Vehicle":{"_df":"25CDB46D8D2115","_v":"Poseidon"},"Capacity":1789}}`, "",
		},
		{
			withCustomMarshaler{A: &aPointerField{Foo: intPtr(12)}, F: customJSONMarshaler(10)},
			`{"fx":"Tendermint","A":{"Foo":12}}`, "",
		},
		{
			func() json.Marshaler { v := customJSONMarshaler(10); return &v }(),
			`"Tendermint"`, "",
		},

		// We don't yet support interface pointer registration i.e. `*interface{}`
		{interfacePtr("a"), "", "Unregistered interface interface {}"},

		{&fp{"Foo", 10}, "<FP-MARSHALJSON>", ""},
		{(*fp)(nil), "null", ""},
		{struct {
			FP      *fp
			Package string
		}{FP: &fp{"Foo", 10}, Package: "bytes"},
			`{"FP":<FP-MARSHALJSON>,"Package":"bytes"}`, ""},
	}

	for i, tt := range cases {
		t.Logf("Trying case #%v", i)
		blob, err := cdc.MarshalJSON(tt.in)
		if tt.wantErr != "" {
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("#%d:\ngot:\n\t%q\nwant non-nil error containing\n\t%q", i,
					err, tt.wantErr)
			}
			continue
		}

		if err != nil {
			t.Errorf("#%d: unexpected error: %v\nblob: %v", i, err, tt.in)
			continue
		}
		if g, w := string(blob), tt.want; g != w {
			t.Errorf("#%d:\ngot:\n\t%s\nwant:\n\t%s", i, g, w)
		}
	}
}

func TestMarshalJSONWithMonotonicTime(t *testing.T) {
	var cdc = wire.NewCodec()

	type SimpleStruct struct {
		String string
		Bytes  []byte
		Time   time.Time
	}

	s := SimpleStruct{
		String: "hello",
		Bytes:  []byte("goodbye"),
		Time:   time.Now().UTC().Truncate(time.Millisecond), // strip monotonic and timezone.
	}

	b, err := cdc.MarshalJSON(s)
	assert.Nil(t, err)

	var s2 SimpleStruct
	err = cdc.UnmarshalJSON(b, &s2)
	assert.Nil(t, err)
	assert.Equal(t, s, s2)
}

type fp struct {
	Name    string
	Version int
}

func (f *fp) MarshalJSON() ([]byte, error) {
	return []byte("<FP-MARSHALJSON>"), nil
}

func (f *fp) UnmarshalJSON(blob []byte) error {
	f.Name = string(blob)
	return nil
}

var _ json.Marshaler = (*fp)(nil)
var _ json.Unmarshaler = (*fp)(nil)

type innerFP struct {
	PC uint64
	FP *fp
}

func TestUnmarshalMap(t *testing.T) {
	binBytes := []byte(`dontcare`)
	jsonBytes := []byte(`{"2": 2}`)
	obj := new(map[string]int)
	cdc := wire.NewCodec()
	// Binary doesn't support decoding to a map...
	assert.Panics(t, func() {
		err := cdc.UnmarshalBinary(binBytes, &obj)
		assert.Fail(t, "should have paniced but got err: %v", err)
	})
	assert.Panics(t, func() {
		err := cdc.UnmarshalBinary(binBytes, obj)
		assert.Fail(t, "should have paniced but got err: %v", err)
	})
	// ... nor encoding it.
	assert.Panics(t, func() {
		bz, err := cdc.MarshalBinary(obj)
		assert.Fail(t, "should have paniced but got bz: %X err: %v", bz, err)
	})
	// JSON doesn't support decoding to a map...
	assert.Panics(t, func() {
		err := cdc.UnmarshalJSON(jsonBytes, &obj)
		assert.Fail(t, "should have paniced but got err: %v", err)
	})
	assert.Panics(t, func() {
		err := cdc.UnmarshalJSON(jsonBytes, obj)
		assert.Fail(t, "should have paniced but got err: %v", err)
	})
	// ... nor encoding it.
	assert.Panics(t, func() {
		bz, err := cdc.MarshalJSON(obj)
		assert.Fail(t, "should have paniced but got bz: %X err: %v", bz, err)
	})
}

func TestUnmarshalFunc(t *testing.T) {
	binBytes := []byte(`dontcare`)
	jsonBytes := []byte(`"dontcare"`)
	obj := func() {}
	cdc := wire.NewCodec()
	// Binary doesn't support decoding to a func...
	assert.Panics(t, func() {
		err := cdc.UnmarshalBinary(binBytes, &obj)
		assert.Fail(t, "should have paniced but got err: %v", err)
	})
	assert.Panics(t, func() {
		err := cdc.UnmarshalBinary(binBytes, obj)
		assert.Fail(t, "should have paniced but got err: %v", err)
	})
	// ... nor encoding it.
	assert.Panics(t, func() {
		bz, err := cdc.MarshalBinary(obj)
		assert.Fail(t, "should have paniced but got bz: %X err: %v", bz, err)
	})
	// JSON doesn't support decoding to a func...
	assert.Panics(t, func() {
		err := cdc.UnmarshalJSON(jsonBytes, &obj)
		assert.Fail(t, "should have paniced but got err: %v", err)
	})
	assert.Panics(t, func() {
		err := cdc.UnmarshalJSON(jsonBytes, obj)
		assert.Fail(t, "should have paniced but got err: %v", err)
	})
	// ... nor encoding it.
	assert.Panics(t, func() {
		bz, err := cdc.MarshalJSON(obj)
		assert.Fail(t, "should have paniced but got bz: %X err: %v", bz, err)
	})
}

func TestUnmarshalJSON(t *testing.T) {
	t.Parallel()
	cases := []struct {
		blob    string
		in      interface{}
		want    interface{}
		wantErr string
	}{
		{
			"null", 2, nil, "expects a pointer",
		},
		{
			"null", new(int), new(int), "",
		},
		{
			"2", new(int), intPtr(2), "",
		},
		{
			`{"null"}`, new(int), nil, "invalid character",
		},
		{
			`{"_df":"AEB127E121A6B2","_v":{"Vehicle":null,"Capacity":0}}`, new(Transport), new(Transport), "",
		},
		{
			`{"_df":"AEB127E121A6B2","_v":{"Vehicle":{"_df":"2B2961A431B23C","_v":"Bugatti"},"Capacity":10}}`,
			new(Transport),
			&Transport{
				Vehicle:  Car("Bugatti"),
				Capacity: 10,
			}, "",
		},
		{
			`{"_df":"2B2961A431B23C","_v":"Bugatti"}`, new(Car), func() *Car { c := Car("Bugatti"); return &c }(), "",
		},
		{
			`[1, 2, 3]`, new([]int), func() interface{} {
				v := []int{1, 2, 3}
				return &v
			}(), "",
		},
		{
			`["1", "2", "3"]`, new([]string), func() interface{} {
				v := []string{"1", "2", "3"}
				return &v
			}(), "",
		},
		{
			`[1, "2", ["foo", "bar"]]`, new([]interface{}), nil, "Unregistered",
		},
		{
			`2.34`, floatPtr(2.34), nil, "float* support requires",
		},

		{"<FooBar>", new(fp), &fp{"<FooBar>", 0}, ""},
		{"10", new(fp), &fp{Name: "10"}, ""},
		{`{"PC":125,"FP":"10"}`, new(innerFP), &innerFP{PC: 125, FP: &fp{Name: `"10"`}}, ""},
		{`{"PC":125,"FP":"<FP-FOO>"}`, new(innerFP), &innerFP{PC: 125, FP: &fp{Name: `"<FP-FOO>"`}}, ""},
	}

	for i, tt := range cases {
		err := cdc.UnmarshalJSON([]byte(tt.blob), tt.in)
		if tt.wantErr != "" {
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("#%d:\ngot:\n\t%q\nwant non-nil error containing\n\t%q", i,
					err, tt.wantErr)
			}
			continue
		}

		if err != nil {
			t.Errorf("#%d: unexpected error: %v\nblob: %s\nin: %+v\n", i, err, tt.blob, tt.in)
			continue
		}
		if g, w := tt.in, tt.want; !reflect.DeepEqual(g, w) {
			gb, _ := json.MarshalIndent(g, "", "  ")
			wb, _ := json.MarshalIndent(w, "", "  ")
			t.Errorf("#%d:\ngot:\n\t%#v\n(%s)\n\nwant:\n\t%#v\n(%s)", i, g, gb, w, wb)
		}
	}
}

func TestJSONCodecRoundTrip(t *testing.T) {
	type allInclusive struct {
		Tr      Transport `json:"trx"`
		Vehicle Vehicle   `json:"v,omitempty"`
		Comment string
		Data    []byte
	}

	cases := []struct {
		in      interface{}
		want    interface{}
		out     interface{}
		wantErr string
	}{
		0: {
			in: &allInclusive{
				Tr: Transport{
					Vehicle: Boat("Oracle"),
				},
				Comment: "To the Cosmos! баллинг в космос",
				Data:    []byte("祝你好运"),
			},
			out: new(allInclusive),
			want: &allInclusive{
				Tr: Transport{
					Vehicle: Boat("Oracle"),
				},
				Comment: "To the Cosmos! баллинг в космос",
				Data:    []byte("祝你好运"),
			},
		},

		1: {
			in:   Transport{Vehicle: Plane{Name: "G6", MaxAltitude: 51e3}, Capacity: 18},
			out:  new(Transport),
			want: &Transport{Vehicle: Plane{Name: "G6", MaxAltitude: 51e3}, Capacity: 18},
		},
	}

	for i, tt := range cases {
		mBlob, err := cdc.MarshalJSON(tt.in)
		if tt.wantErr != "" {
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("#%d:\ngot:\n\t%q\nwant non-nil error containing\n\t%q", i,
					err, tt.wantErr)
			}
			continue
		}

		if err != nil {
			t.Errorf("#%d: unexpected error after MarshalJSON: %v", i, err)
			continue
		}

		if err := cdc.UnmarshalJSON(mBlob, tt.out); err != nil {
			t.Errorf("#%d: unexpected error after UnmarshalJSON: %v\nmBlob: %s", i, err, mBlob)
			continue
		}

		// Now check that the input is exactly equal to the output
		uBlob, err := cdc.MarshalJSON(tt.out)
		if err := cdc.UnmarshalJSON(mBlob, tt.out); err != nil {
			t.Errorf("#%d: unexpected error after second MarshalJSON: %v", i, err)
			continue
		}
		if !reflect.DeepEqual(tt.want, tt.out) {
			t.Errorf("#%d: After roundtrip UnmarshalJSON\ngot: \t%v\nwant:\t%v", i, tt.out, tt.want)
		}
		if !bytes.Equal(mBlob, uBlob) {
			t.Errorf("#%d: After roundtrip MarshalJSON\ngot: \t%s\nwant:\t%s", i, uBlob, mBlob)
		}
	}
}

func intPtr(i int) *int {
	return &i
}

func floatPtr(f float64) *float64 {
	return &f
}

type noFields struct{}
type noExportedFields struct {
	a int
	b string
}

type oneExportedField struct {
	_Foo int
	A    string
	b    string
}

type aPointerField struct {
	Foo  *int
	Name string `json:"nm,omitempty"`
}

type doublyEmbedded struct {
	Inner *aPointerFieldAndEmbeddedField
	Year  int64 `json:"year"`
}

type aPointerFieldAndEmbeddedField struct {
	Foo  *int
	Name string `json:"nm,omitempty"`
	*oneExportedField
	B *oneExportedField `json:"bz,omitempty"`
}

type customJSONMarshaler int

var _ json.Marshaler = (*customJSONMarshaler)(nil)

func (cm customJSONMarshaler) MarshalJSON() ([]byte, error) {
	return []byte(`"Tendermint"`), nil
}

type withCustomMarshaler struct {
	F customJSONMarshaler `json:"fx"`
	A *aPointerField
}

type Transport struct {
	Vehicle
	Capacity int
}

type Vehicle interface {
	Move() error
}

type Asset interface {
	Value() float64
}

func (c Car) Value() float64 {
	return 60000.0
}

type BalanceSheet struct {
	Assets []Asset `json:"assets"`
}

type Car string
type Boat string
type Plane struct {
	Name        string
	MaxAltitude int64
}
type insurancePlan int

func (ip insurancePlan) Value() float64 { return float64(ip) }

func (c Car) Move() error   { return nil }
func (b Boat) Move() error  { return nil }
func (p Plane) Move() error { return nil }

func interfacePtr(v interface{}) *interface{} {
	return &v
}
