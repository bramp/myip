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
	"bufio"
	"strings"
)

const (
	WHOIS_KEY = "whois:"
)

func ParseWhois(response string) (map[string]string, error) {
	m := map[string]string{}

	scanner := bufio.NewScanner(strings.NewReader(response))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "%") {
			continue
		}

		row := strings.Fields(line)
		if len(row) >= 2 {
			m[row[0]] = row[1]
		}
	}

	return m, scanner.Err()
}
