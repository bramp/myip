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

// myip is a web application to returns the client's IP address and other information.
// by Andrew Brampton (https://bramp.net/)
package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"

	"text/template"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/ua-parser/uap-go/uaparser"
	"lib/conf"
	"lib/dns"
	"lib/location"
	"lib/whois"
)

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

var config = loadConfig()

type errResponse struct {
	Error string
}

type myIPResponse struct {
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

func init() {
	registerHandlers()
}

func registerHandlers() {
	r := mux.NewRouter()

	r.Methods("GET").Path("/json").Handler(app(handleMyIP))
	r.Methods("GET").Path("/config.js").HandlerFunc(handleConfigJs)

	// App Engine and Compute Engine health checks.
	// TODO only set if compiled for app engine
	r.Methods("GET").Path("/_ah/health").HandlerFunc(healthCheckHandler)

	// Log all requests using the standard Apache format.
	http.Handle("/", handlers.CombinedLoggingHandler(os.Stderr, r))
}

func getRemoteAddr(req *http.Request) (string, error) {

	remoteAddr := req.RemoteAddr

	// If debug allow replacing the host
	if host := req.URL.Query().Get("host"); host != "" && config.Debug {
		return host, nil
	}

	if config.IPHeader != "" {
		if addr := req.Header.Get(config.IPHeader); addr != "" {
			remoteAddr = addr
		}
	}

	// Some systems (namely App Engine Flex encode the remoteAddr with a port)
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		// for now assume the RemoteAddr was just a addr (with no port)
		// TODO check if remoteAddr is a valid IPv6/IPv4 address
		return remoteAddr, nil
	}

	return host, err
}

func healthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprint(w, "ok")
}

func handleConfigJs(w http.ResponseWriter, _ *http.Request) {
	// TODO Eventually add a long cache-expire time

	tmpl, err := template.New("config").Parse(`
	   var SERVERS = {
		   "IPv4": "{{.Host4}}",
		   "IPv6": "{{.Host6}}"
	   };

	   var MAPS_API_KEY = "{{.MapsApiKey}}";
   `)

	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "text/javascript")
	err = tmpl.Execute(w, config)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
}

type app func(*http.Request) (interface{}, error)

func (fn app) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	obj, err := fn(r)
	if err != nil {
		w.WriteHeader(500)
		obj = &errResponse{err.Error()}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", config.Host)
	json.NewEncoder(w).Encode(obj)
}
