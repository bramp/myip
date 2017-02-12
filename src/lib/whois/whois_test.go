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

package whois

import (
	"io/ioutil"
	"path"
	"testing"
)

func TestParseWhois(t *testing.T) {
	data := []struct {
		query  string
		want   string
		result string
	}{
		{query: "1.2.3.4", want: "whois.apnic.net", result: "whois-1.txt"},
		{query: "8.8.8.8", want: "whois.arin.net", result: "whois-2.txt"},
		{query: "8.8.8.8", want: "", result: "whois-3.txt"},
		{query: "2601:646:c200:b466:0:0:0:1", want: "whois.arin.net", result: "whois-4.txt"},
	}

	for _, test := range data {
		input, err := ioutil.ReadFile(path.Join("testdata", test.result))
		if err != nil {
			t.Fatalf("Failed to read test data %q: %s", test.result, err)
		}
		m, err := parseWhois(string(input))
		if err != nil {
			t.Errorf("parseWhois(%q) err: %q, want nil", test.result, err)
			continue
		}
		if m[WHOIS_KEY] != test.want {
			t.Errorf("parseWhois(%q)[%q] = %q, want %q", test.result, WHOIS_KEY, m[WHOIS_KEY], test.want)
		}
	}

}
