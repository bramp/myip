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

package conf

import "encoding/json"

// Config contains all the configuration options for this application.
// TODO Turn this into a config file that gets parsed onstartup
type Config struct {
	// The build version
	Version string

	// The build time
	BuildTime string

	Host  string `json:",omitempty"`
	Host4 string `json:",omitempty"`
	Host6 string `json:",omitempty"`

	// Debug enables unsafe options for debugging
	Debug bool `json:",omitempty"`

	// LatLongHeader is the header with the LatLong information
	// Examples:
	//   "X-Appengine-Citylatlong" for App Engine (Standard)
	LatLongHeader string `json:",omitempty"`

	// CityHeader is the header with the city information
	// Examples:
	//   "Cf-Ipcountry" for CloudFlare
	//   "X-Appengine-City" for App Engine (Standard)
	CityHeader string `json:",omitempty"`

	// TODO Document
	RegionHeader  string `json:",omitempty"`
	CountryHeader string `json:",omitempty"`

	// RequestIDHeader is the header with the Request ID
	// Examples:
	//   "Cf-Ray" for CloudFlare
	//   "X-Appengine-City" for App Engine (Standard)
	RequestIDHeader string `json:",omitempty"`

	// DisallowedHeaders is a list of headers filtered from the response. These either add no value
	// or leak information that we don't want displayed to the user.
	DisallowedHeaders []string `json:",omitempty"`

	// MapsAPIKey is used to render static Google Maps.
	// Request your own at https://developers.google.com/maps/documentation/static-maps/
	MapsAPIKey string `json:",omitempty"`

	// MapsAPISigningKey is a secret key that allows you to sign the map URL request
	// to prove we are the owning of the static map api key.
	MapsAPISigningKey []byte `json:",omitempty"`
}

// ApplyDefaults returns a new config with any zero field in config, set to the default value.
func ApplyDefaults(config, defaults *Config) (*Config, error) {
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
