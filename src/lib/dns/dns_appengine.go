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

package dns

import (
	"fmt"
	"math/rand"
	"net"
	"time"

	mdns "github.com/miekg/dns"
	"golang.org/x/net/context"
	"google.golang.org/appengine/socket"
)

// TODO Change this to be newAppEngineClient (that looks like the dnsClient)
func appEngineExchange(ctx context.Context, m *mdns.Msg) (*mdns.Msg, error) {

	// Pick a server at random
	server := dnsServers[rand.Intn(len(dnsServers))]
	server = net.JoinHostPort(server, "53")

	c, err := socket.DialTimeout(ctx, "udp", server, dnsTimeout)
	if err != nil {
		return nil, err
	}

	co := &mdns.Conn{Conn: c}
	defer co.Close()

	deadline := time.Now().Add(dnsTimeout)

	co.SetWriteDeadline(deadline)
	if err = co.WriteMsg(m); err != nil {
		return nil, err
	}

	co.SetReadDeadline(deadline)
	return co.ReadMsg()
}

// LookupAddr performs a reverse lookup for the given address, returning a list
// of names mapping to that address. It uses the github.com/miekg/dns library instead of the native
// net.LookupAddr which does not works on a standard App Engine.
//
// Note: App Engine does not seem to support net.LookupAddr(ipAddr) it returns "on [::1]:53: dial udp [::1]:53: socket: operation not permitted"
// TODO: Change ipAddr to be a slice
func LookupAddr(ctx context.Context, ipAddr string) ([]string, error) {

	// Special case localhost
	if ip := net.ParseIP(ipAddr); ip.IsLoopback() {
		if ip.To4() != nil {
			return []string{"localhost"}, nil
		}
		return []string{"ip6-localhost"}, nil
	}

	name, err := mdns.ReverseAddr(ipAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to ReverseAddr(%q): %s", ipAddr, err)
	}

	m := new(mdns.Msg)
	m.Id = mdns.Id()
	m.RecursionDesired = true
	m.Question = make([]mdns.Question, 1)
	m.Question[0] = mdns.Question{name, mdns.TypePTR, mdns.ClassINET}

	// TODO Add some kind of retry logic (for lost UDP packets)

	in, err := appEngineExchange(ctx, m)
	if err != nil {
		return nil, err
	}

	result := []string{}
	for _, record := range in.Answer {
		if t, ok := record.(*mdns.PTR); ok {
			result = append(result, t.Ptr)
		}
	}

	return result, err
}

func HandleReverseDns(ctx context.Context, ipAddr string) *Response {
	names, err := LookupAddr(ctx, ipAddr)
	resp := &Response{
		Query: ipAddr,
		Names: names,
	}
	if err != nil {
		resp.Error = err.Error()
	}

	return resp
}
