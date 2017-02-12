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

package main

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
)

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

func handleLocation(req *http.Request) *locationResponse {
	lat, long, _ := parseLatLong(req.Header.Get(config.LatLongHeader))
	locationResponse := &locationResponse{
		City:    req.Header.Get(config.CityHeader),
		Region:  req.Header.Get(config.RegionHeader),
		Country: req.Header.Get(config.CountryHeader),
		Lat:     lat,
		Long:    long,
	}

	locationResponse.City = "san mateo"

	return locationResponse
}
