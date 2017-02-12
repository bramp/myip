// Copyright 2015 Google Inc. All Rights Reserved.
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

package main

import (
	"testing"
)

func TestParseLatLong(t *testing.T) {
	data := []struct {
		input             string
		wantLat, wantLong float64
	}{
		{input: "37.562992,-122.325525", wantLat: 37.562992, wantLong: -122.325525},
	}

	for _, test := range data {
		lat, long, err := parseLatLong(test.input)
		if err != nil {
			t.Errorf("parseLatLong(%q) err: %q, want nil", test.input, err)
			continue
		}
		if lat != test.wantLat || long != test.wantLong {
			t.Errorf("parseLatLong(%q) = (%v, %v), want (%v, %v)", test.input, lat, long, test.wantLat, test.wantLong)
		}
	}

}
