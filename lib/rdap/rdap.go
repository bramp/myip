package rdap

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	openrdap "github.com/openrdap/rdap"
	log "github.com/sirupsen/logrus"
)

const (
	// RDAPTimeout is the timeout for RDAP requests.
	RDAPTimeout = 10 * time.Second
)

// Response contains the RDAP data we send to the user.
type Response struct {
	Query string

	// Network fields
	Name         string `json:",omitempty"`
	Handle       string `json:",omitempty"`
	StartAddress string `json:",omitempty"`
	EndAddress   string `json:",omitempty"`
	CIDR         string `json:",omitempty"`
	IPVersion    string `json:",omitempty"`
	Country      string `json:",omitempty"`
	Type         string `json:",omitempty"`
	ParentHandle string `json:",omitempty"`
	Status       string `json:",omitempty"`
	Port43       string `json:",omitempty"`

	// Events (registration, last changed, etc.)
	Events []Event `json:",omitempty"`

	// Remarks from the registry
	Remarks []Remark `json:",omitempty"`

	// Links to related resources
	Links []string `json:",omitempty"`

	Entities []Entity `json:",omitempty"`

	// Body is a human-readable text rendering of the RDAP data.
	Body string `json:",omitempty"`

	Error string `json:",omitempty"`
}

// Event represents a dated event (e.g. registration, last changed).
type Event struct {
	Action string `json:",omitempty"`
	Date   string `json:",omitempty"`
}

// Remark contains a titled description from the registry.
type Remark struct {
	Title       string   `json:",omitempty"`
	Description []string `json:",omitempty"`
}

// Entity contains information about an organization or contact.
type Entity struct {
	Handle string   `json:",omitempty"`
	Name   string   `json:",omitempty"`
	Roles  []string `json:",omitempty"`

	// VCard contact details
	Address string `json:",omitempty"`
	Phone   string `json:",omitempty"`
	Fax     string `json:",omitempty"`
	Email   string `json:",omitempty"`

	// Nested entities (e.g. abuse contact under an org)
	Entities []Entity `json:",omitempty"`

	// Events on this entity
	Events []Event `json:",omitempty"`
}

// Client wraps the openrdap client and provides methods for querying IP addresses.
type Client struct {
	client *openrdap.Client
}

// NewClient creates a new RDAP client with the given timeout.
func NewClient(timeout time.Duration) *Client {
	return &Client{
		client: &openrdap.Client{
			HTTP: &http.Client{
				Timeout: timeout,
			},
		},
	}
}

// QueryIP performs an RDAP lookup for the given IP address and returns a Response.
func (c *Client) QueryIP(ctx context.Context, ipAddr string) *Response {
	req := &openrdap.Request{
		Type:    openrdap.IPRequest,
		Query:   ipAddr,
		Timeout: RDAPTimeout,
	}
	req = req.WithContext(ctx)

	log.Infof("RDAP request for %q", ipAddr)

	resp, err := c.client.Do(req)
	if err != nil {
		log.Warningf("RDAP failed for %q: %s", ipAddr, err)
		return &Response{
			Query: ipAddr,
			Error: err.Error(),
		}
	}

	ipNet, ok := resp.Object.(*openrdap.IPNetwork)
	if !ok {
		log.Warningf("RDAP returned non-IPNetwork response for %q", ipAddr)
		return &Response{
			Query: ipAddr,
			Error: "unexpected RDAP response type",
		}
	}

	return ipNetworkToResponse(ipAddr, ipNet)
}

// ipNetworkToResponse converts an openrdap.IPNetwork to our Response type.
func ipNetworkToResponse(ipAddr string, ipNet *openrdap.IPNetwork) *Response {
	entities := convertEntities(ipNet.Entities)
	events := convertEvents(ipNet.Events)
	remarks := convertRemarks(ipNet.Remarks)

	links := make([]string, 0, len(ipNet.Links))
	for _, l := range ipNet.Links {
		if l.Href != "" {
			links = append(links, l.Href)
		}
	}

	resp := &Response{
		Query:        ipAddr,
		Name:         ipNet.Name,
		Handle:       ipNet.Handle,
		StartAddress: ipNet.StartAddress,
		EndAddress:   ipNet.EndAddress,
		CIDR:         cidrFromRange(ipNet.StartAddress, ipNet.EndAddress),
		IPVersion:    ipNet.IPVersion,
		Country:      ipNet.Country,
		Type:         ipNet.Type,
		ParentHandle: ipNet.ParentHandle,
		Status:       strings.Join(ipNet.Status, ", "),
		Port43:       ipNet.Port43,
		Events:       events,
		Remarks:      remarks,
		Links:        links,
		Entities:     entities,
	}

	resp.Body = formatTextBody(resp)
	return resp
}

// convertEntities recursively converts openrdap entities to our Entity type.
func convertEntities(src []openrdap.Entity) []Entity {
	entities := make([]Entity, 0, len(src))
	for _, e := range src {
		entity := Entity{
			Handle:   e.Handle,
			Roles:    e.Roles,
			Entities: convertEntities(e.Entities),
			Events:   convertEvents(e.Events),
		}

		if e.VCard != nil {
			entity.Name = e.VCard.Name()
			entity.Phone = e.VCard.Tel()
			entity.Fax = e.VCard.Fax()
			entity.Email = e.VCard.Email()
			entity.Address = buildAddress(e.VCard)
		}

		entities = append(entities, entity)
	}
	return entities
}

