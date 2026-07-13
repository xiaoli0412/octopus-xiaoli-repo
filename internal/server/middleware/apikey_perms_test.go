package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

func TestCapabilityForPath(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		path string
		want string
	}{
		{name: "chat completions", path: "/v1/chat/completions", want: capabilityChat},
		{name: "embeddings", path: "/v1/embeddings", want: capabilityEmbedding},
		{name: "responses", path: "/v1/responses", want: capabilityResponse},
		{name: "messages", path: "/v1/messages", want: capabilityMessage},
		{name: "unknown path", path: "/v1/unknown", want: ""},
		{name: "management path", path: "/api/v1/channel/list", want: ""},
		{name: "root path", path: "/", want: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := capabilityForPath(tc.path)
			if got != tc.want {
				t.Fatalf("capabilityForPath(%q) = %q, want %q", tc.path, got, tc.want)
			}
		})
	}
}

func TestClientIPAllowed(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		clientIP string
		cidrs    []string
		want     bool
	}{
		// IPv4 CIDR
		{name: "ipv4 in range", clientIP: "10.0.1.5", cidrs: []string{"10.0.0.0/8"}, want: true},
		{name: "ipv4 out of range", clientIP: "11.0.0.1", cidrs: []string{"10.0.0.0/8"}, want: false},
		{name: "ipv4 exact subnet", clientIP: "192.168.1.100", cidrs: []string{"192.168.1.0/24"}, want: true},
		{name: "ipv4 boundary lower", clientIP: "192.168.1.0", cidrs: []string{"192.168.1.0/24"}, want: true},
		{name: "ipv4 boundary upper", clientIP: "192.168.1.255", cidrs: []string{"192.168.1.0/24"}, want: true},
		{name: "ipv4 just outside", clientIP: "192.168.2.0", cidrs: []string{"192.168.1.0/24"}, want: false},

		// IPv6 CIDR
		{name: "ipv6 in range", clientIP: "::1", cidrs: []string{"::1/128"}, want: true},
		{name: "ipv6 in subnet", clientIP: "2001:db8::1", cidrs: []string{"2001:db8::/32"}, want: true},
		{name: "ipv6 out of range", clientIP: "2001:db9::1", cidrs: []string{"2001:db8::/32"}, want: false},

		// Bare IP (no CIDR mask)
		{name: "bare ipv4 match", clientIP: "10.0.0.1", cidrs: []string{"10.0.0.1"}, want: true},
		{name: "bare ipv4 no match", clientIP: "10.0.0.2", cidrs: []string{"10.0.0.1"}, want: false},
		{name: "bare ipv6 match", clientIP: "::1", cidrs: []string{"::1"}, want: true},

		// Multiple CIDRs
		{name: "multi cidr first match", clientIP: "10.0.0.1", cidrs: []string{"10.0.0.0/8", "192.168.0.0/16"}, want: true},
		{name: "multi cidr second match", clientIP: "192.168.1.1", cidrs: []string{"10.0.0.0/8", "192.168.0.0/16"}, want: true},
		{name: "multi cidr no match", clientIP: "172.16.0.1", cidrs: []string{"10.0.0.0/8", "192.168.0.0/16"}, want: false},

		// Edge cases
		{name: "invalid client ip", clientIP: "not-an-ip", cidrs: []string{"10.0.0.0/8"}, want: false},
		{name: "invalid cidr skipped", clientIP: "10.0.0.1", cidrs: []string{"invalid-cidr", "10.0.0.0/8"}, want: true},
		{name: "empty cidr list", clientIP: "10.0.0.1", cidrs: []string{}, want: false},
		{name: "empty cidr entries skipped", clientIP: "10.0.0.1", cidrs: []string{"", "10.0.0.0/8"}, want: true},
		{name: "invalid bare ip cidr skipped", clientIP: "10.0.0.1", cidrs: []string{"not-an-ip", "10.0.0.0/8"}, want: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := clientIPAllowed(tc.clientIP, tc.cidrs)
			if got != tc.want {
				t.Fatalf("clientIPAllowed(%q, %v) = %v, want %v", tc.clientIP, tc.cidrs, got, tc.want)
			}
		})
	}
}

