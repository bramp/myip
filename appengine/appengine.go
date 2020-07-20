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

// appengine provides a Google App Engine (Standard) specific implementation of myip
package main // import "bramp.net/myip/appengine"

import (
	"fmt"
	"net/http"
	"os"
	"sync"

	"bramp.net/myip/lib/conf"
	"bramp.net/myip/lib/dns"
	"bramp.net/myip/lib/location"
	"bramp.net/myip/lib/myip"
	"bramp.net/myip/lib/ua"
	"bramp.net/myip/lib/whois"

	log "github.com/sirupsen/logrus"
	"github.com/ua-parser/uap-go/uaparser"
	"google.golang.org/appengine"
)

var debugConfig = &conf.Config{
	Debug: true,

	Host:  "localhost:8080",
	Host4: "127.0.0.1:8080",
	Host6: "[::1]:8080",

	MapsAPIKey: "AIzaSyA6-HIkxuJEX6Hf3rzVx07no32YM3N5V9s",

	DisallowedHeaders: []string{"none"},
}

var prodConfig = &conf.Config{
	Debug: false,

	Host:  "ip.bramp.net",
	Host4: "ip4.bramp.net",
	Host6: "ip6.bramp.net",

	MapsAPIKey: "AIzaSyA6-HIkxuJEX6Hf3rzVx07no32YM3N5V9s",

	// If behind CloudFlare use the following:
	//IPHeader: "Cf-Connecting-Ip",
	//RequestIDHeader: "Cf-Ray",
}

var appengineDefaultConfig = &conf.Config{
	IPHeader: "X-Appengine-User-Ip",

	RequestIDHeader: "X-Cloud-Trace-Context",
	ProtoHeader:     "X-Forwarded-Proto",
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

func main() {
	config := debugConfig
	log.SetLevel(log.DebugLevel)

	if appengine.IsAppEngine() {
		config = prodConfig
		log.SetLevel(log.WarnLevel)
	}

	config, err := conf.ApplyDefaults(config, appengineDefaultConfig)
	if err != nil {
		log.Fatalf("Failed to ApplyDefaults: %s", err)
	}

	myip.Register(&server{
		myip.DefaultServer{
			Config: config,
		},
	}, config)

	http.HandleFunc("/_ah/warmup", func(w http.ResponseWriter, r *http.Request) {
		log.Println("warmup done")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failing ListenAndServe(%s): %s:", port, err)
	}
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

func (app *server) HandleMyIP(req *http.Request) (*myip.Response, error) {
	ctx := req.Context()
	wg := &sync.WaitGroup{}

	host, err := app.GetRemoteAddr(req)
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
		locationResponse = location.Handle(app.Config, req)
	})

	requestID := req.Header.Get(app.Config.RequestIDHeader)

	for _, remove := range app.Config.DisallowedHeaders {
		req.Header.Del(remove)
	}

	// Wait for all the responses to come back
	wg.Wait()

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
