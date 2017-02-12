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

// +build appengine
package whois

import (
	"bufio"
	"errors"
	"log"
	"net"
	"strings"
	"time"

	"github.com/domainr/whois"
	"golang.org/x/net/context"
	"google.golang.org/appengine/socket"
)

type appEngineWhoisClient struct {
	whois.Client
}

const (
	IANA_WHOIS   = "whois.iana.org"
	WhoisTimeout = 10 * time.Second
)

func NewAppEngineWhoisClient(ctx context.Context) *appEngineWhoisClient {
	dial := func(network, address string) (net.Conn, error) {
		deadline := time.Now().Add(WhoisTimeout)
		conn, err := socket.DialTimeout(ctx, network, address, WhoisTimeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(deadline)
		return conn, nil
	}

	return &appEngineWhoisClient{
		whois.Client{
			Dial: dial,
		},
	}
}

// TODO pass context to QueryWhois
func (client *appEngineWhoisClient) QueryWhois(query, host string) (string, error) {

	if host == "whois.arin.net" {
		// ARIN's whois servers will reply with "Query terms are ambiguous" if the query
		// is not prefixed with a n
		query = "n " + query
	}

	request := &whois.Request{
		Query: query,
		Host:  host,
	}
	if err := request.Prepare(); err != nil {
		return "", err
	}

	log.Printf("Whois requesting %s from %s", query, host)

	response, err := client.Fetch(request)
	if err != nil {
		return "", err
	}

	log.Printf("Whois response: %s", response.String())

	return response.String(), err
}

// queryIpWhois issues two whois queries, the first to find the right whois server, and the 2nd to
// that server.
func (client *appEngineWhoisClient) QueryIpWhois(ipAddr string) (string, error) {
	response, err := client.QueryWhois(ipAddr, IANA_WHOIS)

	m, err := parseWhois(response)
	if err != nil {
		return "", err
	}

	host, found := m[WHOIS_KEY]
	if !found {
		return "", errors.New("no whois server found")
	}

	return client.QueryWhois(ipAddr, host)
}

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
