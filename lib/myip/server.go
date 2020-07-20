package myip

import (
	"net"
	"net/http"

	"bramp.net/myip/lib/conf"
	"github.com/gorilla/mux"
	"github.com/unrolled/secure"
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

const host = "Host"

// DefaultServer is a default implementation of Server with some good defaults.
type DefaultServer struct {
	Config *conf.Config
}

// URLHeaders sets both the scheme and host in the Request.URL
func URLHeaders(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		r.URL.Scheme = scheme

		if r.Host != "" {
			// Set the host with the value received in the request
			r.URL.Host = r.Host
		}
		// Call the next handler in the chain.
		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

// Register this myip.Server. Should only be called once.
func Register(r *mux.Router, config *conf.Config) { // TODO Refactor so we don't need config here
	app := &DefaultServer{
		Config: config,
	}

	// Documented here: https://godoc.org/github.com/unrolled/secure#Options
	secureConfig := secure.Options{
		IsDevelopment: config.Debug,

		// TODO Fix this (it causes constant 301s) every since I implemented URLHeaders
		// SSLRedirect: true,
		// SSLHost:     "", // Use same host

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

	r.Use(URLHeaders)
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
