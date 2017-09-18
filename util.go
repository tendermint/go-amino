// Copyright 2015 Tendermint. All Rights Reserved.
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

package wire

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"

	"golang.org/x/crypto/ripemd160"

	cmn "github.com/tendermint/tmlibs/common"
)

func BinaryBytes(o interface{}) []byte {
	w, n, err := new(bytes.Buffer), new(int), new(error)
	WriteBinary(o, w, n, err)
	if *err != nil {
		cmn.PanicSanity(*err)
	}
	return w.Bytes()
}

// ptr: a pointer to the object to be filled
func ReadBinaryBytes(d []byte, ptr interface{}) error {
	r, n, err := bytes.NewBuffer(d), new(int), new(error)
	ReadBinaryPtr(ptr, r, len(d), n, err)
	return *err
}

func JSONBytes(o interface{}) []byte {
	w, n, err := new(bytes.Buffer), new(int), new(error)
	WriteJSON(o, w, n, err)
	if *err != nil {
		cmn.PanicSanity(*err)
	}
	return w.Bytes()
}

// NOTE: inefficient
func JSONBytesPretty(o interface{}) []byte {
	jsonBytes := JSONBytes(o)
	var object interface{}
	err := json.Unmarshal(jsonBytes, &object)
	if err != nil {
		cmn.PanicSanity(err)
	}
	jsonBytes, err = json.MarshalIndent(object, "", "\t")
	if err != nil {
		cmn.PanicSanity(err)
	}
	return jsonBytes
}

// ptr: a pointer to the object to be filled
func ReadJSONBytes(d []byte, ptr interface{}) (err error) {
	ReadJSONPtr(ptr, d, &err)
	return
}

// NOTE: does not care about the type, only the binary representation.
func BinaryEqual(a, b interface{}) bool {
	aBytes := BinaryBytes(a)
	bBytes := BinaryBytes(b)
	return bytes.Equal(aBytes, bBytes)
}

// NOTE: does not care about the type, only the binary representation.
func BinaryCompare(a, b interface{}) int {
	aBytes := BinaryBytes(a)
	bBytes := BinaryBytes(b)
	return bytes.Compare(aBytes, bBytes)
}

// NOTE: only use this if you need 32 bytes.
func BinarySha256(o interface{}) []byte {
	hasher, n, err := sha256.New(), new(int), new(error)
	WriteBinary(o, hasher, n, err)
	if *err != nil {
		cmn.PanicSanity(*err)
	}
	return hasher.Sum(nil)
}

// NOTE: The default hash function is Ripemd160.
func BinaryRipemd160(o interface{}) []byte {
	hasher, n, err := ripemd160.New(), new(int), new(error)
	WriteBinary(o, hasher, n, err)
	if *err != nil {
		cmn.PanicSanity(*err)
	}
	return hasher.Sum(nil)
}
