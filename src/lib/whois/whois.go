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

package whois

import (
	"bufio"
	"golang.org/x/net/context"
	"strings"
)

const (
	// TODO Strip the ':' from the key
	whoisKey = "whois:"
)

// Response contains the Whois data we send to the user.
type Response struct {
	Query string

	// One of the following
	Body  string `json:",omitempty"`
	Error string `json:",omitempty"`
}

// parseWhois takes a whois response, and splits it into key-value pairs, so fields can easily
// be extracted.
func parseWhois(response string) (map[string]string, error) {
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

// longestCommonString searches the slice (element by element) to find any repeated data, and
// returns the start/end postition of the 2nd occurence of the largest common area.
// Runs in O(n^2) time
//
// This implementation was adapted from:
// 	 https://en.wikibooks.org/wiki/Algorithm_Implementation/Strings/Longest_common_substring
//   Under the Creative Commons Attribution-ShareAlike License.
func longestCommonString(lines []string) (int, int) {

	// Matrix to keep track of the longest match found starting on each line
	var m = make([][]int, 1+len(lines))
	for i := 0; i < len(m); i++ {
		m[i] = make([]int, 1+len(lines))
	}

	longest := 0
	yLongest := 0
	for x := 1; x < 1+len(lines); x++ {
		for y := x + 1; y < 1+len(lines); y++ {
			if lines[x-1] == lines[y-1] {
				m[x][y] = m[x-1][y-1] + 1
				if m[x][y] > longest {
					longest = m[x][y]
					yLongest = y
				}
			}
		}
	}

	return yLongest - longest, yLongest
}

// cleanupWhois trims whitespace, and excessive (but useless) text from the whois response
func cleanupWhois(response string) string {
	lines := []string{}

	scanner := bufio.NewScanner(strings.NewReader(response))
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}

	if scanner.Err() == nil {
		// TODO Repeat this multiple times, until there is no change.

		// If we parsed successfully we now try and dedup data in the string.
		// A good example of this, is that ARIN puts the same copyright header at the top and
		// bottom. Thus we can remove the bottom occurence.
		start, end := longestCommonString(lines)

		// Only delete if there is multiple lines of dup
		if end-start > 3 {
			response = ""
			for i, line := range lines {
				if i < start || i > end {
					response += line + "\n"
				}
			}
		}
	}

	return strings.TrimSpace(response)
}

// Handle generates a whois.Response
func Handle(ctx context.Context, ipAddr string) *Response {
	client := NewAppEngineWhoisClient(ctx) // We shouldn't store ctx in the client, but there is no alternative

	body, err := client.QueryIpWhois(ipAddr)
	resp := &Response{
		Query: ipAddr,
		Body:  cleanupWhois(body),
	}
	if err != nil {
		resp.Error = err.Error()
	}

	return resp
}
