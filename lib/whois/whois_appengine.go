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
	"net"
	"time"

	domainr "github.com/domainr/whois"
	"golang.org/x/net/context"
	"google.golang.org/appengine/socket"
)

func QueryWhois(ctx context.Context, query, host string) (string, error) {
	// Create a client that uses the appengine sockets
	client := &domainr.Client{
		Dial: func(network, address string) (net.Conn, error) {
			deadline := time.Now().Add(WhoisTimeout)
			conn, err := socket.DialTimeout(ctx, network, address, WhoisTimeout)
			if err != nil {
				return nil, err
			}
			return conn, conn.SetDeadline(deadline)
		},
	}

	return queryWhoisWithClient(ctx, client, query, host)
}
