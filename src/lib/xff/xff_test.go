// Copyright 2017 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package xff

import (
	"reflect"
	"testing"
)

func TestParseXFF(t *testing.T) {
	data := []struct {
		header string
		want   []string
	}{
		{header: "2601:646:c200:b466:0:0:0:1, 2607:f8b0:4005:801::2014", want: []string{"2601:646:c200:b466:0:0:0:1", "2607:f8b0:4005:801::2014"}},
		{header: "1.2.3.4", want: []string{"1.2.3.4"}},
		{header: "1.2.3.4, 8.8.8.8", want: []string{"1.2.3.4", "8.8.8.8"}},
		{header: "1.2.3.4, 192.168.0.1", want: []string{"1.2.3.4", "192.168.0.1"}},
		{header: "", want: []string{}},
	}

	for _, test := range data {
		got := parseXFF(test.header)
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("parseXFF(%q) = %q want %q", test.header, got, test.want)
		}
	}
}
