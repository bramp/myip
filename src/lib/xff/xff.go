package xff

import "strings"

// parseXFF parses a x-forwarded-for header and returns all the addresses found
// The header has the form: client, proxy1, proxy2
func parseXFF(header string) []string {
	addrs := []string{}

	for _, addr := range strings.Split(header, ",") {
		addr := strings.TrimSpace(addr)
		if len(addr) == 0 {
			continue
		}
		addrs = append(addrs, addr)
	}

	return addrs
}
