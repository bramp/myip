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
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/ua-parser/uap-go/uaparser"
	"lib/conf"
	"lib/dns"
	"lib/location"
	"lib/whois"
)

/*
type Server struct {
	Config func() *config.Config
	HandleMyIP func(req *http.Request) (interface{}, error)
	HandleConfigJs func (w http.ResponseWriter, _ *http.Request)
	GetRemoteAddr func(req *http.Request) (string, error)
	//ServeJson func(req*http.Request) (interface{}, error)
}
*/

// Server is the interface all instances of the myip application should implement.
type Server interface {
	GetRemoteAddr(req *http.Request) (string, error)

	HandleMyIP(req *http.Request) (*Response, error)
	HandleConfigJs(w http.ResponseWriter, _ *http.Request)

	WriteJSON(w http.ResponseWriter, req *http.Request, obj interface{}, err error)
}

// DefaultServer is a default implementation of Server with some good defaults.
type DefaultServer struct {
	Config *conf.Config
}

// ErrResponse is returned in the case of a error.
type ErrResponse struct {
	Error string
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

type objHandler func(req *http.Request) (interface{}, error)

// Register this myip.Server. Should only be called once.
func Register(app Server) {
	r := mux.NewRouter()

	mainHandle := func(w http.ResponseWriter, req *http.Request) {
		response, err := app.HandleMyIP(req)
		if err != nil {
			response = addInsights(req, response)
		}
		app.WriteJSON(w, req, response, err)
	}

	r.Methods("GET").Path("/json").HandlerFunc(mainHandle)
	r.Methods("GET").Path("/config.js").HandlerFunc(app.HandleConfigJs)

	// App Engine and Compute Engine health checks.
	// TODO only set if compiled for app engine
	r.Methods("GET").Path("/_ah/health").HandlerFunc(healthCheckHandler)

	// Log all requests using the standard Apache format.
	http.Handle("/", handlers.CombinedLoggingHandler(os.Stderr, r))
}

// GetRemoteAddr returns the remote address, either the real one, or one passed via a header, or
// finally if in debug one passed as a query param.
func (s *DefaultServer) GetRemoteAddr(req *http.Request) (string, error) {
	remoteAddr := req.RemoteAddr

	// If debug allow replacing the host
	if host := req.URL.Query().Get("host"); host != "" && s.Config.Debug {
		return host, nil
	}

	if s.Config.IPHeader != "" {
		if addr := req.Header.Get(s.Config.IPHeader); addr != "" {
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

// WriteJSON takes the given obj and error, and returns appropriate JSON to the user
func (s *DefaultServer) WriteJSON(w http.ResponseWriter, req *http.Request, obj interface{}, err error) {
	if err != nil {
		w.WriteHeader(500)
		obj = &ErrResponse{err.Error()}
	}

	scheme := "http://"
	if req.TLS != nil {
		scheme = "https://"
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", scheme+s.Config.Host)
	json.NewEncoder(w).Encode(obj)
}
