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

// Google App Engine (Standard) specific implementation
package main

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
)

func ternary(b bool, t, f string) string {
	if b {
		return t
	}
	return f
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

func loadConfig() *conf.Config {
	config := prodConfig
	if appengine.IsDevAppServer() {
		config = debugConfig
	}

	config, err := conf.ApplyDefaults(config, appengineDefaultConfig)
	if err != nil {
		// No choice but to panic here
		panic(err.Error())
	}

	return config
}

// TODO Make this not app engine specific by factoring out the HandleCode.
func handleMyIP(req *http.Request) (interface{}, error) {

	ctx := appengine.NewContext(req)

	host, err := getRemoteAddr(req)
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
			dnsResp = dns.HandleReverseDns(ctx, host)
		}

		if req.URL.Query().Get("whois") != "false" {
			whoisResp = whois.Handle(ctx, host)
		}
	}

	var userAgentClient *uaparser.Client
	if useragent := req.Header.Get("User-Agent"); useragent != "" {
		userAgentClient = ua.DetermineUA(useragent)
	}

	locationResponse := location.Handle(config, req)

	requestId := req.Header.Get(config.RequestIDHeader)

	for _, remove := range config.DisallowedHeaders {
		req.Header.Del(remove)
	}

	return addInsights(req, &myIPResponse{
		RequestID: requestId,

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
	}), nil
}
