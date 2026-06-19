package op

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

const maxNewAPIInspectResponseBytes int64 = 4 << 20

func normalizeNewAPIBaseURL(raw string) (siteBase string, apiBase string, err error) {
	base, err := NormalizePublicBaseURL(raw)
	if err != nil {
		return "", "", err
	}
	if base == "" {
		return "", "", fmt.Errorf("base_url is required")
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", "", fmt.Errorf("base_url is invalid: %w", err)
	}
	path := strings.TrimRight(parsed.EscapedPath(), "/")
	if strings.EqualFold(path, "/v1") {
		site := *parsed
		site.Path = ""
		site.RawPath = ""
		site.RawQuery = ""
		site.Fragment = ""
		return strings.TrimRight(site.String(), "/"), base, nil
	}
	site := *parsed
	site.RawQuery = ""
	site.Fragment = ""
	siteBase = strings.TrimRight(site.String(), "/")
	apiBase = siteBase + "/v1"
	return siteBase, apiBase, nil
}

func fetchNewAPIModels(ctx context.Context, httpClient *http.Client, apiBase, token string) ([]string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(apiBase, "/")+"/models", nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Accept", "application/json")
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(response.Body, maxNewAPIInspectResponseBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(payload)) > maxNewAPIInspectResponseBytes {
		return nil, fmt.Errorf("new api models response too large")
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("new api models request failed: %s", response.Status)
	}
	models := parseOpenAIModelList(payload)
	if len(models) == 0 {
		return nil, fmt.Errorf("new api models response does not contain model ids")
	}
	sort.Strings(models)
	return models, nil
}

func parseOpenAIModelList(payload []byte) []string {
	var wrapped struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &wrapped); err == nil && len(wrapped.Data) > 0 {
		out := make([]string, 0, len(wrapped.Data))
		seen := make(map[string]struct{}, len(wrapped.Data))
		for _, item := range wrapped.Data {
			name := strings.TrimSpace(item.ID)
			if name == "" {
				continue
			}
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			out = append(out, name)
		}
		return out
	}
	var plain []string
	if err := json.Unmarshal(payload, &plain); err == nil {
		out := make([]string, 0, len(plain))
		seen := make(map[string]struct{}, len(plain))
		for _, item := range plain {
			name := strings.TrimSpace(item)
			if name == "" {
				continue
			}
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			out = append(out, name)
		}
		return out
	}
	return nil
}

func inferNewAPIRequestCapabilities(models []string) []string {
	capabilities := map[string]struct{}{
		model.RequestCapabilityOpenAIChat:      {},
		model.RequestCapabilityOpenAIResponses: {},
	}
	for _, modelName := range models {
		lower := strings.ToLower(strings.TrimSpace(modelName))
		switch {
		case strings.Contains(lower, "embedding"):
			capabilities[model.RequestCapabilityOpenAIEmbeddings] = struct{}{}
		case strings.HasPrefix(lower, "claude-"):
			capabilities[model.RequestCapabilityAnthropicMessages] = struct{}{}
		case strings.HasPrefix(lower, "gemini-"):
			capabilities[model.RequestCapabilityGeminiContents] = struct{}{}
		}
	}
	out := make([]string, 0, len(capabilities))
	for capability := range capabilities {
		out = append(out, capability)
	}
	sort.Strings(out)
	return out
}

func fetchNewAPITokenUsage(ctx context.Context, httpClient *http.Client, siteBase, token string) (model.NewAPITokenUsage, []string) {
	for _, path := range []string{"/api/usage/token", "/api/usage/token/", "/api/token/self"} {
		payload, ok := fetchJSONWithBearerUserCached(ctx, httpClient, strings.TrimRight(siteBase, "/")+path, token, "")
		if !ok {
			continue
		}
		usage, ok := parseNewAPITokenUsage(payload)
		if ok {
			return usage, nil
		}
	}
	return model.NewAPITokenUsage{Available: false}, []string{"token usage endpoint is not available for this New API token"}
}

func parseNewAPITokenUsage(payload []byte) (model.NewAPITokenUsage, bool) {
	var raw any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return model.NewAPITokenUsage{}, false
	}
	records := flattenObjectRecords(raw)
	for _, record := range records {
		usage := model.NewAPITokenUsage{Available: true}
		hasUsageField := false
		if value, ok := numberField(record, "used_quota", "usedQuota", "used", "usage", "quota_used", "total_used"); ok {
			usage.UsedQuota = value
			hasUsageField = true
		}
		if value, ok := numberField(record, "remain_quota", "remainQuota", "remaining", "balance", "quota", "total_available"); ok {
			usage.RemainQuota = value
			hasUsageField = true
		}
		if value, ok := boolField(record, "unlimited", "is_unlimited", "unlimited_quota"); ok {
			usage.Unlimited = value
			hasUsageField = true
		}
		if value, ok := stringField(record, "status", "message"); ok {
			usage.RawStatusText = value
			hasUsageField = true
		}
		if hasUsageField {
			return usage, true
		}
	}
	return model.NewAPITokenUsage{}, false
}

func flattenObjectRecords(value any) []map[string]any {
	switch typed := value.(type) {
	case map[string]any:
		out := []map[string]any{typed}
		for _, key := range []string{
			"data", "result", "results", "list", "items", "records", "rows",
			"token", "tokens", "key", "keys", "api_keys", "apiKeys",
			"user", "profile", "account",
			"group", "groups", "subscriptions", "plans", "models", "pricing", "prices",
		} {
			if nested, ok := typed[key]; ok {
				out = append(out, flattenObjectRecords(nested)...)
			}
		}
		return out
	case []any:
		out := make([]map[string]any, 0, len(typed))
		for _, item := range typed {
			out = append(out, flattenObjectRecords(item)...)
		}
		return out
	default:
		return nil
	}
}

func numberField(record map[string]any, keys ...string) (float64, bool) {
	for _, key := range keys {
		switch value := record[key].(type) {
		case float64:
			return value, true
		case int:
			return float64(value), true
		case string:
			if parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64); err == nil {
				return parsed, true
			}
		}
	}
	return 0, false
}

func boolField(record map[string]any, keys ...string) (bool, bool) {
	for _, key := range keys {
		switch value := record[key].(type) {
		case bool:
			return value, true
		case string:
			if parsed, err := strconv.ParseBool(strings.TrimSpace(value)); err == nil {
				return parsed, true
			}
		}
	}
	return false, false
}

func stringField(record map[string]any, keys ...string) (string, bool) {
	for _, key := range keys {
		if value, ok := record[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value), true
		}
	}
	return "", false
}
