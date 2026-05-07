// Package xurl provides utilities for URL parsing and manipulation.
package xurl

import (
	"fmt"
	"net/url"
	"strings"
)

// ValidateAbsoluteHTTPURL checks whether raw is an absolute http or https URL.
func ValidateAbsoluteHTTPURL(raw, fieldName string) error {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return fmt.Errorf("%s is invalid: %w", fieldName, err)
	}
	scheme := strings.ToLower(strings.TrimSpace(parsed.Scheme))
	if !parsed.IsAbs() || (scheme != "http" && scheme != "https") || strings.TrimSpace(parsed.Host) == "" {
		return fmt.Errorf("%s must be absolute http or https URL", fieldName)
	}
	if parsed.User != nil {
		return fmt.Errorf("%s must not include credentials", fieldName)
	}
	return nil
}

// ValidateProxyURL checks whether raw is an absolute proxy URL with an allowed scheme.
// Supported schemes: http, https, socks, socks5.
func ValidateProxyURL(raw, fieldName string) error {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return fmt.Errorf("%s is invalid: %w", fieldName, err)
	}
	scheme := strings.ToLower(strings.TrimSpace(parsed.Scheme))
	if !parsed.IsAbs() {
		return fmt.Errorf("%s must be absolute proxy URL", fieldName)
	}
	switch scheme {
	case "http", "https", "socks", "socks5":
	default:
		return fmt.Errorf("%s scheme must be http, https, socks, or socks5", fieldName)
	}
	if strings.TrimSpace(parsed.Host) == "" {
		return fmt.Errorf("%s must have a host", fieldName)
	}
	if parsed.User != nil {
		return fmt.Errorf("%s must not include credentials", fieldName)
	}
	return nil
}
