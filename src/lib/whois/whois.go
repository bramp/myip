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

// Handle generates a whois.Response
func Handle(ctx context.Context, ipAddr string) *Response {
	client := NewAppEngineWhoisClient(ctx) // We shouldn't store ctx in the client, but there is no alternative

	body, err := client.QueryIpWhois(ipAddr)
	resp := &Response{
		Query: ipAddr,
		Body:  strings.TrimSpace(body),
	}
	if err != nil {
		resp.Error = err.Error()
	}

	return resp
}
