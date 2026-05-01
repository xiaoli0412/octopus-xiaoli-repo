package op

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

func resolveAIAutomationBaseURL(raw string, useLocalDefault bool) (string, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(raw), "/")
	if useLocalDefault && baseURL == "" {
		baseURL = strings.TrimRight(model.DefaultAIAutomationBaseURL, "/")
	}
	if baseURL == "" {
		return "", fmt.Errorf("ai automation base URL is required")
	}
	if err := validateAIAutomationBaseURL(baseURL); err != nil {
		return "", err
	}
	return baseURL, nil
}

func validateAIAutomationBaseURL(raw string) error {
	trimmed := strings.TrimRight(strings.TrimSpace(raw), "/")
	if trimmed == "" {
		return nil
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return fmt.Errorf("ai automation base URL is invalid: %w", err)
	}
	scheme := strings.ToLower(strings.TrimSpace(parsed.Scheme))
	if !parsed.IsAbs() || (scheme != "http" && scheme != "https") || strings.TrimSpace(parsed.Host) == "" {
		return fmt.Errorf("ai automation base URL must be absolute http or https URL")
	}
	if strings.TrimSpace(parsed.RawQuery) != "" || strings.TrimSpace(parsed.Fragment) != "" {
		return fmt.Errorf("ai automation base URL must not contain query or fragment")
	}
	if isDefaultAIAutomationBaseURL(trimmed) {
		return nil
	}
	host := normalizeAIAutomationBaseURLHost(parsed.Hostname())
	if host == "localhost" {
		return fmt.Errorf("ai automation base URL must not target localhost; use the built-in local default instead")
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return nil
	}
	if isAIAutomationRestrictedIP(ip) {
		return fmt.Errorf("ai automation base URL must not target loopback or private IP addresses; use the built-in local default instead")
	}
	return nil
}

func isDefaultAIAutomationBaseURL(raw string) bool {
	return normalizeAIAutomationBaseURL(raw) == normalizeAIAutomationBaseURL(model.DefaultAIAutomationBaseURL)
}

func normalizeAIAutomationBaseURL(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return ""
	}
	host := normalizeAIAutomationBaseURLHost(parsed.Hostname())
	port := parsed.Port()
	path := strings.TrimRight(parsed.EscapedPath(), "/")
	authority := host
	if port != "" {
		authority = net.JoinHostPort(host, port)
	}
	return strings.ToLower(parsed.Scheme) + "://" + authority + path
}

func normalizeAIAutomationBaseURLHost(host string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(host)), ".")
}

func isAIAutomationRestrictedIP(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified()
}
