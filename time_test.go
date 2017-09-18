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

package wire

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type Input struct {
	Date time.Time `json:"date"`
}

func TestJSONTimeParse(t *testing.T) {
	cases := []struct {
		input    string
		expected time.Time
		encoded  string
	}{
		{
			"2017-03-31T16:45:15Z",
			time.Date(2017, 3, 31, 16, 45, 15, 0, time.UTC),
			"2017-03-31T16:45:15.000Z",
		},
		{
			"2017-03-31T16:45:15.972Z",
			time.Date(2017, 3, 31, 16, 45, 15, 972000000, time.UTC),
			"2017-03-31T16:45:15.972Z",
		},
		{
			"2017-03-31T16:45:15.972167Z",
			time.Date(2017, 3, 31, 16, 45, 15, 972167000, time.UTC),
			"2017-03-31T16:45:15.972Z",
		},
	}

	for _, tc := range cases {
		var err error
		var parsed Input
		data := []byte(fmt.Sprintf(`{"date":"%s"}`, tc.input))
		ReadJSONPtr(&parsed, data, &err)
		if assert.Nil(t, err, "%s: %+v", tc.input, err) {
			assert.Equal(t, tc.expected, parsed.Date)
			out := JSONBytes(parsed)
			assert.Equal(t, fmt.Sprintf(`{"date":"%s"}`, tc.encoded), string(out))
		}
	}
}
