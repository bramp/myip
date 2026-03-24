package myip

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"bramp.net/myip/lib/dns"
	"bramp.net/myip/lib/location"
	"bramp.net/myip/lib/rdap"
	"bramp.net/myip/lib/whois"
)

// fullResponse returns a Response with both RDAP and WHOIS populated.
func fullResponse() *Response {
	return &Response{
		RemoteAddr:       "198.212.195.91",
		RemoteAddrFamily: "IPv4",
		RemoteAddrReverse: &dns.Response{
			Query: "198.212.195.91",
			Names: []string{"esn-195-91.espacenetworks.io."},
		},
		RemoteAddrRDAP: &rdap.Response{
			Query:        "198.212.195.91",
			Name:         "EN-139",
			Handle:       "NET-198-212-194-0-1",
			StartAddress: "198.212.194.0",
			EndAddress:   "198.212.195.255",
			CIDR:         "198.212.194.0/23",
			IPVersion:    "v4",
			Type:         "DIRECT ALLOCATION",
			ParentHandle: "NET-198-0-0-0-0",
			Status:       "active",
			Port43:       "whois.arin.net",
			Events: []rdap.Event{
				{Action: "registration", Date: "2024-06-21T14:33:59-04:00"},
				{Action: "last changed", Date: "2025-10-28T12:12:07-04:00"},
			},
			Remarks: []rdap.Remark{
				{Title: "Registration Comments", Description: []string{"Geofeed https://www.quvia.ai/geofeed"}},
			},
			Links: []string{"https://rdap.arin.net/registry/ip/198.212.194.0"},
			Entities: []rdap.Entity{
				{
					Handle:  "EN-139",
					Name:    "ESpace Networks",
					Roles:   []string{"registrant"},
					Address: "3100 South West 145TH Avenue Suite 310, Miramar, FL, 33027, US",
					Phone:   "+1-786-340-5177",
					Email:   "admin@espacenetworks.io",
					Entities: []rdap.Entity{
						{Handle: "ABUSE9298-ARIN", Name: "Abuse", Roles: []string{"abuse"}, Email: "network-abuse@quvia.ai"},
					},
				},
			},
			Body: "Name:            EN-139\nHandle:          NET-198-212-194-0-1\nRange:           198.212.194.0 - 198.212.195.255\nCIDR:            198.212.194.0/23\nIP Version:      v4\nType:            DIRECT ALLOCATION\nStatus:          active",
		},
		RemoteAddrWhois: &whois.Response{
			Query: "198.212.195.91",
			Body:  "NetRange:       198.212.194.0 - 198.212.195.255\nNetName:        EN-139\nOrgName:        ESpace Networks",
		},
		Location: &location.Response{
			City:    "Miramar",
			Region:  "FL",
			Country: "US",
		},
		Method: "GET",
		URL:    "http://localhost:8080/?host=198.212.195.91",
		Proto:  "HTTP/1.1",
		Header: http.Header{
			"User-Agent": []string{"curl/8.0"},
		},
	}
}

func TestCLITemplateFullResponse(t *testing.T) {
	resp := fullResponse()

	var buf bytes.Buffer
	if err := cliTmpl.Execute(&buf, resp); err != nil {
		t.Fatalf("CLI template execution failed: %v", err)
	}

	body := buf.String()

	for _, want := range []string{
		"IP: 198.212.195.91",
		"DNS: esn-195-91.espacenetworks.io.",
		"RDAP:",
		"EN-139",
		"NET-198-212-194-0-1",
		"198.212.194.0/23",
		"WHOIS:",
		"NetRange:",
		"ESpace Networks",
		"Location:",
		"Miramar",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("CLI output missing %q\nGot:\n%s", want, body)
		}
	}
}

func TestCLITemplateRDAPOnly(t *testing.T) {
	resp := fullResponse()
	resp.RemoteAddrWhois = nil

	var buf bytes.Buffer
	if err := cliTmpl.Execute(&buf, resp); err != nil {
		t.Fatalf("CLI template execution failed with nil WHOIS: %v", err)
	}

	body := buf.String()

	if strings.Contains(body, "WHOIS:") {
		t.Error("CLI output should not contain WHOIS section when RemoteAddrWhois is nil")
	}
	if !strings.Contains(body, "RDAP:") {
		t.Error("CLI output should contain RDAP section")
	}
	if !strings.Contains(body, "EN-139") {
		t.Error("CLI output should contain RDAP body content")
	}
}

func TestCLITemplateWhoisOnly(t *testing.T) {
	resp := fullResponse()
	resp.RemoteAddrRDAP = &rdap.Response{
		Query: "198.212.195.91",
		Error: "no RDAP servers found",
	}

	var buf bytes.Buffer
	if err := cliTmpl.Execute(&buf, resp); err != nil {
		t.Fatalf("CLI template execution failed with RDAP error: %v", err)
	}

	body := buf.String()

	if !strings.Contains(body, "WHOIS:") {
		t.Error("CLI output should contain WHOIS section")
	}
	if !strings.Contains(body, "NetRange:") {
		t.Error("CLI output should contain WHOIS body content")
	}
}

func TestCLITemplateNilRDAP(t *testing.T) {
	resp := fullResponse()
	resp.RemoteAddrRDAP = nil

	var buf bytes.Buffer
	if err := cliTmpl.Execute(&buf, resp); err != nil {
		t.Fatalf("CLI template execution failed with nil RDAP: %v", err)
	}

	body := buf.String()

	if strings.Contains(body, "RDAP:") {
		t.Error("CLI output should not contain RDAP section when RemoteAddrRDAP is nil")
	}
	if !strings.Contains(body, "WHOIS:") {
		t.Error("CLI output should still contain WHOIS section")
	}
}

