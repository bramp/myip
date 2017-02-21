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

// Package appengine provides a Google App Engine (Standard) specific implementation of myip
package appengine

import (
	"fmt"
	"net/http"

	"google.golang.org/appengine"

	"lib/dns"
	"lib/ua"
	"lib/whois"

	"github.com/ua-parser/uap-go/uaparser"
	"lib/conf"
	"lib/location"
	"lib/myip"
)

func ternary(b bool, t, f string) string {
	if b {
		return t
	}
	return f
}

var debugConfig = &conf.Config{
	Host:  "http://localhost:8080",
	Host4: "http://127.0.0.1:8080",
	Host6: "http://[::1]:8080",

	MapsAPIKey: "AIzaSyA6-HIkxuJEX6Hf3rzVx07no32YM3N5V9s",

	DisallowedHeaders: []string{"none"},
}

var prodConfig = &conf.Config{
	Host:  "http://ip.bramp.net",
	Host4: "http://ip4.bramp.net",
	Host6: "http://ip6.bramp.net",

	MapsAPIKey: "AIzaSyA6-HIkxuJEX6Hf3rzVx07no32YM3N5V9s",

	// If behind CloudFlare use the following:
	//IPHeader: "Cf-Connecting-Ip",
	//RequestIDHeader: "Cf-Ray",
}

var appengineDefaultConfig = &conf.Config{
	Debug: appengine.IsDevAppServer(),

	IPHeader:        "",
	RequestIDHeader: ternary(appengine.IsDevAppServer(), "X-Appengine-Request-Log-Id", "X-Cloud-Trace-Context"),
	LatLongHeader:   "X-Appengine-Citylatlong",
	CityHeader:      "X-Appengine-City",

	DisallowedHeaders: []string{
		"X-Appengine-Default-Namespace",
		"X-Appengine-Request-Id-Hash",
		"X-Appengine-Request-Log-Id",
		"X-Appengine-Default-Version-Hostname",
		"X-Appengine-User-Email",
		"X-Appengine-User-Id",
		"X-Appengine-User-Is-Admin",
		"X-Appengine-User-Nickname",
		"X-Appengine-User-Organization",
		"X-Appengine-Server-Name",
		"X-Appengine-Server-Port",
		"X-Appengine-Server-Protocol",
		"X-Appengine-Server-Software",
		"X-Appengine-Remote-Addr",

		"X-Cloud-Trace-Context",
		"X-Google-Apps-Metadata",
		"X-Zoo",

		"Cf-Connecting-Ip",
		"Cf-Ipcountry",
		"Cf-Ray",
		"Cf-Visitor",
	},
}

type server struct {
	myip.DefaultServer
}

func init() {
	config := prodConfig
	if appengine.IsDevAppServer() {
		config = debugConfig
	}

	config, err := conf.ApplyDefaults(config, appengineDefaultConfig)
	if err != nil {
		// No choice but to panic here
		panic(err.Error())
	}

	myip.Register(&server{
		myip.DefaultServer{config},
	})
}

// TODO Make this not app engine specific by factoring out the HandleCode.
func (app *server) HandleMyIP(req *http.Request) (*myip.Response, error) {

	ctx := appengine.NewContext(req)

	host, err := app.GetRemoteAddr(req)
	if err != nil {
		return nil, fmt.Errorf("getting remote addr: %s", err)
	}

	family := "IPv4"
	if f := req.URL.Query().Get("family"); f != "" {
		family = f
	}

	var dnsResp *dns.Response
	var whoisResp *whois.Response

	if host != "" {
		if req.URL.Query().Get("reverse") != "false" {
			dnsResp = dns.HandleReverseDNS(ctx, host)
		}

		if req.URL.Query().Get("whois") != "false" {
			whoisResp = whois.Handle(ctx, host)
		}
	}

	var userAgentClient *uaparser.Client
	if useragent := req.Header.Get("User-Agent"); useragent != "" {
		userAgentClient = ua.DetermineUA(useragent)
	}

	locationResponse := location.Handle(app.Config, req)

	requestID := req.Header.Get(app.Config.RequestIDHeader)

	for _, remove := range app.Config.DisallowedHeaders {
		req.Header.Del(remove)
	}

	return &myip.Response{
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
