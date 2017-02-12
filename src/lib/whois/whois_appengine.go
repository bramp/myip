// +build appengine

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
	"errors"
	"log"
	"net"
	"time"

	domainr "github.com/domainr/whois"
	"golang.org/x/net/context"
	"google.golang.org/appengine/socket"
)

type appEngineWhoisClient struct {
	domainr.Client
}

const (
	// WhoisTimeout is the dial/read timeout for the whois requests.
	WhoisTimeout = 10 * time.Second

	// ianaWhoisServer is the address of the Internet Assigned Numbers Authority whois server.
	ianaWhoisServer = "whois.iana.org"
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
		domainr.Client{
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

	request := &domainr.Request{
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
	response, err := client.QueryWhois(ipAddr, ianaWhoisServer)

	m, err := parseWhois(response)
	if err != nil {
		return "", err
	}

	host, found := m[whoisKey]
	if !found {
		return "", errors.New("no whois server found")
	}

	return client.QueryWhois(ipAddr, host)
}