func TestCLITemplateNilBothLookups(t *testing.T) {
	resp := &Response{
		RemoteAddr:        "127.0.0.1",
		RemoteAddrFamily:  "IPv4",
		RemoteAddrReverse: &dns.Response{Query: "127.0.0.1"},
		RemoteAddrRDAP:    &rdap.Response{Query: "127.0.0.1", Error: "no RDAP servers found"},
		Location:          &location.Response{},
		Method:            "GET",
		URL:               "http://localhost:8080/",
		Proto:             "HTTP/1.1",
		Header:            http.Header{},
	}

	var buf bytes.Buffer
	if err := cliTmpl.Execute(&buf, resp); err != nil {
		t.Fatalf("CLI template execution failed for localhost: %v", err)
	}

	body := buf.String()

	if !strings.Contains(body, "IP: 127.0.0.1") {
		t.Errorf("CLI output should contain IP, got:\n%s", body)
	}
}

// JSON output tests

func TestJSONOutputFullResponse(t *testing.T) {
	resp := fullResponse()

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	// Verify RDAP fields
	rdapData, ok := decoded["RemoteAddrRDAP"].(map[string]interface{})
	if !ok {
		t.Fatal("RemoteAddrRDAP missing or not an object")
	}
	for _, field := range []string{"Query", "Name", "Handle", "StartAddress", "EndAddress", "CIDR", "IPVersion", "Type", "ParentHandle", "Status", "Port43", "Events", "Remarks", "Links", "Entities", "Body"} {
		if _, ok := rdapData[field]; !ok {
			t.Errorf("RemoteAddrRDAP missing field %q", field)
		}
	}

	// Verify entities have contact details
	entities, ok := rdapData["Entities"].([]interface{})
	if !ok || len(entities) == 0 {
		t.Fatal("RemoteAddrRDAP.Entities missing or empty")
	}
	entity := entities[0].(map[string]interface{})
	for _, field := range []string{"Handle", "Name", "Roles", "Address", "Phone", "Email", "Entities"} {
		if _, ok := entity[field]; !ok {
			t.Errorf("RDAP Entity missing field %q", field)
		}
	}

	// Verify nested entity
	nestedEntities, ok := entity["Entities"].([]interface{})
	if !ok || len(nestedEntities) == 0 {
		t.Fatal("Nested entities missing")
	}
	nestedEntity := nestedEntities[0].(map[string]interface{})
	if nestedEntity["Handle"] != "ABUSE9298-ARIN" {
		t.Errorf("Nested entity Handle = %v, want ABUSE9298-ARIN", nestedEntity["Handle"])
	}

	// Verify WHOIS fields
	whoisData, ok := decoded["RemoteAddrWhois"].(map[string]interface{})
	if !ok {
		t.Fatal("RemoteAddrWhois missing or not an object")
	}
	if whoisData["Body"] == nil || whoisData["Body"] == "" {
		t.Error("RemoteAddrWhois.Body should not be empty")
	}
}

func TestJSONOutputRDAPOnly(t *testing.T) {
	resp := fullResponse()
	resp.RemoteAddrWhois = nil

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if _, ok := decoded["RemoteAddrRDAP"]; !ok {
		t.Error("RemoteAddrRDAP should be present")
	}
	if _, ok := decoded["RemoteAddrWhois"]; ok {
		t.Error("RemoteAddrWhois should be omitted when nil")
	}
}

func TestJSONOutputWhoisOnly(t *testing.T) {
	resp := fullResponse()
	resp.RemoteAddrRDAP = nil

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if _, ok := decoded["RemoteAddrRDAP"]; ok {
		t.Error("RemoteAddrRDAP should be omitted when nil")
	}
	if _, ok := decoded["RemoteAddrWhois"]; !ok {
		t.Error("RemoteAddrWhois should be present")
	}
}

func TestJSONOutputRDAPEmptyBody(t *testing.T) {
	resp := fullResponse()
	resp.RemoteAddrRDAP = &rdap.Response{
		Query: "198.212.195.91",
		Error: "no RDAP servers found",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	rdapData := decoded["RemoteAddrRDAP"].(map[string]interface{})
	if rdapData["Error"] != "no RDAP servers found" {
		t.Errorf("RDAP Error = %v, want 'no RDAP servers found'", rdapData["Error"])
	}
}

func TestJSONOutputRDAPEventsAndRemarks(t *testing.T) {
	resp := fullResponse()

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	var decoded map[string]interface{}
	json.Unmarshal(data, &decoded)

	rdapData := decoded["RemoteAddrRDAP"].(map[string]interface{})

	// Verify events
	events := rdapData["Events"].([]interface{})
	if len(events) != 2 {
		t.Fatalf("Expected 2 events, got %d", len(events))
	}
	event0 := events[0].(map[string]interface{})
	if event0["Action"] != "registration" {
		t.Errorf("Event[0].Action = %v, want 'registration'", event0["Action"])
	}
	if event0["Date"] != "2024-06-21T14:33:59-04:00" {
		t.Errorf("Event[0].Date = %v", event0["Date"])
	}

	// Verify remarks
	remarks := rdapData["Remarks"].([]interface{})
	if len(remarks) != 1 {
		t.Fatalf("Expected 1 remark, got %d", len(remarks))
	}
	remark := remarks[0].(map[string]interface{})
	if remark["Title"] != "Registration Comments" {
		t.Errorf("Remark.Title = %v", remark["Title"])
	}

	// Verify links
	links := rdapData["Links"].([]interface{})
	if len(links) != 1 {
		t.Fatalf("Expected 1 link, got %d", len(links))
	}
	if links[0] != "https://rdap.arin.net/registry/ip/198.212.194.0" {
		t.Errorf("Link = %v", links[0])
	}
}