// convertEvents converts openrdap events to our Event type.
func convertEvents(src []openrdap.Event) []Event {
	if len(src) == 0 {
		return nil
	}
	events := make([]Event, 0, len(src))
	for _, e := range src {
		events = append(events, Event{
			Action: e.Action,
			Date:   e.Date,
		})
	}
	return events
}

// convertRemarks converts openrdap remarks to our Remark type.
func convertRemarks(src []openrdap.Remark) []Remark {
	if len(src) == 0 {
		return nil
	}
	remarks := make([]Remark, 0, len(src))
	for _, r := range src {
		remarks = append(remarks, Remark{
			Title:       r.Title,
			Description: r.Description,
		})
	}
	return remarks
}

// buildAddress constructs a single-line address from VCard fields.
func buildAddress(vc *openrdap.VCard) string {
	parts := []string{}
	for _, s := range []string{
		vc.StreetAddress(),
		vc.Locality(),
		vc.Region(),
		vc.PostalCode(),
		vc.Country(),
	} {
		if s != "" {
			parts = append(parts, s)
		}
	}
	return strings.Join(parts, ", ")
}

// cidrFromRange attempts to compute CIDR notation from start/end addresses.
// Returns empty string if it cannot be computed.
func cidrFromRange(start, end string) string {
	startIP := net.ParseIP(start)
	endIP := net.ParseIP(end)
	if startIP == nil || endIP == nil {
		return ""
	}

	// Use 4-byte form for IPv4, 16-byte for IPv6.
	startBytes := startIP.To4()
	endBytes := endIP.To4()
	if startBytes == nil || endBytes == nil {
		startBytes = startIP.To16()
		endBytes = endIP.To16()
	}
	if startBytes == nil || endBytes == nil || len(startBytes) != len(endBytes) {
		return ""
	}

	// XOR start and end to find the differing bits.
	diff := make([]byte, len(startBytes))
	for i := range startBytes {
		diff[i] = startBytes[i] ^ endBytes[i]
	}

	// Count leading zeros in the XOR result to find the prefix length.
	prefixLen := 0
	for _, b := range diff {
		if b == 0 {
			prefixLen += 8
			continue
		}
		for bit := 7; bit >= 0; bit-- {
			if b&(1<<uint(bit)) == 0 {
				prefixLen++
			} else {
				goto done
			}
		}
	}
done:

	// Verify: the host bits should all be 1 in end address.
	mask := net.CIDRMask(prefixLen, len(startBytes)*8)
	network := startIP.Mask(mask)
	if !network.Equal(startIP) {
		return "" // start is not a network address
	}

	return fmt.Sprintf("%s/%d", start, prefixLen)
}

// formatTextBody produces a human-readable text rendering of the RDAP response,
// similar to what a whois response looks like.
func formatTextBody(resp *Response) string {
	var b strings.Builder

	writeLine := func(key, value string) {
		if value != "" {
			fmt.Fprintf(&b, "%-16s %s\n", key+":", value)
		}
	}

	writeLine("Name", resp.Name)
	writeLine("Handle", resp.Handle)
	if resp.StartAddress != "" && resp.EndAddress != "" {
		writeLine("Range", resp.StartAddress+" - "+resp.EndAddress)
	}
	writeLine("CIDR", resp.CIDR)
	writeLine("IP Version", resp.IPVersion)
	writeLine("Type", resp.Type)
	writeLine("Parent", resp.ParentHandle)
	writeLine("Country", resp.Country)
	writeLine("Status", resp.Status)
	writeLine("Port43", resp.Port43)

	for _, ev := range resp.Events {
		writeLine("Event", ev.Action+" @ "+ev.Date)
	}

	for _, link := range resp.Links {
		writeLine("Link", link)
	}

	for _, r := range resp.Remarks {
		if r.Title != "" {
			b.WriteString("\n")
			writeLine("Remark", r.Title)
			for _, d := range r.Description {
				fmt.Fprintf(&b, "  %s\n", d)
			}
		}
	}

	for _, e := range resp.Entities {
		writeEntity(&b, &e, "")
	}

	return strings.TrimSpace(b.String())
}

// writeEntity writes a formatted entity block, recursing into nested entities.
func writeEntity(b *strings.Builder, e *Entity, indent string) {
	writeLine := func(key, value string) {
		if value != "" {
			fmt.Fprintf(b, "%s%-16s %s\n", indent, key+":", value)
		}
	}

	b.WriteString("\n")
	displayName := e.Name
	if displayName == "" {
		displayName = e.Handle
	}
	writeLine("Entity", displayName)
	if e.Handle != "" && e.Handle != displayName {
		writeLine("  Handle", e.Handle)
	}
	if len(e.Roles) > 0 {
		writeLine("  Roles", strings.Join(e.Roles, ", "))
	}
	writeLine("  Address", e.Address)
	writeLine("  Phone", e.Phone)
	writeLine("  Fax", e.Fax)
	writeLine("  Email", e.Email)

	for _, ev := range e.Events {
		writeLine("  Event", ev.Action+" @ "+ev.Date)
	}

	for _, nested := range e.Entities {
		writeEntity(b, &nested, indent+"  ")
	}
}

// Handle performs an RDAP lookup for the given IP address using a default client.
// This is the main entry point, matching the signature of the old whois.Handle.
func Handle(ctx context.Context, ipAddr string) *Response {
	client := NewClient(RDAPTimeout)
	return client.QueryIP(ctx, ipAddr)
}
