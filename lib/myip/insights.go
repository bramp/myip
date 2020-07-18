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
	"net"
	"net/http"
	"strings"
)

func addressFamily(addr string) string {
	ip := net.ParseIP(addr)
	if len(ip) == net.IPv4len || ip.To4() != nil {
		return "IPv4"
	}
	if len(ip) == net.IPv6len {
		return "IPv6"
	}
	return "Unknown"
}

func addInsights(req *http.Request, resp *Response) *Response {
	resp.Insights = make(map[string]string)

	if s := req.Header.Get("Via"); strings.Contains(s, "Chrome-Compression-Proxy") {
		// https://developer.chrome.com/multidevice/data-compression
		resp.Insights["Proxy"] = "Chrome Compression Proxy"
	}

	if actual := addressFamily(resp.RemoteAddr); req.URL.Query().Get("family") != actual {
		resp.Insights["AddressMismatch"] = actual
	}

	return resp
}
