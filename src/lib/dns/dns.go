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

package dns

import (
	"time"
)

const (
	dnsTimeout = 4 * time.Second
)

// Response contains the DNS data we send to the user.
type Response struct {
	Query string

	// One of the following
	Names []string `json:",omitempty"`
	Error string   `json:",omitempty"`
}

// Google IPv4 and IPv6 DNS servers // TODO Make configurable
var dnsServers = []string{"8.8.8.8", "8.8.4.4", "2001:4860:4860::8888", "2001:4860:4860::8844"}
