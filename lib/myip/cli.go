package myip

import (
	"net/http"
	"strings"
	"text/template"

	"github.com/gorilla/mux"
)

var cliTmpl = template.Must(template.New("test").Parse(
	"IP: {{.RemoteAddr}}\n" +
		"{{range .RemoteAddrReverse.Names}}" +
		"DNS: {{.}}\n" +
		"{{end}}\n" +
		"WHOIS:\n" +
		"{{.RemoteAddrWhois.Body}}\n\n" +
		"Location: " +
		"{{.Location.City}} {{.Location.Region}} {{.Location.Country}}" +
		"{{if (and (ne .Location.Lat 0.0) (ne .Location.Long 0.0))}} ({{.Location.Lat}}, {{.Location.Long}}) {{end}}\n\n" +
		"ID: {{.RequestID}}\n"))

// isCLI returns true iif the request is coming from a cli tool, such as curl, or wget
func isCLI(req *http.Request, _ *mux.RouteMatch) bool {
	ua := req.Header.Get("User-Agent")
	return strings.HasPrefix(ua, "curl/") || strings.HasPrefix(ua, "Wget/")
}

// CLIHandler handles a CLI request to the service.
func (s *DefaultServer) CLIHandler(w http.ResponseWriter, req *http.Request) {
	response, err := s.MyIPHandler(req)

	w.Header().Set("Content-Type", "text/plain")

	if err == nil {
		err = cliTmpl.Execute(w, response)
		// Drop though with a new err
	}

	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
	}
}
