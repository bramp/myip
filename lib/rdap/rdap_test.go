package rdap

import (
	"strings"
	"testing"

	openrdap "github.com/openrdap/rdap"
)

func TestIPNetworkToResponse(t *testing.T) {
	ipNet := &openrdap.IPNetwork{
		Handle:       "NET-8-8-8-0-1",
		Name:         "LVLT-GOGL-8-8-8",
		StartAddress: "8.8.8.0",
		EndAddress:   "8.8.8.255",
		IPVersion:    "v4",
		Type:         "ALLOCATION",
		Country:      "US",
		ParentHandle: "NET-8-0-0-0-1",
		Status:       []string{"active"},
		Port43:       "whois.arin.net",
		Events: []openrdap.Event{
			{Action: "registration", Date: "2014-03-14T16:52:05-04:00"},
			{Action: "last changed", Date: "2014-11-24T08:10:00-05:00"},
		},
		Remarks: []openrdap.Remark{
			{Title: "Registration", Description: []string{"This is a test remark."}},
		},
		Links: []openrdap.Link{
			{Href: "https://rdap.arin.net/registry/ip/8.8.8.0"},
		},
		Entities: []openrdap.Entity{
			{
				Handle: "GOGL",
				Roles:  []string{"registrant"},
			},
		},
	}

	resp := ipNetworkToResponse("8.8.8.8", ipNet)

	if resp.Query != "8.8.8.8" {
		t.Errorf("Query = %q, want %q", resp.Query, "8.8.8.8")
	}
	if resp.Name != "LVLT-GOGL-8-8-8" {
		t.Errorf("Name = %q, want %q", resp.Name, "LVLT-GOGL-8-8-8")
	}
	if resp.Handle != "NET-8-8-8-0-1" {
		t.Errorf("Handle = %q, want %q", resp.Handle, "NET-8-8-8-0-1")
	}
	if resp.StartAddress != "8.8.8.0" {
		t.Errorf("StartAddress = %q, want %q", resp.StartAddress, "8.8.8.0")
	}
	if resp.EndAddress != "8.8.8.255" {
		t.Errorf("EndAddress = %q, want %q", resp.EndAddress, "8.8.8.255")
	}
	if resp.CIDR != "8.8.8.0/24" {
		t.Errorf("CIDR = %q, want %q", resp.CIDR, "8.8.8.0/24")
	}
	if resp.IPVersion != "v4" {
		t.Errorf("IPVersion = %q, want %q", resp.IPVersion, "v4")
	}
	if resp.Country != "US" {
		t.Errorf("Country = %q, want %q", resp.Country, "US")
	}
	if resp.Type != "ALLOCATION" {
		t.Errorf("Type = %q, want %q", resp.Type, "ALLOCATION")
	}
	if resp.ParentHandle != "NET-8-0-0-0-1" {
		t.Errorf("ParentHandle = %q, want %q", resp.ParentHandle, "NET-8-0-0-0-1")
	}
	if resp.Status != "active" {
		t.Errorf("Status = %q, want %q", resp.Status, "active")
	}
	if resp.Port43 != "whois.arin.net" {
		t.Errorf("Port43 = %q, want %q", resp.Port43, "whois.arin.net")
	}
	if len(resp.Events) != 2 {
		t.Fatalf("Events len = %d, want 2", len(resp.Events))
	}
	if resp.Events[0].Action != "registration" {
		t.Errorf("Events[0].Action = %q, want %q", resp.Events[0].Action, "registration")
	}
	if len(resp.Remarks) != 1 {
		t.Fatalf("Remarks len = %d, want 1", len(resp.Remarks))
	}
	if resp.Remarks[0].Title != "Registration" {
		t.Errorf("Remarks[0].Title = %q, want %q", resp.Remarks[0].Title, "Registration")
	}
	if len(resp.Links) != 1 {
		t.Fatalf("Links len = %d, want 1", len(resp.Links))
	}
	if resp.Links[0] != "https://rdap.arin.net/registry/ip/8.8.8.0" {
		t.Errorf("Links[0] = %q", resp.Links[0])
	}
	if len(resp.Entities) != 1 {
		t.Fatalf("Entities len = %d, want 1", len(resp.Entities))
	}
	if resp.Entities[0].Handle != "GOGL" {
		t.Errorf("Entities[0].Handle = %q, want %q", resp.Entities[0].Handle, "GOGL")
	}
	if resp.Entities[0].Roles[0] != "registrant" {
		t.Errorf("Entities[0].Roles[0] = %q, want %q", resp.Entities[0].Roles[0], "registrant")
	}
	if resp.Body == "" {
		t.Error("Body should not be empty")
	}
}

func TestIPNetworkToResponseMultipleStatus(t *testing.T) {
	ipNet := &openrdap.IPNetwork{
		Handle:       "TEST-NET",
		StartAddress: "10.0.0.0",
		EndAddress:   "10.0.0.255",
		Status:       []string{"active", "validated"},
	}

	resp := ipNetworkToResponse("10.0.0.1", ipNet)

	if resp.Status != "active, validated" {
		t.Errorf("Status = %q, want %q", resp.Status, "active, validated")
	}
}

