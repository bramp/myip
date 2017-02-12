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
// Google App Engine (Standard) specific implementation
//
package main

import (
	"fmt"
	"net/http"

	"golang.org/x/net/context"
	"google.golang.org/appengine"

	"encoding/json"
	"lib/dns"
	"lib/ua"
	"lib/whois"
	"strings"

	"github.com/ua-parser/uap-go/uaparser"
)

func ternary(b bool, t, f string) string {
	if b {
		return t
	}
	return f
}

var appengineDefaultConfig = &Config{
	Debug: appengine.IsDevAppServer(),

	IpHeader:        "",
	RequestIdHeader: ternary(appengine.IsDevAppServer(), "X-Appengine-Request-Log-Id", "X-Cloud-Trace-Context"),
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

// applyDefaults returns a new config with any zero field in config, set to the default value.
func applyDefaults(config, defaults *Config) (*Config, error) {

	configCopy := &Config{}
	*configCopy = *defaults // Copy default

	// Hack, Marshal the config, and Unmarshalling it over a copy of the defaults. Thus replacing
	// any fields that were explictly set, and zero
	tmp, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(tmp, configCopy); err != nil {
		return nil, err
	}

	return configCopy, nil
}

func loadConfig() *Config {
	config := prodConfig
	if appengine.IsDevAppServer() {
		config = debugConfig
	}

	config, err := applyDefaults(config, appengineDefaultConfig)
	if err != nil {
		// No choice but to panic here
		panic(err.Error())
	}

	return config
}

// TODO Make this not app engine specific
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

	var dnsResp *dnsResponse
	var whoisResp *whoisResponse

	if host != "" {
		if req.URL.Query().Get("reverse") != "false" {
			dnsResp = handleReverseDns(ctx, host)
		}

		if req.URL.Query().Get("whois") != "false" {
			whoisResp = handleWhois(ctx, host)
		}
	}

	var userAgentClient *uaparser.Client
	if useragent := req.Header.Get("User-Agent"); useragent != "" {
		userAgentClient = ua.DetermineUA(useragent)
	}

	locationResponse := handleLocation(req)

	requestId := req.Header.Get(config.RequestIdHeader)

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

func handleReverseDns(ctx context.Context, ipAddr string) *dnsResponse {
	names, err := dns.LookupAddr(ctx, ipAddr)
	resp := &dnsResponse{
		Query: ipAddr,
		Names: names,
	}
	if err != nil {
		resp.Error = err.Error()
	}

	return resp
}

func handleWhois(ctx context.Context, ipAddr string) *whoisResponse {
	client := whois.NewAppEngineWhoisClient(ctx) // We shouldn't store ctx in the client, but there is no alternative

	body, err := client.QueryIpWhois(ipAddr)
	resp := &whoisResponse{
		Query: ipAddr,
		Body:  strings.TrimSpace(body),
	}
	if err != nil {
		resp.Error = err.Error()
	}

	return resp
}
