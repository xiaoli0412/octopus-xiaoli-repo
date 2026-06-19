package op

import (
	"net/http"
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

func TestClientIPFromRequestOnlyTrustsConfiguredProxy(t *testing.T) {
	SetupOpTestDB(t)

	req := &http.Request{
		RemoteAddr: "10.0.0.2:4567",
		Host:       "internal.local:1088",
		Header: http.Header{
			"X-Forwarded-For":   []string{"203.0.113.42, 10.0.0.2"},
			"X-Forwarded-Proto": []string{"https"},
			"X-Forwarded-Host":  []string{"api.example.com"},
		},
	}

	if got := ClientIPFromRequest(req); got != "10.0.0.2" {
		t.Fatalf("ClientIPFromRequest() = %q, want untrusted remote IP", got)
	}
	if got := CurrentRequestBaseURL(req); got != "http://internal.local:1088" {
		t.Fatalf("CurrentRequestBaseURL() = %q, want direct request host", got)
	}

	if err := SettingSetString(model.SettingKeyTrustedProxyCIDRs, "10.0.0.0/8"); err != nil {
		t.Fatalf("SettingSetString(trusted_proxy_cidrs) error = %v", err)
	}

	if got := ClientIPFromRequest(req); got != "203.0.113.42" {
		t.Fatalf("ClientIPFromRequest() = %q, want forwarded client IP", got)
	}
	if got := CurrentRequestBaseURL(req); got != "https://api.example.com" {
		t.Fatalf("CurrentRequestBaseURL() = %q, want trusted forwarded base URL", got)
	}
}

func TestMaskClientIP(t *testing.T) {
	if got := MaskClientIP("203.0.113.42"); got != "203.0.113.*" {
		t.Fatalf("MaskClientIP(v4) = %q", got)
	}
	if got := MaskClientIP("2001:db8:abcd::1"); got != "2001:db8:*" {
		t.Fatalf("MaskClientIP(v6) = %q", got)
	}
}