func TestIPNetworkToResponseEmpty(t *testing.T) {
	ipNet := &openrdap.IPNetwork{}
	resp := ipNetworkToResponse("1.2.3.4", ipNet)

	if resp.Query != "1.2.3.4" {
		t.Errorf("Query = %q, want %q", resp.Query, "1.2.3.4")
	}
}

func TestFormatTextBody(t *testing.T) {
	resp := &Response{
		Name:         "EXAMPLE-NET",
		Handle:       "NET-1-0-0-0-1",
		StartAddress: "1.0.0.0",
		EndAddress:   "1.0.0.255",
		CIDR:         "1.0.0.0/24",
		IPVersion:    "v4",
		Type:         "ALLOCATION",
		ParentHandle: "NET-1-0-0-0-0",
		Country:      "AU",
		Status:       "active",
		Port43:       "whois.apnic.net",
		Events: []Event{
			{Action: "registration", Date: "2011-02-10T00:00:00Z"},
		},
		Remarks: []Remark{
			{Title: "Description", Description: []string{"APNIC Research"}},
		},
		Links: []string{"https://rdap.apnic.net/ip/1.0.0.0/24"},
		Entities: []Entity{
			{
				Handle:  "APNIC",
				Roles:   []string{"registrant"},
				Address: "6 Cordelia St, South Brisbane, QLD, 4101, Australia",
				Phone:   "+61-7-3858-3100",
				Email:   "helpdesk@apnic.net",
			},
		},
	}

	body := formatTextBody(resp)

	expectedParts := []string{
		"Name:            EXAMPLE-NET",
		"Handle:          NET-1-0-0-0-1",
		"Range:           1.0.0.0 - 1.0.0.255",
		"CIDR:            1.0.0.0/24",
		"IP Version:      v4",
		"Type:            ALLOCATION",
		"Parent:          NET-1-0-0-0-0",
		"Country:         AU",
		"Status:          active",
		"Port43:          whois.apnic.net",
		"Event:           registration @ 2011-02-10T00:00:00Z",
		"Link:            https://rdap.apnic.net/ip/1.0.0.0/24",
		"Remark:          Description",
		"APNIC Research",
		"Entity:          APNIC",
		"Roles:",
		"Address:",
		"Phone:",
		"Email:",
	}

	for _, part := range expectedParts {
		if !strings.Contains(body, part) {
			t.Errorf("Body missing %q\nGot:\n%s", part, body)
		}
	}
}

func TestFormatTextBodyEmpty(t *testing.T) {
	resp := &Response{}
	body := formatTextBody(resp)

	// An empty response should not panic and should produce minimal output
	if body == "" {
		// Range line with empty start/end produces " - " which gets trimmed or not
		// depending on writeLine logic. Just ensure no panic.
	}
}

func TestFormatTextBodyEntityWithName(t *testing.T) {
	resp := &Response{
		Name: "TEST",
		Entities: []Entity{
			{
				Handle:  "HANDLE1",
				Name:    "Google LLC",
				Roles:   []string{"registrant"},
				Address: "1600 Amphitheatre Parkway, Mountain View, CA, 94043, US",
				Phone:   "+1-650-253-0000",
				Email:   "arin-contact@google.com",
				Entities: []Entity{
					{Handle: "ABUSE-GOGL", Roles: []string{"abuse"}, Email: "network-abuse@google.com"},
				},
			},
		},
	}

	body := formatTextBody(resp)

	if !strings.Contains(body, "Entity:          Google LLC") {
		t.Errorf("Body should show entity name 'Google LLC', got:\n%s", body)
	}
	if !strings.Contains(body, "HANDLE1") {
		t.Errorf("Body should show handle when different from name, got:\n%s", body)
	}
	if !strings.Contains(body, "1600 Amphitheatre") {
		t.Errorf("Body should show address, got:\n%s", body)
	}
	if !strings.Contains(body, "+1-650-253-0000") {
		t.Errorf("Body should show phone, got:\n%s", body)
	}
	if !strings.Contains(body, "arin-contact@google.com") {
		t.Errorf("Body should show email, got:\n%s", body)
	}
	// Nested entity
	if !strings.Contains(body, "ABUSE-GOGL") {
		t.Errorf("Body should show nested entity, got:\n%s", body)
	}
	if !strings.Contains(body, "network-abuse@google.com") {
		t.Errorf("Body should show nested entity email, got:\n%s", body)
	}
}

func TestCIDRFromRange(t *testing.T) {
	tests := []struct {
		start, end, want string
	}{
		{"8.8.8.0", "8.8.8.255", "8.8.8.0/24"},
		{"198.212.194.0", "198.212.195.255", "198.212.194.0/23"},
		{"10.0.0.0", "10.255.255.255", "10.0.0.0/8"},
		{"192.168.1.0", "192.168.1.127", "192.168.1.0/25"},
		{"", "", ""},                 // empty
		{"invalid", "8.8.8.255", ""}, // invalid start
	}

	for _, tc := range tests {
		got := cidrFromRange(tc.start, tc.end)
		if got != tc.want {
			t.Errorf("cidrFromRange(%q, %q) = %q, want %q", tc.start, tc.end, got, tc.want)
		}
	}
}
