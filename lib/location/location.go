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

package location

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"bramp.net/myip/lib/conf"
)

// Response contains the location data we send to the user.
type Response struct {
	City    string `json:",omitempty"`
	Region  string `json:",omitempty"`
	Country string `json:",omitempty"`

	Lat, Long float64 `json:",omitempty"`

	// MapURL a URL to Google Maps Static map.
	MapURL string `json:",omitempty"`
}

func parseLatLong(latlong string) (float64, float64, error) {
	parts := strings.SplitN(latlong, ",", 2)
	if len(parts) != 2 {
		return 0, 0, errors.New("latlong is not seperated by comma")
	}
	lat, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, 0, err
	}
	long, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, 0, err
	}

	return lat, long, nil
}

type googleMapURLBuilder struct {
	Key    string
	Secret []byte
}

func (b *googleMapURLBuilder) build(r *Response) string {
	base := b.baseURL(r)

	if base == "" {
		// If we couldn't build the URL just bail
		return ""
	}

	if b.Key != "" {
		base += "&key=" + b.Key
	}

	if signature := b.signature(base); signature != "" {
		base += "&signature=" + signature
	}

	return "https://maps.googleapis.com" + base
}

func (b *googleMapURLBuilder) baseURL(r *Response) string {
	base := "/maps/api/staticmap"
	base += "?size=640x400"
	base += "&markers=color:red%7C" // markerStyles %7C location

	if r.Lat != 0 && r.Long != 0 {
		return base + fmt.Sprintf("%f,%f", r.Lat, r.Long)
	}

	if r.City != "" {
		return base + url.QueryEscape(r.City)
	}

	if r.Region != "" {
		return base + url.QueryEscape(r.Region)
	}

	if r.Country != "" {
		return base + url.QueryEscape(r.Country)
	}

	return ""
}

func (b *googleMapURLBuilder) signature(url string) string {
	if b.Secret == nil {
		return ""
	}

	mac := hmac.New(sha1.New, b.Secret)
	mac.Write([]byte(url))

	return base64.URLEncoding.EncodeToString(mac.Sum(nil))
}

// Handle generates a location.Response
func Handle(config *conf.Config, req *http.Request) *Response {
	lat, long, _ := parseLatLong(req.Header.Get(config.LatLongHeader))
	response := &Response{
		City:    req.Header.Get(config.CityHeader),
		Region:  req.Header.Get(config.RegionHeader),
		Country: req.Header.Get(config.CountryHeader),
		Lat:     lat,
		Long:    long,
	}

	// Optionally add a mapUrl
	if config.MapsAPIKey != "" {
		b := &googleMapURLBuilder{
			Key:    config.MapsAPIKey,
			Secret: config.MapsAPISigningKey,
		}

		response.MapURL = b.build(response)
	}

	return response
}
