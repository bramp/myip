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

package myip

import (
	"bytes"
	"net/http"
	"text/template"

	"bramp.net/myip/lib/conf"
)

var configTemplate = `
var VERSION = "{{.Version}}";
var BUILDTIME = "{{.BuildTime}}";

var MAIN_HOST = "{{.Host}}";

var SERVERS = {
   "IPv4": "{{.Host4}}",
   "IPv6": "{{.Host6}}"
};

var MAPS_API_KEY = "{{.MapsAPIKey}}";`

type configAndVersion struct {
	*conf.Config
	Version   string
	BuildTime string
}

// HandleConfigJs returns the config javascript.
func (s *DefaultServer) HandleConfigJs(w http.ResponseWriter, _ *http.Request) {
	tmpl, err := template.New("config").Parse(configTemplate)

	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	data := struct {
		*conf.Config
		Version   string
		BuildTime string
	}{
		s.Config, Version, BuildTime,
	}

	// Buffer the output so we can put a error at the front if it fails
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		// TODO Consider writing out a nice error js field, instead of invalid js.
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	// TODO Eventually add a long cache-expire time
	w.Header().Set("Content-Type", "text/javascript")
	buf.WriteTo(w)
}
