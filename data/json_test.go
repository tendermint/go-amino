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

package data_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	data "github.com/tendermint/go-wire/data"
)

func TestSimpleJSON(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	cases := []struct {
		foo Fooer
	}{
		{foo: Bar{Name: "Fly"}},
		{foo: Baz{Name: "For Bar"}},
	}

	for _, tc := range cases {
		assert.NotEmpty(tc.foo.Foo())
		wrap := FooerS{tc.foo}
		parsed := FooerS{}
		d, err := json.Marshal(wrap)
		require.Nil(err, "%+v", err)
		err = json.Unmarshal(d, &parsed)
		require.Nil(err, "%+v", err)
		assert.Equal(tc.foo.Foo(), parsed.Foo())
	}
}

func TestNestedJSON(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	cases := []struct {
		expected string
		foo      Fooer
	}{
		{"Bar Fly", Bar{Name: "Fly"}},
		{"Foz Baz", Baz{Name: "For Bar"}},
		{"My: Bar None", Nested{"My", FooerS{Bar{"None"}}}},
	}

	for _, tc := range cases {
		assert.Equal(tc.expected, tc.foo.Foo())
		wrap := FooerS{tc.foo}
		parsed := FooerS{}
		// also works with indentation
		d, err := data.ToJSON(wrap)
		require.Nil(err, "%+v", err)
		err = json.Unmarshal(d, &parsed)
		require.Nil(err, "%+v", err)
		assert.Equal(tc.expected, parsed.Foo())
	}
}
