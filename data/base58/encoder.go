// Copyright 2017 Tendermint. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package base58

import (
	"encoding/json"

	"github.com/pkg/errors"
	data "github.com/tendermint/go-wire/data"
)

var (
	BTCEncoder    data.ByteEncoder = base58Encoder{BTCAlphabet}
	FlickrEncoder                  = base58Encoder{FlickrAlphabet}
)

// base58Encoder implements ByteEncoder encoding the slice as
// base58 url-safe encoding
type base58Encoder struct {
	alphabet string
}

func (e base58Encoder) _assertByteEncoder() data.ByteEncoder {
	return e
}

func (e base58Encoder) Unmarshal(dst *[]byte, src []byte) (err error) {
	var s string
	err = json.Unmarshal(src, &s)
	if err != nil {
		return errors.Wrap(err, "parse string")
	}
	*dst, err = DecodeAlphabet(s, e.alphabet)
	return err
}

func (e base58Encoder) Marshal(bytes []byte) ([]byte, error) {
	s := EncodeAlphabet(bytes, e.alphabet)
	return json.Marshal(s)
}
