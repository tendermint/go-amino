package wire_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/tendermint/go-wire"
)

func TestMarshalJSON(t *testing.T) {
	// Register the concrete types first.
	// wire.RegisterConcrete(&Transport{}, "our/transport", nil)
	wire.RegisterInterface((*Vehicle)(nil), &wire.InterfaceOptions{AlwaysDisambiguate: true})
	wire.RegisterConcrete(Car(""), "car", nil)   // &wire.ConcreteOptions{Disamb: []byte{0xC, 0xA}})
	wire.RegisterConcrete(Boat(""), "boat", nil) // &wire.ConcreteOptions{Disamb: []byte{0xB, 0x0, 0xA}})

	cases := []struct {
		in      interface{}
		want    string
		wantErr string
	}{
		{&noFields{}, "{}", ""},
		{&noExportedFields{a: 10, b: "foo"}, "{}", ""},
		{nil, "null", ""},
		{&oneExportedField{}, `{"A":""}`, ""},
		{&oneExportedField{A: "Z"}, `{"A":"Z"}`, ""},
		{[]string{"a", "bc"}, `["a","bc"]`, ""},
		{[]interface{}{"a", "bc", 10, 10.93, 1e3}, `["a","bc",10,10.93,1000]`, ""},
		{aPointerField{Foo: new(int), Name: "name"}, `{"Foo":0,"nm":"name"}`, ""},
		{
			aPointerFieldAndEmbeddedField{intPtr(11), "ap", nil, &oneExportedField{A: "foo"}},
			`{"Foo":11,"nm":"ap","bz":{"A":"foo"}}`, "",
		},
		{
			Transport{Vehicle: Car("Bugatti")},
			// TODO: Modify me when we've figured out disambiguation for JSON
			`{"Vehicle":{"_df":"00000000000001","_data":"Bugatti"},"Capacity":0}`, "",
		},
		{
			Transport{Vehicle: Boat("Poseidon"), Capacity: 1789},
			// TODO: Modify me when we've figured out disambiguation for JSON
			`{"Vehicle":{"_df":"00000000000002","_data":"Poseidon"},"Capacity":1789}`, "",
		},
		{
			withCustomMarshaler{A: &aPointerField{Foo: intPtr(12)}, F: customJSONMarshaler(10)},
			`{"fx":"Tendermint","A":{"Foo":12}}`, "",
		},
		{
			func() json.Marshaler { v := customJSONMarshaler(10); return &v }(),
			`"Tendermint"`, "",
		},
		{strings.Contains, "", "unsupported type"},
	}

	for i, tt := range cases {
		blob, err := wire.MarshalJSON(tt.in)
		if tt.wantErr != "" {
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("#%d:\ngot:\n\t%q\nwant non-nil error containing\n\t%q", i,
					err, tt.wantErr)
			}
			continue
		}

		if err != nil {
			t.Errorf("#%d: unexpected error: %v", i, err)
			continue
		}
		if g, w := string(blob), tt.want; g != w {
			t.Errorf("#%d:\ngot:\n\t%s\nwant:\n\t%s", i, g, w)
		}
	}
}

func intPtr(i int) *int {
	return &i
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

type Car string
type Boat string
type Plane int

func (c Car) Move() error   { return nil }
func (b Boat) Move() error  { return nil }
func (p Plane) Move() error { return nil }
