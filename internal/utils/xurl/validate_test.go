package xurl

import "testing"

func TestValidateAbsoluteHTTPURL(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		fieldName string
		wantErr   bool
	}{
		{name: "http url", raw: "http://example.com/v1", fieldName: "base_url"},
		{name: "https url with spaces", raw: "  https://example.com/v1  ", fieldName: "base_url"},
		{name: "empty", raw: "", fieldName: "base_url", wantErr: true},
		{name: "relative", raw: "/v1", fieldName: "base_url", wantErr: true},
		{name: "unsupported scheme", raw: "ftp://example.com/v1", fieldName: "base_url", wantErr: true},
		{name: "hostless", raw: "http:///v1", fieldName: "base_url", wantErr: true},
		{name: "embedded credentials", raw: "https://user:pass@example.com/v1", fieldName: "base_url", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateAbsoluteHTTPURL(tc.raw, tc.fieldName)
			if tc.wantErr && err == nil {
				t.Fatalf("ValidateAbsoluteHTTPURL(%q) expected error", tc.raw)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("ValidateAbsoluteHTTPURL(%q) error = %v", tc.raw, err)
			}
		})
	}
}

func TestValidateProxyURL(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		fieldName string
		wantErr   bool
	}{
		{name: "http proxy", raw: "http://127.0.0.1:8080", fieldName: "channel_proxy"},
		{name: "socks5 proxy", raw: "socks5://127.0.0.1:1080", fieldName: "channel_proxy"},
		{name: "empty", raw: "", fieldName: "channel_proxy", wantErr: true},
		{name: "relative", raw: "/proxy", fieldName: "channel_proxy", wantErr: true},
		{name: "unsupported scheme", raw: "ftp://127.0.0.1:21", fieldName: "channel_proxy", wantErr: true},
		{name: "hostless", raw: "http:///proxy", fieldName: "channel_proxy", wantErr: true},
		{name: "embedded credentials", raw: "https://user:pass@example.com:8443", fieldName: "channel_proxy", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateProxyURL(tc.raw, tc.fieldName)
			if tc.wantErr && err == nil {
				t.Fatalf("ValidateProxyURL(%q) expected error", tc.raw)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("ValidateProxyURL(%q) error = %v", tc.raw, err)
			}
		})
	}
}
