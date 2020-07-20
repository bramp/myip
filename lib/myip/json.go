package myip

import (
	"encoding/json"
	"net/http"
)

// ErrResponse is returned in the case of a error.
type ErrResponse struct {
	Error string `json:"error,omitempty"`
}

// JSONHandler does the lookups and returns the results as a JSON object.
func (s *DefaultServer) JSONHandler(w http.ResponseWriter, req *http.Request) {
	response, err := s.MyIPHandler(req)
	if err != nil {
		response = addInsights(req, response)
	}

	if err != nil {
		w.WriteHeader(500)
		s.writeJSON(w, req, &ErrResponse{err.Error()})
		return
	}

	s.writeJSON(w, req, response)
}

func (s *DefaultServer) writeJSON(w http.ResponseWriter, req *http.Request, obj interface{}) {
	w.Header().Set("Content-Type", "application/json")

	scheme := "http://"
	if req.URL.Scheme == "https" {
		scheme = "https://"
	}

	// TODO Consider setting this on all responses
	w.Header().Set("Access-Control-Allow-Origin", scheme+s.Config.Host)
	w.Header().Set("Vary", "Origin")

	// TODO Do something with the returned err
	json.NewEncoder(w).Encode(obj)
}
