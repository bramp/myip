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

// Package myip is a web application to returns the client's IP address and other information.
// by Andrew Brampton (https://bramp.net/)
package myip

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/ua-parser/uap-go/uaparser"

	"bramp.net/myip/lib/dns"
	"bramp.net/myip/lib/location"
	"bramp.net/myip/lib/ua"
	"bramp.net/myip/lib/whois"
)

// Response is a normal response.
type Response struct {
	RequestID string `json:",omitempty"`

	RemoteAddr        string
	RemoteAddrFamily  string
	RemoteAddrReverse *dns.Response   `json:",omitempty"`
	RemoteAddrWhois   *whois.Response `json:",omitempty"`

	ActualRemoteAddr string `json:",omitempty"` // The actual one we observed

	Method string
	URL    string
	Proto  string

	Header http.Header

	Location  *location.Response `json:",omitempty"`
	UserAgent *uaparser.Client   `json:",omitempty"` // TODO Create a ua.Response

	Insights map[string]string `json:",omitempty"`
}

// MyIPHandler is the main code to handle a IP lookup.
func (s *DefaultServer) MyIPHandler(req *http.Request) (*Response, error) {
	ctx := req.Context()
	wg := &sync.WaitGroup{}

	host, err := s.GetRemoteAddr(req)
	if err != nil {
		return nil, fmt.Errorf("getting remote addr: %s", err)
	}

	family := "IPv4"
	if f := req.URL.Query().Get("family"); f != "" {
		// TODO Change this to actually lookup the family of `host`
		family = f
	}

	var dnsResp *dns.Response
	var whoisResp *whois.Response
	var locationResponse *location.Response
	var userAgentClient *uaparser.Client // TODO change this to be a ua.Response

	if host != "" {
		if req.URL.Query().Get("reverse") != "false" {
			addToWg(wg, func() {
				dnsResp = dns.HandleReverseDNS(ctx, host)
			})
		}

		if req.URL.Query().Get("whois") != "false" {
			addToWg(wg, func() {
				whoisResp = whois.Handle(ctx, host)
			})
		}
	}

	if req.URL.Query().Get("ua") != "false" {
		if useragent := req.Header.Get("User-Agent"); useragent != "" {
			addToWg(wg, func() {
				userAgentClient = ua.DetermineUA(useragent)
			})
		}
	}

	addToWg(wg, func() {
		locationResponse = location.Handle(s.Config, req)
	})

	requestID := req.Header.Get(s.Config.RequestIDHeader)

	// Remove all headers we don't want to display to the user
	for _, remove := range s.Config.DisallowedHeaders {
		req.Header.Del(remove)
	}

	// Wait for all the responses to come back
	wg.Wait()

	return &Response{
		RequestID: requestID,

		RemoteAddr:        host,
		RemoteAddrFamily:  family,
		RemoteAddrReverse: dnsResp,
		RemoteAddrWhois:   whoisResp,

		ActualRemoteAddr: req.RemoteAddr,

		UserAgent: userAgentClient,
		Location:  locationResponse,

		Method: req.Method,
		URL:    req.URL.String(),
		Proto:  req.Proto,
		Header: req.Header,
	}, nil
}

// addToWg executes the function in a new gorountine and adds it to the WaitGroup, calling wg.Done
// when finished. This makes it a little eaiser to use the WaitGroup.
func addToWg(wg *sync.WaitGroup, f func()) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		f()
	}()
}
