// +build appengine

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
	"errors"
	"net"
	"time"

	domainr "github.com/domainr/whois"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/socket"
)

const (
	// WhoisTimeout is the dial/read timeout for the whois requests.
	WhoisTimeout = 10 * time.Second

	// ianaWhoisServer is the address of the Internet Assigned Numbers Authority whois server.
	ianaWhoisServer = "whois.iana.org"
)

func clientWithContext(ctx context.Context) *domainr.Client {
	return &domainr.Client{
		Dial: func(network, address string) (net.Conn, error) {
			deadline := time.Now().Add(WhoisTimeout)
			conn, err := socket.DialTimeout(ctx, network, address, WhoisTimeout)
			if err != nil {
				return nil, err
			}
			return conn, conn.SetDeadline(deadline)
		},
	}
}

// QueryWhois issues a WHOIS query to the given host.
func QueryWhois(ctx context.Context, query, host string) (string, error) {

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

	// Create the client on the fly (which is cheap) so we can use the passed in context
	client := clientWithContext(ctx)

	response, err := client.Fetch(request)
	if err != nil {
		log.Warningf(ctx, "Whois failed %q from %q: %s", query, host, err)
		return "", err
	}

	log.Infof(ctx, "Whois response %q from %q", query, host)
	return response.String(), err
}

// QueryIPWhois issues two whois queries, the first to find the right whois server,
// and the 2nd to that server.
func QueryIPWhois(ctx context.Context, ipAddr string) (string, error) {
	response, err := QueryWhois(ctx, ipAddr, ianaWhoisServer)

	m, err := parseWhois(response)
	if err != nil {
		return "", err
	}

	host, found := m[whoisKey]
	if !found {
		return "", errors.New("no whois server found for this ip address")
	}

	return QueryWhois(ctx, ipAddr, host)
}
