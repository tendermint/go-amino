package data_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	data "github.com/tendermint/go-data"
)

func TestEncoders(t *testing.T) {
	assert := assert.New(t)

	hex := data.HexEncoder
	b64 := data.B64Encoder
	rb64 := data.RawB64Encoder
	cases := []struct {
		encoder         data.ByteEncoder
		input, expected []byte
	}{
		// hexidecimal
		{hex, []byte(`"1a2b3c4d"`), []byte{0x1a, 0x2b, 0x3c, 0x4d}},
		{hex, []byte(`"de14"`), []byte{0xde, 0x14}},
		// these are errors
		{hex, []byte(`0123`), nil},     // not in quotes
		{hex, []byte(`"dewq12"`), nil}, // invalid chars
		{hex, []byte(`"abc"`), nil},    // uneven length

		// base64
		{b64, []byte(`"Zm9v"`), []byte("foo")},
		{b64, []byte(`"RCEuM3M="`), []byte("D!.3s")},
		// make sure url encoding!
		{b64, []byte(`"D4_a--1="`), []byte{0x0f, 0x8f, 0xda, 0xfb, 0xed}},
		// these are errors
		{b64, []byte(`"D4/a++1="`), nil}, // non-url encoding
		{b64, []byte(`0123`), nil},       // not in quotes
		{b64, []byte(`"hey!"`), nil},     // invalid chars
		{b64, []byte(`"abc"`), nil},      // length%4 != 0

		// raw base64
		{rb64, []byte(`"Zm9v"`), []byte("foo")},
		{rb64, []byte(`"RCEuM3M"`), []byte("D!.3s")},
		// make sure url encoding!
		{rb64, []byte(`"D4_a--1"`), []byte{0x0f, 0x8f, 0xda, 0xfb, 0xed}},
		// these are errors
		{rb64, []byte(`"D4/a++1"`), nil}, // non-url encoding
		{rb64, []byte(`0123`), nil},      // not in quotes
		{rb64, []byte(`"hey!"`), nil},    // invalid chars
		{rb64, []byte(`"abc="`), nil},    // with padding

	}

	for _, tc := range cases {
		var output []byte
		err := tc.encoder.Unmarshal(&output, tc.input)
		if tc.expected == nil {
			assert.NotNil(err, tc.input)
		} else if assert.Nil(err, "%s: %+v", tc.input, err) {
			assert.Equal(tc.expected, output, tc.input)
		}
	}
}

// BData can be encoded/decoded
type BData struct {
	Count int
	Data  data.Bytes
}

// BView is to unmarshall and check the encoding
type BView struct {
	Count int
	Data  string
}

func TestBytes(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	cases := []struct {
		encoder  data.ByteEncoder
		data     data.Bytes
		expected string
	}{
		{data.HexEncoder, []byte{0x1a, 0x2b, 0x3c, 0x4d}, "1a2b3c4d"},
		{data.B64Encoder, []byte("D!.3s"), "RCEuM3M="},
		{data.RawB64Encoder, []byte("D!.3s"), "RCEuM3M"},
	}

	for i, tc := range cases {
		data.Encoder = tc.encoder
		// encode the data
		in := BData{Count: 15, Data: tc.data}
		d, err := json.Marshal(in)
		require.Nil(err, "%d: %+v", i, err)
		// recover the data
		out := BData{}
		err = json.Unmarshal(d, &out)
		require.Nil(err, "%d: %+v", i, err)
		assert.Equal(in.Count, out.Count, "%d", i)
		assert.Equal(in.Data, out.Data, "%d", i)
		// check the encoding
		view := BView{}
		err = json.Unmarshal(d, &view)
		require.Nil(err, "%d: %+v", i, err)
		assert.Equal(tc.expected, view.Data)
	}
}
