package net

import (
	"reflect"
	"testing"
)

func TestParseURL(t *testing.T) {
	tests := []struct {
		rawURL string
		want   *ParsedURL
	}{
		{
			rawURL: "trojan://xxa.com:8080?param1=value1&param2=value2",
			want: &ParsedURL{
				Scheme: "trojan",
				Host:   "xxa.com",
				Port:   8080,
				URI:    "",
				Params: map[string][]string{"param1": {"value1"}, "param2": {"value2"}},
			},
		},
	}

	for _, test := range tests {
		got, err := ParseURL(test.rawURL)
		if err != nil {
			t.Errorf("ParseURL(%q) = %v", test.rawURL, err)
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("ParseURL(%q) = %v, want %v", test.rawURL, got, test.want)
		}
		t.Logf("ParseURL(%q) = %v", test.rawURL, got)
	}
}

func TestGetEthInterface(t *testing.T) {
	iface, err := GetEthInterface()
	if err != nil {
		t.Errorf("GetEthInterface() = %v", err)
	}
	t.Logf("GetEthInterface() = %v", iface)
}
