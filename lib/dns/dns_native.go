// +build !appengine

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
	"context"
	"net"
)

// LookupAddr performs a reverse lookup for the given address, returning a list
// of names mapping to that address.
func LookupAddr(ctx context.Context, ipAddr string) ([]string, error) {

	// Special case localhost
	// TODO Dedup this with dns_appengine.go
	if ip := net.ParseIP(ipAddr); ip.IsLoopback() {
		if ip.To4() != nil {
			return []string{"localhost"}, nil
		}
		return []string{"ip6-localhost"}, nil
	}

	return net.LookupAddr(ipAddr)
}
