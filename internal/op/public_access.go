package op

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/xurl"
)

func splitSettingCSV(value string) []string {
	normalized := strings.ReplaceAll(value, "\n", ",")
	normalized = strings.ReplaceAll(normalized, ";", ",")
	parts := strings.Split(normalized, ",")
	out := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		part = strings.TrimRight(strings.TrimSpace(part), "/")
		if part == "" {
			continue
		}
		key := strings.ToLower(part)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, part)
	}
	return out
}

func NormalizePublicBaseURL(raw string) (string, error) {
	trimmed := strings.TrimRight(strings.TrimSpace(raw), "/")
	if trimmed == "" {
		return "", nil
	}
	if err := xurl.ValidateAbsoluteHTTPURL(trimmed, "api base URL"); err != nil {
		return "", err
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("api base URL is invalid: %w", err)
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("api base URL must not contain query or fragment")
	}
	if parsed.User != nil {
		return "", fmt.Errorf("api base URL must not include credentials")
	}
	return trimmed, nil
}

func PublicBaseURLs() (string, []string) {
	primary, _ := SettingGetString(model.SettingKeyAPIBaseURL)
	primary, _ = NormalizePublicBaseURL(primary)
	alternateRaw, _ := SettingGetString(model.SettingKeyAPIAlternateBaseURLs)
	alternate := make([]string, 0)
	seen := make(map[string]struct{})
	if primary != "" {
		seen[strings.ToLower(primary)] = struct{}{}
	}
	for _, item := range splitSettingCSV(alternateRaw) {
		normalized, err := NormalizePublicBaseURL(item)
		if err != nil || normalized == "" {
			continue
		}
		key := strings.ToLower(normalized)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		alternate = append(alternate, normalized)
	}
	return primary, alternate
}

func CurrentRequestBaseURL(req *http.Request) string {
	scheme := "http"
	host := "localhost"
	if req != nil {
		if req.TLS != nil {
			scheme = "https"
		}
		if h := strings.TrimSpace(req.Host); h != "" {
			host = h
		}
		if remoteIPTrusted(remoteIPFromRequest(req)) {
			if forwardedProto := firstHeaderValue(req.Header.Get("X-Forwarded-Proto")); forwardedProto == "http" || forwardedProto == "https" {
				scheme = forwardedProto
			}
			if forwardedHost := firstHeaderValue(req.Header.Get("X-Forwarded-Host")); forwardedHost != "" {
				host = forwardedHost
			}
		}
	}
	return strings.TrimRight(fmt.Sprintf("%s://%s", scheme, host), "/")
}

func TrustedProxyCIDRs() []string {
	raw, _ := SettingGetString(model.SettingKeyTrustedProxyCIDRs)
	return splitSettingCSV(raw)
}

func trustedProxyNets() []*net.IPNet {
	values := TrustedProxyCIDRs()
	out := make([]*net.IPNet, 0, len(values))
	for _, value := range values {
		if strings.Contains(value, "/") {
			if _, cidr, err := net.ParseCIDR(value); err == nil {
				out = append(out, cidr)
			}
			continue
		}
		ip := net.ParseIP(value)
		if ip == nil {
			continue
		}
		bits := 32
		if ip.To4() == nil {
			bits = 128
		}
		out = append(out, &net.IPNet{IP: ip, Mask: net.CIDRMask(bits, bits)})
	}
	return out
}

func remoteIPFromRequest(req *http.Request) string {
	if req == nil {
		return ""
	}
	remoteAddr := strings.TrimSpace(req.RemoteAddr)
	if remoteAddr == "" {
		return ""
	}
	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil {
		return normalizeIP(host)
	}
	return normalizeIP(remoteAddr)
}

func normalizeIP(value string) string {
	value = strings.Trim(strings.TrimSpace(value), "[]")
	if value == "" {
		return ""
	}
	ip := net.ParseIP(value)
	if ip == nil {
		return value
	}
	return ip.String()
}

func remoteIPTrusted(remoteIP string) bool {
	ip := net.ParseIP(normalizeIP(remoteIP))
	if ip == nil {
		return false
	}
	for _, cidr := range trustedProxyNets() {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

func firstHeaderIP(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	for _, part := range strings.Split(value, ",") {
		ip := normalizeIP(part)
		if net.ParseIP(ip) != nil {
			return ip
		}
	}
	return ""
}

func firstHeaderValue(value string) string {
	for _, part := range strings.Split(value, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func ClientIPFromRequest(req *http.Request) string {
	remoteIP := remoteIPFromRequest(req)
	if req == nil || !remoteIPTrusted(remoteIP) {
		return remoteIP
	}
	for _, header := range []string{"X-Forwarded-For", "X-Real-IP", "CF-Connecting-IP", "True-Client-IP"} {
		if ip := firstHeaderIP(req.Header.Get(header)); ip != "" {
			return ip
		}
	}
	return remoteIP
}

func OpsIPDisplayMode() string {
	value, _ := SettingGetString(model.SettingKeyOpsIPDisplayMode)
	value = strings.ToLower(strings.TrimSpace(value))
	if value == model.OpsIPDisplayModeFull {
		return model.OpsIPDisplayModeFull
	}
	return model.OpsIPDisplayModeMasked
}

func MaskClientIP(ip string) string {
	ip = normalizeIP(ip)
	if ip == "" {
		return ""
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return ip
	}
	if v4 := parsed.To4(); v4 != nil {
		return fmt.Sprintf("%d.%d.%d.*", v4[0], v4[1], v4[2])
	}
	segments := strings.Split(parsed.String(), ":")
	if len(segments) <= 2 {
		return parsed.String()
	}
	return strings.Join(segments[:2], ":") + ":*"
}

func IPDisplayLabel(ip string) string {
	if OpsIPDisplayMode() == model.OpsIPDisplayModeFull {
		return normalizeIP(ip)
	}
	return MaskClientIP(ip)
}

func PublicAccessInfo(req *http.Request) model.PublicAccessInfo {
	primary, alternate := PublicBaseURLs()
	clientIP := ClientIPFromRequest(req)
	return model.PublicAccessInfo{
		PrimaryBaseURL:     primary,
		AlternateBaseURLs:  alternate,
		CurrentBaseURL:     CurrentRequestBaseURL(req),
		TrustedProxyCIDRs:  TrustedProxyCIDRs(),
		OpsIPDisplayMode:   OpsIPDisplayMode(),
		CurrentClientIP:    clientIP,
		CurrentClientLabel: IPDisplayLabel(clientIP),
	}
}
