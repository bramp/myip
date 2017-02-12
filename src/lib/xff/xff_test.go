package xff

import (
	"reflect"
	"testing"
)

func TestParseXFF(t *testing.T) {
	data := []struct {
		header string
		want   []string
	}{
		{header: "2601:646:c200:b466:0:0:0:1, 2607:f8b0:4005:801::2014", want: []string{"2601:646:c200:b466:0:0:0:1", "2607:f8b0:4005:801::2014"}},
		{header: "1.2.3.4", want: []string{"1.2.3.4"}},
		{header: "1.2.3.4, 8.8.8.8", want: []string{"1.2.3.4", "8.8.8.8"}},
		{header: "1.2.3.4, 192.168.0.1", want: []string{"1.2.3.4", "192.168.0.1"}},
		{header: "", want: []string{}},
	}

	for _, test := range data {
		got := parseXFF(test.header)
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("parseXFF(%q) = %q want %q", test.header, got, test.want)
		}
	}
}
