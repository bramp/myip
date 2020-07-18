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

package ua

import (
	"github.com/kylelemons/godebug/pretty"
	"github.com/ua-parser/uap-go/uaparser"
	"testing"
)

func TestDetermineUA(t *testing.T) {
	// Really simple (and possibly flakey) sanity check
	ua := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.113 Safari/537.36"
	got := DetermineUA(ua)
	want := &uaparser.Client{
		UserAgent: &uaparser.UserAgent{
			Family: "Chrome",
			Major:  "60",
			Minor:  "0",
			Patch:  "3112",
		},
		Os: &uaparser.Os{
			Family: "Mac OS X",
			Major:  "10",
			Minor:  "12",
			Patch:  "6",
		},
		Device: &uaparser.Device{
			Family: "Other",
		},
	}

	if diff := pretty.Compare(want, got); diff != "" {
		t.Errorf("DetermineUA(%q) diff: (-got +want)\n%s", ua, diff)
	}
}