func TestEnforceAPIKeyPermissionsBackwardCompatible(t *testing.T) {
	t.Parallel()

	// 所有权限字段为空时应该通过（向后兼容）
	apiKey := model.APIKey{
		ID:     1,
		Name:   "test-backward-compat",
		APIKey: "sk-test-backward-compat",
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	if !enforceAPIKeyPermissions(c, apiKey) {
		t.Fatalf("enforceAPIKeyPermissions() returned false for empty permissions, expected true (backward compatible)")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	// 验证 context 中 AllowedChannels/AllowedGroups 为 nil（不限制）
	if ch := APIKeyAllowedChannelsFromContext(c); ch != nil {
		t.Fatalf("expected nil AllowedChannels, got %v", ch)
	}
	if g := APIKeyAllowedGroupsFromContext(c); g != nil {
		t.Fatalf("expected nil AllowedGroups, got %v", g)
	}
}

func TestEnforceAPIKeyPermissionsCapabilityAllowed(t *testing.T) {
	t.Parallel()

	apiKey := model.APIKey{
		ID:                  2,
		Name:                "test-cap-allowed",
		APIKey:              "sk-test-cap-allowed",
		AllowedCapabilities: `["chat","embedding"]`,
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	if !enforceAPIKeyPermissions(c, apiKey) {
		t.Fatalf("enforceAPIKeyPermissions() returned false for allowed capability, expected true")
	}
}

func TestEnforceAPIKeyPermissionsCapabilityDenied(t *testing.T) {
	t.Parallel()

	apiKey := model.APIKey{
		ID:                  3,
		Name:                "test-cap-denied",
		APIKey:              "sk-test-cap-denied",
		AllowedCapabilities: `["embedding"]`,
	}

	// 请求 chat 但只允许 embedding → 应拒绝
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	if enforceAPIKeyPermissions(c, apiKey) {
		t.Fatalf("enforceAPIKeyPermissions() returned true for denied capability, expected false")
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestEnforceAPIKeyPermissionsCapabilityEmbeddingAllowed(t *testing.T) {
	t.Parallel()

	apiKey := model.APIKey{
		ID:                  4,
		Name:                "test-emb-allowed",
		APIKey:              "sk-test-emb-allowed",
		AllowedCapabilities: `["embedding"]`,
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/embeddings", nil)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	if !enforceAPIKeyPermissions(c, apiKey) {
		t.Fatalf("enforceAPIKeyPermissions() returned false for allowed embedding, expected true")
	}
}

func TestEnforceAPIKeyPermissionsCapabilityUnknownPath(t *testing.T) {
	t.Parallel()

	// 配置了能力限制但请求路径无法推断能力 → 应拒绝
	apiKey := model.APIKey{
		ID:                  5,
		Name:                "test-unknown-path",
		APIKey:              "sk-test-unknown-path",
		AllowedCapabilities: `["chat"]`,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/channel/list", nil)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	if enforceAPIKeyPermissions(c, apiKey) {
		t.Fatalf("enforceAPIKeyPermissions() returned true for unknown path with capability restriction, expected false")
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", rec.Code)
	}
}

func TestEnforceAPIKeyPermissionsIPAllowed(t *testing.T) {
	t.Parallel()

	apiKey := model.APIKey{
		ID:             6,
		Name:           "test-ip-allowed",
		APIKey:         "sk-test-ip-allowed",
		AllowedIPCIDRs: `["10.0.0.0/8"]`,
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	req.RemoteAddr = "10.0.1.5:12345"
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	if !enforceAPIKeyPermissions(c, apiKey) {
		t.Fatalf("enforceAPIKeyPermissions() returned false for allowed IP, expected true")
	}
}

func TestEnforceAPIKeyPermissionsIPDenied(t *testing.T) {
	t.Parallel()

	apiKey := model.APIKey{
		ID:             7,
		Name:           "test-ip-denied",
		APIKey:         "sk-test-ip-denied",
		AllowedIPCIDRs: `["10.0.0.0/8"]`,
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	req.RemoteAddr = "11.0.0.1:12345"
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	if enforceAPIKeyPermissions(c, apiKey) {
		t.Fatalf("enforceAPIKeyPermissions() returned true for denied IP, expected false")
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", rec.Code)
	}
}

func TestEnforceAPIKeyPermissionsStoresChannelGroupInContext(t *testing.T) {
	t.Parallel()

	apiKey := model.APIKey{
		ID:              8,
		Name:            "test-ctx-store",
		APIKey:          "sk-test-ctx-store",
		AllowedChannels: `[1,3,5]`,
		AllowedGroups:   `[2,4]`,
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	if !enforceAPIKeyPermissions(c, apiKey) {
		t.Fatalf("enforceAPIKeyPermissions() returned false, expected true")
	}

	channels := APIKeyAllowedChannelsFromContext(c)
	if channels == nil {
		t.Fatalf("expected non-nil AllowedChannels in context")
	}
	if len(channels) != 3 {
		t.Fatalf("expected 3 channels, got %d", len(channels))
	}

	groups := APIKeyAllowedGroupsFromContext(c)
	if groups == nil {
		t.Fatalf("expected non-nil AllowedGroups in context")
	}
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
}

func TestAPIKeyAllowedChannelsFromContextNil(t *testing.T) {
	t.Parallel()

	// 无 context 或无 key 时返回 nil
	if got := APIKeyAllowedChannelsFromContext(nil); got != nil {
		t.Fatalf("expected nil for nil context, got %v", got)
	}

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	if got := APIKeyAllowedChannelsFromContext(c); got != nil {
		t.Fatalf("expected nil for context without key, got %v", got)
	}
}

func TestAPIKeyAllowedGroupsFromContextNil(t *testing.T) {
	t.Parallel()

	if got := APIKeyAllowedGroupsFromContext(nil); got != nil {
		t.Fatalf("expected nil for nil context, got %v", got)
	}

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	if got := APIKeyAllowedGroupsFromContext(c); got != nil {
		t.Fatalf("expected nil for context without key, got %v", got)
	}
}

func TestEnforceAPIKeyPermissionsNilContext(t *testing.T) {
	t.Parallel()

	// nil context 应该通过（不阻塞）
	apiKey := model.APIKey{ID: 9, Name: "test-nil-ctx", APIKey: "sk-test-nil-ctx"}
	if !enforceAPIKeyPermissions(nil, apiKey) {
		t.Fatalf("enforceAPIKeyPermissions(nil, ...) returned false, expected true")
	}
}

func TestEnforceAPIKeyPermissionsInvalidJSONFields(t *testing.T) {
	t.Parallel()

	// 无效 JSON 字段应该被视为"不限制"（向后兼容）
	apiKey := model.APIKey{
		ID:                  10,
		Name:                "test-invalid-json",
		APIKey:              "sk-test-invalid-json",
		AllowedCapabilities: `invalid json`,
		AllowedIPCIDRs:      `also invalid`,
		AllowedChannels:     `not json`,
		AllowedGroups:       `bad`,
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	// 无效 JSON 被解析为 nil → 不限制 → 应该通过
	if !enforceAPIKeyPermissions(c, apiKey) {
		t.Fatalf("enforceAPIKeyPermissions() returned false for invalid JSON fields, expected true (treated as no restriction)")
	}
}
