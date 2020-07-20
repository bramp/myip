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
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/ua-parser/uap-go/uaparser"
	"github.com/unrolled/secure"

	"bramp.net/myip/lib/conf"
	"bramp.net/myip/lib/dns"
	"bramp.net/myip/lib/location"
	"bramp.net/myip/lib/ua"
	"bramp.net/myip/lib/whois"
)

// Server is the interface all instances of the myip application should implement.
type Server interface {
	GetRemoteAddr(req *http.Request) (string, error)

	MyIPHandler(req *http.Request) (*Response, error)

	// TODO Merge CLI and JSON together, and use a different marshallers.
	// CLI index page
	CLIHandler(w http.ResponseWriter, req *http.Request)

	// JSON index page
	JSONHandler(w http.ResponseWriter, req *http.Request)

	// Web-app config
	ConfigJSHandler(w http.ResponseWriter, _ *http.Request)
}

// DefaultServer is a default implementation of Server with some good defaults.
type DefaultServer struct {
	Config *conf.Config
}

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

// Register this myip.Server. Should only be called once.
func Register(r *mux.Router, config *conf.Config) { // TODO Refactor so we don't need config here
	app := &DefaultServer{
		Config: config,
	}

	// Documented here: https://godoc.org/github.com/unrolled/secure#Options
	secureConfig := secure.Options{
		IsDevelopment: config.Debug,

		SSLRedirect: true,
		SSLHost:     "", // Use same host

		// Ensure the client is using HTTPS
		STSSeconds:           365 * 24 * 60 * 60,
		STSIncludeSubdomains: true,
		STSPreload:           true,

		FrameDeny:          true, // Don't allow the page embedded in a frame.
		ContentTypeNosniff: true, // Trust the Content-Type and don't second guess them.
		BrowserXssFilter:   true,

		// TODO Find CSP generator to make the next line shorter, and less error prone
		ContentSecurityPolicy: "default-src 'self';" +
			" connect-src *;" +
			" script-src 'self' www.google-analytics.com;" +
			" img-src data: 'self' www.google-analytics.com maps.googleapis.com;",
	}

	r.Use(secure.New(secureConfig).Handler)

	// Fetching with `curl`
	r.MatcherFunc(isCLI).HandlerFunc(app.CLIHandler)

	r.HandleFunc("/json", app.JSONHandler)
	r.HandleFunc("/config.js", app.ConfigJSHandler)

	// Serve the static content
	fs := http.FileServer(http.Dir("./static/"))
	r.PathPrefix("/").Handler(fs)
}

// GetRemoteAddr returns the remote address, either the real one, or if in debug mode one passed as a query param.
func (s *DefaultServer) GetRemoteAddr(req *http.Request) (string, error) {
	// If debug allow replacing the host
	if host := req.URL.Query().Get("host"); host != "" && s.Config.Debug {
		return host, nil
	}

	// Some systems (namely App Engine Flex) encode the remoteAddr with a port
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		// For now assume the RemoteAddr was just a addr (with no port)
		// TODO check if remoteAddr is a valid IPv6/IPv4 address
		return req.RemoteAddr, nil
	}

	return host, err
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
