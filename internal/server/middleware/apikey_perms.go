package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/resp"
)

// API Key 权限粒度相关 context key。
// AllowedChannels/AllowedGroups 以 []int 形式存入 context，供 relay 过滤候选渠道使用。
const (
	ctxKeyAPIKeyAllowedChannels = "api_key_allowed_channels"
	ctxKeyAPIKeyAllowedGroups   = "api_key_allowed_groups"
)

// 能力取值常量，与请求路径的映射关系见 capabilityForPath。
const (
	capabilityChat      = "chat"
	capabilityEmbedding = "embedding"
	capabilityResponse  = "response"
	capabilityMessage   = "message"
)

// capabilityForPath 根据请求路径推断能力类型。
// 路径映射：
//   /v1/chat/completions → chat
//   /v1/embeddings       → embedding
//   /v1/responses        → response
//   /v1/messages         → message
// 未匹配返回空字符串，调用方据此判定是否拒绝（当 AllowedCapabilities 配置时）。
func capabilityForPath(path string) string {
	switch {
	case strings.HasSuffix(path, "/chat/completions"):
		return capabilityChat
	case strings.HasSuffix(path, "/embeddings"):
		return capabilityEmbedding
	case strings.HasSuffix(path, "/responses"):
		return capabilityResponse
	case strings.HasSuffix(path, "/messages"):
		return capabilityMessage
	default:
		return ""
	}
}

// enforceAPIKeyPermissions 在 API Key 认证通过后执行权限粒度校验：
//   - AllowedCapabilities：根据请求路径推断能力，若不在白名单内返回 403
//   - AllowedIPCIDRs：解析 CIDR，检查 client IP 是否在白名单内，不在则返回 403
//   - 将 AllowedChannels/AllowedGroups 存入 gin.Context 供 relay 使用
//
// 所有字段为空时表示不限制，保证向后兼容。
// 返回 true 表示通过，false 表示已拒绝并写入响应（调用方应 Abort）。
func enforceAPIKeyPermissions(c *gin.Context, apiKey model.APIKey) bool {
	if c == nil {
		return true
	}

	// 能力校验
	if allowed := apiKey.APIKeyAllowedCapabilities(); allowed != nil {
		cap := capabilityForPath(c.Request.URL.Path)
		if cap == "" || !containsString(allowed, cap) {
			resp.Error(c, http.StatusForbidden, "API key capability not allowed")
			c.Abort()
			return false
		}
	}

	// IP CIDR 白名单校验
	if cidrs := apiKey.APIKeyAllowedIPCIDRs(); cidrs != nil {
		clientIP := op.ClientIPFromRequest(c.Request)
		if clientIP == "" || !clientIPAllowed(clientIP, cidrs) {
			resp.Error(c, http.StatusForbidden, "client IP not allowed for this API key")
			c.Abort()
			return false
		}
	}

	// 将渠道/分组限制存入 context，供 relay 过滤候选
	c.Set(ctxKeyAPIKeyAllowedChannels, apiKey.APIKeyAllowedChannels())
	c.Set(ctxKeyAPIKeyAllowedGroups, apiKey.APIKeyAllowedGroups())
	return true
}

// clientIPAllowed 检查 client IP 是否落在任一 CIDR 内。
// 支持 IPv4 和 IPv6。无效 CIDR 会被跳过；无效 IP 直接拒绝。
func clientIPAllowed(clientIP string, cidrs []string) bool {
	ip := net.ParseIP(strings.TrimSpace(clientIP))
	if ip == nil {
		return false
	}
	for _, raw := range cidrs {
		cidr := strings.TrimSpace(raw)
		if cidr == "" {
			continue
		}
		// 兼容纯 IP 写法（无掩码）：按 /32 (v4) 或 /128 (v6) 处理
		if !strings.Contains(cidr, "/") {
			parsed := net.ParseIP(cidr)
			if parsed == nil {
				continue
			}
			if parsed.Equal(ip) {
				return true
			}
			continue
		}
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

// APIKeyAllowedChannelsFromContext 从 gin.Context 读取 API Key 允许的渠道 ID 列表。
// 返回 nil 表示不限制。
func APIKeyAllowedChannelsFromContext(c *gin.Context) []int {
	if c == nil {
		return nil
	}
	v, ok := c.Get(ctxKeyAPIKeyAllowedChannels)
	if !ok || v == nil {
		return nil
	}
	ids, ok := v.([]int)
	if !ok {
		return nil
	}
	if len(ids) == 0 {
		return nil
	}
	return ids
}

// APIKeyAllowedGroupsFromContext 从 gin.Context 读取 API Key 允许的分组 ID 列表。
// 返回 nil 表示不限制。
func APIKeyAllowedGroupsFromContext(c *gin.Context) []int {
	if c == nil {
		return nil
	}
	v, ok := c.Get(ctxKeyAPIKeyAllowedGroups)
	if !ok || v == nil {
		return nil
	}
	ids, ok := v.([]int)
	if !ok {
		return nil
	}
	if len(ids) == 0 {
		return nil
	}
	return ids
}
