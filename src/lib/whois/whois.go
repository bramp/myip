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
	"fmt"
	domainr "github.com/domainr/whois"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log" // TODO REMOVE
	"strings"
	"time"
)

const (
	// TODO Strip the ':' from the key
	whoisKey = "whois:"

	// ianaWhoisServer is the address of the Internet Assigned Numbers Authority whois server.
	ianaWhoisServer = "whois.iana.org"

	// WhoisTimeout is the dial/read timeout for the whois requests.
	WhoisTimeout = 10 * time.Second
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

	body, err := QueryIPWhois(ctx, ipAddr)
	resp := &Response{
		Query: ipAddr,
		Body:  cleanupWhois(body),
	}
	if err != nil {
		resp.Error = err.Error()
	}

	return resp
}

// QueryWhois issues a WHOIS query to the given host.
// TODO Refactor to remove the appengine.log
func queryWhoisWithClient(ctx context.Context, client *domainr.Client, query, host string) (string, error) {

	if host == "whois.arin.net" {
		// ARIN's whois servers will reply with "Query terms are ambiguous" if the query
		// is not prefixed with a "n"
		query = "n " + query
	}

	request := &domainr.Request{
		Query: query,
		Host:  host,
	}
	if err := request.Prepare(); err != nil {
		return "", err
	}

	log.Infof(ctx, "Whois request %q from %q", query, host)

	response, err := client.Fetch(request)
	if err != nil {
		log.Warningf(ctx, "Whois failed %q from %q: %s", query, host, err)
		return "", err
	}

	log.Infof(ctx, "Whois response %q from %q:\n%s", query, host, response)
	return response.String(), err
}

// QueryIPWhois issues two whois queries, the first to find the right whois server,
// and the 2nd to that server.
func QueryIPWhois(ctx context.Context, ipAddr string) (string, error) {
	response, err := QueryWhois(ctx, ipAddr, ianaWhoisServer)

	// IANA returns a key value response with a "whois: ..." line to indicate the whois
	// server for the owner of this IP range.
	m, err := parseWhois(response)
	if err != nil {
		return "", err
	}

	host, found := m[whoisKey]
	if !found {
		return response, fmt.Errorf("no whois server found for %q", ipAddr)
	}

	return QueryWhois(ctx, ipAddr, host)
}
