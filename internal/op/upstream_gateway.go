package op

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/llmname"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	transformerOutbound "github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/outbound"
)

const upstreamInspectTimeout = 25 * time.Second

func normalizeUpstreamProvider(input string) string {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case model.UpstreamProviderSub2API, "sub2", "sub2-api":
		return model.UpstreamProviderSub2API
	case model.UpstreamProviderOpenAICompatible, "openai", "openai-compatible":
		return model.UpstreamProviderOpenAICompatible
	default:
		return model.UpstreamProviderNewAPI
	}
}

func normalizeUpstreamAuthMode(input string) string {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case model.UpstreamAuthModeAccessKey, "api_key", "apikey", "gateway_key":
		return model.UpstreamAuthModeAccessKey
	case model.UpstreamAuthModeAccountPassword, "password", "account":
		return model.UpstreamAuthModeAccountPassword
	default:
		return model.UpstreamAuthModeToken
	}
}

func upstreamOfficialSource(providerType, siteBase string) (bool, string) {
	parsed, err := url.Parse(siteBase)
	if err != nil {
		return false, "第三方部署"
	}
	host := strings.ToLower(parsed.Hostname())
	switch providerType {
	case model.UpstreamProviderSub2API:
		if host == "sub2api.org" || strings.HasSuffix(host, ".sub2api.org") || host == "pincc.ai" || strings.HasSuffix(host, ".pincc.ai") {
			return true, "sub2API 官方域名"
		}
		return false, "sub2API 第三方部署"
	case model.UpstreamProviderNewAPI:
		return false, "New API 兼容站点"
	default:
		return false, "OpenAI 兼容站点"
	}
}

func InspectNewAPI(ctx context.Context, req model.NewAPIInspectRequest) (model.NewAPIInspectResult, error) {
	result, err := InspectUpstreamGateway(ctx, model.UpstreamInspectRequest{
		ProviderType: model.UpstreamProviderNewAPI,
		BaseURL:      req.BaseURL,
		AuthMode:     model.UpstreamAuthModeToken,
		Token:        req.Token,
	})
	if err != nil {
		return model.NewAPIInspectResult{}, err
	}
	return model.NewAPIInspectResult{
		BaseURL:             result.BaseURL,
		APIBaseURL:          result.APIBaseURL,
		ModelCount:          result.ModelCount,
		Models:              result.Models,
		RequestCapabilities: result.RequestCapabilities,
		TokenUsage:          result.TokenUsage,
		Warnings:            result.Warnings,
	}, nil
}

func InspectUpstreamGateway(ctx context.Context, req model.UpstreamInspectRequest) (model.UpstreamInspectResult, error) {
	providerType := normalizeUpstreamProvider(req.ProviderType)
	authMode := normalizeUpstreamAuthMode(req.AuthMode)
	siteBase, apiBase, err := normalizeNewAPIBaseURL(req.BaseURL)
	if err != nil {
		return model.UpstreamInspectResult{}, err
	}
	httpClient, err := newHealthHTTPClientNoProxy()
	if err != nil {
		return model.UpstreamInspectResult{}, err
	}
	ctx, cancel := context.WithTimeout(ctx, upstreamInspectTimeout)
	defer cancel()

	official, sourceLabel := upstreamOfficialSource(providerType, siteBase)
	result := model.UpstreamInspectResult{
		ProviderType:   providerType,
		AuthMode:       authMode,
		BaseURL:        siteBase,
		APIBaseURL:     apiBase,
		OfficialSource: official,
		SourceLabel:    sourceLabel,
		Warnings:       make([]string, 0),
	}

	managementToken := strings.TrimSpace(req.Token)
	managementUserID := strings.TrimSpace(req.UserID)
	gatewayKey := strings.TrimSpace(req.AccessKey)
	if gatewayKey == "" && authMode == model.UpstreamAuthModeToken {
		gatewayKey = strings.TrimSpace(req.Token)
	}

	if authMode == model.UpstreamAuthModeAccountPassword && (providerType == model.UpstreamProviderSub2API || providerType == model.UpstreamProviderNewAPI) {
		loginToken, loginUserID, loginWarnings := upstreamGatewayLogin(ctx, httpClient, providerType, siteBase, req.Username, req.Password)
		result.Warnings = append(result.Warnings, loginWarnings...)
		if loginToken != "" {
			managementToken = loginToken
			result.ManagementToken = loginToken
			gatewayKey = ""
		}
		if managementUserID == "" {
			managementUserID = loginUserID
		}
	}
	if managementToken != "" && result.ManagementToken == "" {
		result.ManagementToken = managementToken
	}
	if gatewayKey != "" {
		result.GatewayAccessKey = gatewayKey
	}

	if providerType == model.UpstreamProviderSub2API && managementToken != "" {
		usage, subscriptions, warnings := fetchSub2APIProfileUsage(ctx, httpClient, siteBase, managementToken)
		if usage.Available {
			result.TokenUsage = usage
		}
		result.Subscriptions = subscriptions
		result.Warnings = append(result.Warnings, warnings...)

		keys, warnings := fetchSub2APIKeys(ctx, httpClient, siteBase, managementToken)
		result.Keys = keys
		result.Warnings = append(result.Warnings, warnings...)
		groups, warnings := fetchSub2APIGroups(ctx, httpClient, siteBase, managementToken)
		result.Groups = mergeUpstreamGroups(result.Groups, groups)
		result.Warnings = append(result.Warnings, warnings...)
		subscriptions, warnings = fetchSub2APISubscriptions(ctx, httpClient, siteBase, managementToken)
		result.Subscriptions = mergeUpstreamSubscriptions(result.Subscriptions, subscriptions)
		result.Warnings = append(result.Warnings, warnings...)
		if gatewayKey == "" {
			for _, key := range keys {
				if strings.TrimSpace(key.Key) != "" {
					gatewayKey = strings.TrimSpace(key.Key)
					break
				}
			}
		}
	}

	if providerType == model.UpstreamProviderNewAPI && managementToken != "" {
		usage, warnings := fetchNewAPIAccountUsage(ctx, httpClient, siteBase, managementToken, managementUserID)
		if usage.Available {
			result.TokenUsage = usage
		}
		result.Warnings = append(result.Warnings, warnings...)
		groups, warnings := fetchNewAPIGroups(ctx, httpClient, siteBase, managementToken, managementUserID)
		result.Groups = mergeUpstreamGroups(result.Groups, groups)
		result.Warnings = append(result.Warnings, warnings...)
		models, warnings := fetchNewAPIUserModels(ctx, httpClient, siteBase, managementToken, managementUserID)
		result.Models = append(result.Models, models...)
		result.Warnings = append(result.Warnings, warnings...)
		keys, warnings := fetchNewAPIKeys(ctx, httpClient, siteBase, managementToken, managementUserID)
		result.Keys = keys
		result.Warnings = append(result.Warnings, warnings...)
	}

	if gatewayKey != "" {
		models, err := fetchUpstreamGatewayModels(ctx, httpClient, apiBase, gatewayKey, providerType, authMode)
		if err != nil {
			result.Warnings = append(result.Warnings, err.Error())
		} else {
			result.Models = append(result.Models, models...)
		}

		if providerType == model.UpstreamProviderSub2API {
			usage, warnings := fetchSub2APIUsage(ctx, httpClient, siteBase, gatewayKey)
			if usage.Available {
				result.TokenUsage = usage
			}
			result.Warnings = append(result.Warnings, warnings...)
		}
	} else if providerType == model.UpstreamProviderSub2API {
		result.Warnings = append(result.Warnings, "未提供或未读取到 gateway API key，已跳过 /v1/models 探测")
	} else if len(result.Models) == 0 {
		result.Warnings = append(result.Warnings, "token is required")
	}

	if providerType != model.UpstreamProviderSub2API && gatewayKey != "" {
		usage, warnings := fetchNewAPITokenUsage(ctx, httpClient, siteBase, gatewayKey)
		if usage.Available && !result.TokenUsage.Available {
			result.TokenUsage = usage
		}
		result.Warnings = append(result.Warnings, warnings...)
	}

	result.Models = normalizeModelNames(append(result.Models, modelsFromUpstreamKeys(result.Keys)...))
	result.ModelCount = len(result.Models)
	result.RequestCapabilities = inferNewAPIRequestCapabilities(result.Models)
	if len(result.RequestCapabilities) == 0 && len(result.Models) > 0 {
		result.RequestCapabilities = []string{model.RequestCapabilityOpenAIChat}
	}
	upstreamPrices, warnings := fetchUpstreamPriceCandidates(ctx, httpClient, siteBase, managementToken, managementUserID, providerType)
	result.Warnings = append(result.Warnings, warnings...)
	result.PriceCandidates = buildUpstreamPriceCandidates(result.Models, providerType, result.SourceLabel, upstreamPrices)
	result.Warnings = compactStrings(result.Warnings)

	if result.ModelCount == 0 && len(result.Keys) == 0 && !result.TokenUsage.Available {
		return model.UpstreamInspectResult{}, fmt.Errorf("unable to inspect upstream site; provide a valid gateway key or account token")
	}
	return result, nil
}

func upstreamGatewayAuthProbeOrder(providerType, authMode string) []string {
	if authMode == model.UpstreamAuthModeAccessKey || providerType == model.UpstreamProviderSub2API {
		return []string{"x-api-key", "bearer"}
	}
	return []string{"bearer", "x-api-key"}
}

func fetchUpstreamGatewayModels(ctx context.Context, httpClient *http.Client, apiBase, token, providerType, authMode string) ([]string, error) {
	endpoint := strings.TrimRight(apiBase, "/") + "/models"
	var lastErr error
	for _, mode := range upstreamGatewayAuthProbeOrder(providerType, authMode) {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, err
		}
		request.Header.Set("Accept", "application/json")
		if mode == "x-api-key" {
			request.Header.Set("x-api-key", token)
		} else {
			request.Header.Set("Authorization", "Bearer "+token)
		}
		response, err := httpClient.Do(request)
		if err != nil {
			lastErr = err
			continue
		}
		payload, readErr := io.ReadAll(io.LimitReader(response.Body, maxNewAPIInspectResponseBytes+1))
		_ = response.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}
		if int64(len(payload)) > maxNewAPIInspectResponseBytes {
			return nil, fmt.Errorf("upstream models response too large")
		}
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			lastErr = fmt.Errorf("upstream models request failed with %s auth: %s", mode, response.Status)
			continue
		}
		models := parseOpenAIModelList(payload)
		if len(models) == 0 {
			lastErr = fmt.Errorf("upstream models response does not contain model ids")
			continue
		}
		return normalizeModelNames(models), nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("upstream models request failed")
	}
	return nil, lastErr
}

func upstreamGatewayLogin(ctx context.Context, httpClient *http.Client, providerType, siteBase, username, password string) (string, string, []string) {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)
	if username == "" || password == "" {
		return "", "", []string{"账号密码模式需要填写用户名和密码"}
	}
	body, _ := json.Marshal(map[string]string{
		"email":    username,
		"username": username,
		"password": password,
	})
	paths := []string{"/api/v1/auth/login", "/auth/login"}
	if providerType == model.UpstreamProviderNewAPI {
		paths = []string{"/api/user/login", "/api/auth/login", "/api/v1/auth/login", "/auth/login"}
	}
	for _, path := range paths {
		request, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(siteBase, "/")+path, bytes.NewReader(body))
		if err != nil {
			continue
		}
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Accept", "application/json")
		response, err := httpClient.Do(request)
		if err != nil {
			continue
		}
		payload, readErr := io.ReadAll(io.LimitReader(response.Body, maxNewAPIInspectResponseBytes+1))
		_ = response.Body.Close()
		if readErr != nil || int64(len(payload)) > maxNewAPIInspectResponseBytes {
			continue
		}
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			continue
		}
		if token, requires2FA := parseAccessToken(payload); token != "" {
			userID := parseUpstreamUserID(payload)
			return token, userID, nil
		} else if requires2FA {
			return "", "", []string{"上游登录需要二次验证，已跳过账号管理接口"}
		}
	}
	return "", "", []string{"上游账号登录失败，已继续尝试 gateway key 探测"}
}

func parseAccessToken(payload []byte) (string, bool) {
	var raw any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return "", false
	}
	records := flattenObjectRecords(raw)
	for _, record := range records {
		if value, ok := boolField(record, "requires_2fa"); ok && value {
			return "", true
		}
		if value, ok := stringField(record, "access_token", "accessToken", "token", "jwt"); ok {
			return value, false
		}
	}
	return "", false
}

func parseUpstreamUserID(payload []byte) string {
	var raw any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return ""
	}
	for _, record := range flattenObjectRecords(raw) {
		if value, ok := stringField(record, "user_id", "userId", "id", "uid"); ok {
			return value
		}
		if value, ok := numberField(record, "user_id", "userId", "id", "uid"); ok {
			if value == float64(int64(value)) {
				return strconv.FormatInt(int64(value), 10)
			}
			return strconv.FormatFloat(value, 'f', -1, 64)
		}
	}
	return ""
}

func fetchSub2APIProfileUsage(ctx context.Context, httpClient *http.Client, siteBase, token string) (model.NewAPITokenUsage, []model.UpstreamSubscription, []string) {
	for _, path := range []string{"/api/v1/user/profile", "/api/v1/auth/me"} {
		payload, ok := fetchJSONWithBearer(ctx, httpClient, strings.TrimRight(siteBase, "/")+path, token)
		if !ok {
			continue
		}
		usage, parsed := parseNewAPITokenUsage(payload)
		subscriptions := parseUpstreamSubscriptions(payload)
		if parsed {
			return usage, subscriptions, nil
		}
		if len(subscriptions) > 0 {
			return model.NewAPITokenUsage{Available: false}, subscriptions, nil
		}
	}
	return model.NewAPITokenUsage{Available: false}, nil, nil
}

func fetchSub2APIKeys(ctx context.Context, httpClient *http.Client, siteBase, token string) ([]model.UpstreamGatewayKey, []string) {
	keys := make([]model.UpstreamGatewayKey, 0)
	for _, path := range []string{"/api/v1/api-keys", "/api/v1/keys"} {
		for _, payload := range fetchPagedJSONWithBearer(ctx, httpClient, siteBase, path, token, "", 4, 50) {
			keys = append(keys, parseUpstreamGatewayKeys(payload)...)
		}
		if len(keys) > 0 {
			break
		}
	}
	if len(keys) == 0 {
		return nil, []string{"API key 列表为空或字段不可识别"}
	}
	return dedupeUpstreamGatewayKeys(keys), nil
}

func fetchNewAPIKeys(ctx context.Context, httpClient *http.Client, siteBase, token, userID string) ([]model.UpstreamGatewayKey, []string) {
	keys := make([]model.UpstreamGatewayKey, 0)
	for _, path := range []string{"/api/token/search", "/api/token/"} {
		for page := 0; page < 4; page++ {
			endpoint := strings.TrimRight(siteBase, "/") + path
			separator := "?"
			if strings.Contains(endpoint, "?") {
				separator = "&"
			}
			endpoint += separator + "p=" + strconv.Itoa(page) + "&page_size=50&size=50"
			payload, ok := fetchJSONWithBearerUser(ctx, httpClient, endpoint, token, userID)
			if !ok {
				if page == 0 {
					continue
				}
				break
			}
			pageKeys := parseUpstreamGatewayKeys(payload)
			if len(pageKeys) == 0 {
				if page == 0 {
					continue
				}
				break
			}
			keys = append(keys, pageKeys...)
			if len(pageKeys) < 50 {
				break
			}
		}
		if len(keys) > 0 {
			break
		}
	}
	if len(keys) > 0 {
		return dedupeUpstreamGatewayKeys(keys), nil
	}
	return nil, []string{"New API token 列表不可用或未返回可识别 token"}
}

func fetchJSONWithBearer(ctx context.Context, httpClient *http.Client, endpoint, token string) ([]byte, bool) {
	return fetchJSONWithBearerUser(ctx, httpClient, endpoint, token, "")
}

func fetchJSONWithBearerUser(ctx context.Context, httpClient *http.Client, endpoint, token, userID string) ([]byte, bool) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, false
	}
	request.Header.Set("Authorization", "Bearer "+token)
	if strings.TrimSpace(userID) != "" {
		request.Header.Set("New-Api-User", strings.TrimSpace(userID))
	}
	request.Header.Set("Accept", "application/json")
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, false
	}
	defer response.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(response.Body, maxNewAPIInspectResponseBytes+1))
	if err != nil || response.StatusCode < 200 || response.StatusCode >= 300 || int64(len(payload)) > maxNewAPIInspectResponseBytes {
		return nil, false
	}
	return payload, true
}

func fetchPagedJSONWithBearer(ctx context.Context, httpClient *http.Client, siteBase, path, token, userID string, maxPages, pageSize int) [][]byte {
	out := make([][]byte, 0)
	if maxPages <= 0 {
		maxPages = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}
	for page := 1; page <= maxPages; page++ {
		endpoint := strings.TrimRight(siteBase, "/") + path
		separator := "?"
		if strings.Contains(endpoint, "?") {
			separator = "&"
		}
		endpoint += separator + "page=" + strconv.Itoa(page) + "&page_size=" + strconv.Itoa(pageSize)
		payload, ok := fetchJSONWithBearerUser(ctx, httpClient, endpoint, token, userID)
		if !ok {
			if page == 1 {
				plainEndpoint := strings.TrimRight(siteBase, "/") + path
				payload, ok = fetchJSONWithBearerUser(ctx, httpClient, plainEndpoint, token, userID)
			}
			if !ok {
				break
			}
		}
		out = append(out, payload)
		if len(flattenObjectRecordsFromPayload(payload)) < pageSize {
			break
		}
	}
	return out
}

func fetchNewAPIAccountUsage(ctx context.Context, httpClient *http.Client, siteBase, token, userID string) (model.NewAPITokenUsage, []string) {
	for _, path := range []string{"/api/user/self", "/api/user/self/", "/api/user/profile", "/api/user"} {
		payload, ok := fetchJSONWithBearerUser(ctx, httpClient, strings.TrimRight(siteBase, "/")+path, token, userID)
		if !ok {
			continue
		}
		if usage, parsed := parseNewAPITokenUsage(payload); parsed {
			return usage, nil
		}
	}
	return model.NewAPITokenUsage{Available: false}, []string{"New API 账户余额不可用"}
}

func fetchNewAPIUserModels(ctx context.Context, httpClient *http.Client, siteBase, token, userID string) ([]string, []string) {
	for _, path := range []string{"/api/user/models", "/api/user/self/models", "/api/models"} {
		payload, ok := fetchJSONWithBearerUser(ctx, httpClient, strings.TrimRight(siteBase, "/")+path, token, userID)
		if !ok {
			continue
		}
		models := parseUpstreamModelNames(payload)
		if len(models) > 0 {
			return models, nil
		}
	}
	return nil, []string{"New API 用户模型目录不可用"}
}

func fetchNewAPIGroups(ctx context.Context, httpClient *http.Client, siteBase, token, userID string) ([]model.UpstreamGroup, []string) {
	for _, path := range []string{"/api/user/self/groups", "/api/user/groups", "/api/groups"} {
		payload, ok := fetchJSONWithBearerUser(ctx, httpClient, strings.TrimRight(siteBase, "/")+path, token, userID)
		if !ok {
			continue
		}
		groups := parseUpstreamGroups(payload, "New API")
		if len(groups) > 0 {
			return groups, nil
		}
	}
	return nil, []string{"New API 分组目录不可用"}
}

func fetchSub2APIGroups(ctx context.Context, httpClient *http.Client, siteBase, token string) ([]model.UpstreamGroup, []string) {
	groups := make([]model.UpstreamGroup, 0)
	for _, path := range []string{"/api/v1/groups", "/api/v1/groups/available", "/api/v1/groups/rates"} {
		payload, ok := fetchJSONWithBearer(ctx, httpClient, strings.TrimRight(siteBase, "/")+path, token)
		if !ok {
			continue
		}
		groups = append(groups, parseUpstreamGroups(payload, "sub2API")...)
	}
	if len(groups) == 0 {
		return nil, []string{"sub2API 分组目录不可用"}
	}
	return mergeUpstreamGroups(nil, groups), nil
}

func fetchSub2APISubscriptions(ctx context.Context, httpClient *http.Client, siteBase, token string) ([]model.UpstreamSubscription, []string) {
	subscriptions := make([]model.UpstreamSubscription, 0)
	for _, path := range []string{"/api/v1/subscriptions", "/api/v1/subscriptions/active", "/api/v1/subscriptions/summary"} {
		payload, ok := fetchJSONWithBearer(ctx, httpClient, strings.TrimRight(siteBase, "/")+path, token)
		if !ok {
			continue
		}
		subscriptions = append(subscriptions, parseUpstreamSubscriptions(payload)...)
	}
	if len(subscriptions) == 0 {
		return nil, []string{"sub2API 订阅明细不可用"}
	}
	return mergeUpstreamSubscriptions(nil, subscriptions), nil
}

func fetchUpstreamPriceCandidates(ctx context.Context, httpClient *http.Client, siteBase, token, userID, providerType string) (map[string]model.UpstreamPriceCandidate, []string) {
	if strings.TrimSpace(token) == "" {
		return nil, nil
	}
	paths := []string{"/api/pricing", "/api/ratio_config", "/api/user/pricing", "/api/user/models"}
	if providerType == model.UpstreamProviderSub2API {
		paths = []string{"/api/v1/pricing", "/api/v1/model/pricing", "/api/v1/models/pricing", "/api/v1/groups/rates", "/api/v1/groups/available"}
		userID = ""
	}
	out := make(map[string]model.UpstreamPriceCandidate)
	for _, path := range paths {
		payload, ok := fetchJSONWithBearerUser(ctx, httpClient, strings.TrimRight(siteBase, "/")+path, token, userID)
		if !ok {
			continue
		}
		for key, candidate := range parseUpstreamPriceCandidates(payload, providerType+" upstream") {
			out[key] = candidate
		}
	}
	if len(out) == 0 {
		return nil, []string{"上游价格/倍率目录不可用"}
	}
	return out, nil
}

func parseUpstreamModelNames(payload []byte) []string {
	models := parseOpenAIModelList(payload)
	if len(models) > 0 {
		return normalizeModelNames(models)
	}
	var raw any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil
	}
	out := make([]string, 0)
	for _, record := range flattenObjectRecords(raw) {
		if value, ok := stringField(record, "id", "name", "model", "model_name", "modelName"); ok {
			out = append(out, value)
		}
		out = append(out, splitLooseModelList(record["models"], record["model_names"], record["available_models"])...)
	}
	return normalizeModelNames(out)
}

func flattenObjectRecordsFromPayload(payload []byte) []map[string]any {
	var raw any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil
	}
	return flattenObjectRecords(raw)
}

func parseUpstreamGroups(payload []byte, source string) []model.UpstreamGroup {
	var raw any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil
	}
	out := make([]model.UpstreamGroup, 0)
	for _, record := range flattenObjectRecords(raw) {
		name, _ := stringField(record, "name", "group_name", "group", "title")
		id, _ := stringField(record, "id", "group_id", "key")
		if id == "" {
			if value, ok := numberField(record, "id", "group_id"); ok {
				id = strconv.FormatFloat(value, 'f', -1, 64)
			}
		}
		if name == "" && id == "" {
			continue
		}
		description, _ := stringField(record, "description", "desc")
		platform, _ := stringField(record, "platform", "provider")
		status, _ := stringField(record, "status", "state")
		rate, _ := numberField(record, "rate_multiplier", "rate", "multiplier", "group_ratio")
		groupsModels := splitLooseModelList(record["models"], record["model_names"], record["supported_models"], record["available_models"])
		out = append(out, model.UpstreamGroup{
			ID:                  id,
			Name:                firstNonEmptyUpstreamValue(name, id),
			Description:         description,
			Platform:            platform,
			Status:              status,
			RateMultiplier:      rate,
			Models:              groupsModels,
			RequestCapabilities: requestCapabilitiesFromRecord(record),
			Source:              source,
		})
	}
	return mergeUpstreamGroups(nil, out)
}

func parseUpstreamPriceCandidates(payload []byte, source string) map[string]model.UpstreamPriceCandidate {
	var raw any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil
	}
	out := make(map[string]model.UpstreamPriceCandidate)
	var walk func(any)
	walk = func(value any) {
		switch typed := value.(type) {
		case map[string]any:
			if ratioMap, ok := objectField(typed, "model_ratio", "modelRatio", "model_rates", "modelRates"); ok {
				completionMap, _ := objectField(typed, "completion_ratio", "completionRatio", "completion_rates", "completionRates")
				for modelName, ratioValue := range ratioMap {
					ratio, ok := numberFromAny(ratioValue)
					if !ok {
						continue
					}
					outputRatio := ratio
					if completion, ok := numberFromAny(completionMap[modelName]); ok && completion > 0 {
						outputRatio = ratio * completion
					}
					addUpstreamPriceCandidate(out, modelName, model.UpstreamPriceCandidate{
						Name:            modelName,
						LLMPrice:        model.LLMPrice{Input: ratio, Output: outputRatio},
						PriceSource:     source + " ratio",
						PriceMatchedKey: modelName,
						Sources:         []string{source},
					})
				}
			}
			if candidate, ok := priceCandidateFromRecord(typed, source); ok {
				addUpstreamPriceCandidate(out, candidate.Name, candidate)
			}
			for _, nested := range typed {
				walk(nested)
			}
		case []any:
			for _, item := range typed {
				walk(item)
			}
		}
	}
	walk(raw)
	return out
}

func objectField(record map[string]any, keys ...string) (map[string]any, bool) {
	for _, key := range keys {
		if value, ok := record[key].(map[string]any); ok {
			return value, true
		}
	}
	return nil, false
}

func priceCandidateFromRecord(record map[string]any, source string) (model.UpstreamPriceCandidate, bool) {
	name, _ := stringField(record, "name", "id", "model", "model_name", "modelName")
	if name == "" {
		return model.UpstreamPriceCandidate{}, false
	}
	input, hasInput := numberField(record, "input", "input_price", "prompt_price", "prompt", "input_cost", "input_cost_per_million", "input_cost_per_1m", "model_ratio", "rate_multiplier")
	output, hasOutput := numberField(record, "output", "output_price", "completion_price", "completion", "output_cost", "output_cost_per_million", "output_cost_per_1m")
	cacheRead, hasCacheRead := numberField(record, "cache_read", "cache_read_price", "cached_input", "cached_input_price")
	cacheWrite, hasCacheWrite := numberField(record, "cache_write", "cache_write_price", "cache_creation", "cache_creation_price")
	if !hasInput && !hasOutput && !hasCacheRead && !hasCacheWrite {
		return model.UpstreamPriceCandidate{}, false
	}
	price := model.LLMPrice{Input: input, Output: output, CacheRead: cacheRead, CacheWrite: cacheWrite}
	return model.UpstreamPriceCandidate{
		Name:            name,
		LLMPrice:        price,
		PriceSource:     source,
		PriceMatchedKey: name,
		Sources:         []string{source},
	}, true
}

func addUpstreamPriceCandidate(target map[string]model.UpstreamPriceCandidate, name string, candidate model.UpstreamPriceCandidate) {
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	key := strings.ToLower(name)
	if existing, ok := target[key]; ok {
		candidate.LLMPrice = mergeUpstreamLLMPrice(existing.LLMPrice, candidate.LLMPrice)
		candidate.Sources = compactStrings(append(existing.Sources, candidate.Sources...))
		if candidate.PriceSource == "" {
			candidate.PriceSource = existing.PriceSource
		}
		if candidate.PriceMatchedKey == "" {
			candidate.PriceMatchedKey = existing.PriceMatchedKey
		}
	}
	candidate.Name = name
	target[key] = candidate
}

func mergeUpstreamLLMPrice(current, incoming model.LLMPrice) model.LLMPrice {
	if current.Input == 0 {
		current.Input = incoming.Input
	}
	if current.Output == 0 {
		current.Output = incoming.Output
	}
	if current.CacheRead == 0 {
		current.CacheRead = incoming.CacheRead
	}
	if current.CacheWrite == 0 {
		current.CacheWrite = incoming.CacheWrite
	}
	return current
}

func parseUpstreamGatewayKeys(payload []byte) []model.UpstreamGatewayKey {
	var raw any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil
	}
	records := flattenObjectRecords(raw)
	keys := make([]model.UpstreamGatewayKey, 0)
	for _, record := range records {
		key, _ := stringField(record, "key", "api_key", "apiKey", "token", "value")
		key = strings.TrimSpace(key)
		importable := looksLikeImportableGatewayKey(key)
		maskedKey := safeMaskedGatewayKeyLabel(key)
		if importable {
			maskedKey = maskSecret(key)
		} else {
			key = ""
		}
		if maskedKey == "" {
			continue
		}
		name, _ := stringField(record, "name", "remark", "description", "id")
		allowedModels := splitLooseModelList(record["models"], record["allowed_models"], record["model_names"], record["model_limits"])
		groups := splitLooseStrings(record["groups"], record["group"], record["group_name"], record["group_id"], record["tags"], record["scopes"])
		status, _ := stringField(record, "status", "state")
		expiresAt, _ := stringField(record, "expires_at", "expire_at", "expired_at", "end_time", "expires")
		quota, _ := numberField(record, "quota", "remain_quota", "remaining", "limit")
		quotaUsed, _ := numberField(record, "quota_used", "used_quota", "used", "usage")
		keys = append(keys, model.UpstreamGatewayKey{
			Name:                name,
			Key:                 key,
			MaskedKey:           maskedKey,
			AllowedModels:       allowedModels,
			RequestCapabilities: requestCapabilitiesFromRecord(record),
			Groups:              groups,
			Status:              status,
			Quota:               quota,
			QuotaUsed:           quotaUsed,
			ExpiresAt:           expiresAt,
			Importable:          importable,
			SourceType:          model.ChannelKeySourceTypePaidMetered,
		})
	}
	return keys
}

func safeMaskedGatewayKeyLabel(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if strings.Contains(value, "*") {
		if strings.HasPrefix(strings.ToLower(value), "sk-") {
			return value
		}
		return "sk-" + value
	}
	if looksLikeImportableGatewayKey(value) {
		return maskSecret(value)
	}
	if strings.HasPrefix(strings.ToLower(value), "sk-") {
		return maskSecret(value)
	}
	return maskSecret("sk-" + value)
}

func looksLikeImportableGatewayKey(value string) bool {
	value = strings.TrimSpace(value)
	lower := strings.ToLower(value)
	if !strings.HasPrefix(lower, "sk-") {
		return false
	}
	if strings.Contains(value, "*") || strings.Contains(lower, "placeholder") || strings.Contains(lower, "your-token") {
		return false
	}
	return len(value) >= 12
}

func parseUpstreamSubscriptions(payload []byte) []model.UpstreamSubscription {
	var raw any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil
	}
	records := flattenObjectRecords(raw)
	out := make([]model.UpstreamSubscription, 0)
	seen := make(map[string]struct{})
	for _, record := range records {
		name, _ := stringField(record, "subscription", "subscription_name", "name", "package", "plan_name", "group_name")
		plan, _ := stringField(record, "plan", "plan_type", "tier", "level")
		status, _ := stringField(record, "status", "state")
		expiresAt, _ := stringField(record, "expires_at", "expire_at", "expired_at", "end_time", "expires")
		balance, _ := numberField(record, "balance", "remaining", "remain_quota", "quota", "monthly_limit_usd", "daily_limit_usd")
		source, _ := stringField(record, "source", "provider")
		if source == "" {
			source, _ = stringField(record, "platform")
		}
		if name == "" && plan == "" && status == "" && expiresAt == "" && balance == 0 {
			continue
		}
		key := strings.ToLower(strings.Join([]string{name, plan, status, expiresAt, fmt.Sprintf("%.6f", balance)}, "|"))
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, model.UpstreamSubscription{
			Name:      name,
			Plan:      plan,
			Status:    status,
			Balance:   balance,
			ExpiresAt: expiresAt,
			Source:    source,
		})
		if len(out) >= 20 {
			break
		}
	}
	return out
}

func requestCapabilitiesFromRecord(record map[string]any) []string {
	values := splitLooseStrings(
		record["request_capabilities"],
		record["capabilities"],
		record["request_types"],
		record["protocols"],
		record["api_types"],
		record["endpoints"],
	)
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		capability := model.NormalizeRequestCapability(value)
		if capability == "" {
			continue
		}
		if _, ok := seen[capability]; ok {
			continue
		}
		seen[capability] = struct{}{}
		out = append(out, capability)
	}
	sort.Strings(out)
	return out
}

func splitLooseStrings(values ...any) []string {
	out := make([]string, 0)
	for _, value := range values {
		switch typed := value.(type) {
		case string:
			out = append(out, strings.FieldsFunc(typed, func(r rune) bool {
				return r == ',' || r == '\n' || r == ';' || r == '|'
			})...)
		case float64:
			if typed == float64(int64(typed)) {
				out = append(out, strconv.FormatInt(int64(typed), 10))
			} else {
				out = append(out, strconv.FormatFloat(typed, 'f', -1, 64))
			}
		case int:
			out = append(out, strconv.Itoa(typed))
		case map[string]any:
			if value, ok := stringField(typed, "name", "group_name", "id", "value", "key"); ok {
				out = append(out, value)
			}
		case []any:
			for _, item := range typed {
				if s, ok := item.(string); ok {
					out = append(out, s)
					continue
				}
				if m, ok := item.(map[string]any); ok {
					if value, ok := stringField(m, "name", "group_name", "id", "value", "key"); ok {
						out = append(out, value)
					}
				}
			}
		}
	}
	return compactStrings(out)
}

func splitLooseModelList(values ...any) []string {
	out := make([]string, 0)
	for _, value := range values {
		switch typed := value.(type) {
		case string:
			out = append(out, strings.FieldsFunc(typed, func(r rune) bool {
				return r == ',' || r == '\n' || r == ';' || r == '|'
			})...)
		case map[string]any:
			if value, ok := stringField(typed, "id", "name", "model", "model_name"); ok {
				out = append(out, value)
			}
		case []any:
			for _, item := range typed {
				if s, ok := item.(string); ok {
					out = append(out, s)
					continue
				}
				if m, ok := item.(map[string]any); ok {
					if value, ok := stringField(m, "id", "name", "model", "model_name"); ok {
						out = append(out, value)
					}
				}
			}
		}
	}
	return normalizeModelNames(out)
}

func fetchSub2APIUsage(ctx context.Context, httpClient *http.Client, siteBase, token string) (model.NewAPITokenUsage, []string) {
	endpoint := strings.TrimRight(siteBase, "/") + "/v1/usage"
	for _, mode := range []string{"x-api-key", "bearer"} {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			continue
		}
		request.Header.Set("Accept", "application/json")
		if mode == "x-api-key" {
			request.Header.Set("x-api-key", token)
		} else {
			request.Header.Set("Authorization", "Bearer "+token)
		}
		response, err := httpClient.Do(request)
		if err != nil {
			continue
		}
		payload, readErr := io.ReadAll(io.LimitReader(response.Body, maxNewAPIInspectResponseBytes+1))
		_ = response.Body.Close()
		if readErr != nil || response.StatusCode < 200 || response.StatusCode >= 300 || int64(len(payload)) > maxNewAPIInspectResponseBytes {
			continue
		}
		usage, ok := parseSub2APIUsage(payload)
		if ok {
			return usage, nil
		}
	}
	return model.NewAPITokenUsage{Available: false}, []string{"sub2API /v1/usage 不可用"}
}

func parseSub2APIUsage(payload []byte) (model.NewAPITokenUsage, bool) {
	if usage, ok := parseNewAPITokenUsage(payload); ok {
		return usage, true
	}
	var raw any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return model.NewAPITokenUsage{}, false
	}
	for _, record := range flattenObjectRecords(raw) {
		usage := model.NewAPITokenUsage{Available: true}
		hasUsageField := false
		if value, ok := numberField(record, "remaining", "balance", "quota", "remain_quota"); ok {
			usage.RemainQuota = value
			hasUsageField = true
		}
		if value, ok := numberField(record, "used", "usage", "quota_used", "total_used"); ok {
			usage.UsedQuota = value
			hasUsageField = true
		}
		if value, ok := stringField(record, "mode", "status"); ok {
			usage.RawStatusText = value
			hasUsageField = true
		}
		if hasUsageField {
			return usage, true
		}
	}
	return model.NewAPITokenUsage{}, false
}

func normalizeModelNames(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, value)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return strings.ToLower(out[i]) < strings.ToLower(out[j])
	})
	return out
}

func modelsFromUpstreamKeys(keys []model.UpstreamGatewayKey) []string {
	out := make([]string, 0)
	for _, key := range keys {
		out = append(out, key.AllowedModels...)
	}
	return normalizeModelNames(out)
}

func dedupeUpstreamGatewayKeys(keys []model.UpstreamGatewayKey) []model.UpstreamGatewayKey {
	seen := make(map[string]struct{}, len(keys))
	out := make([]model.UpstreamGatewayKey, 0, len(keys))
	for _, key := range keys {
		identity := strings.ToLower(strings.TrimSpace(key.MaskedKey))
		if identity == "" {
			identity = strings.ToLower(strings.TrimSpace(key.Name))
		}
		if identity == "" {
			continue
		}
		if _, ok := seen[identity]; ok {
			continue
		}
		seen[identity] = struct{}{}
		out = append(out, key)
	}
	return out
}

func mergeUpstreamGroups(current, incoming []model.UpstreamGroup) []model.UpstreamGroup {
	seen := make(map[string]int, len(current)+len(incoming))
	out := make([]model.UpstreamGroup, 0, len(current)+len(incoming))
	add := func(group model.UpstreamGroup) {
		group.Name = strings.TrimSpace(group.Name)
		group.ID = strings.TrimSpace(group.ID)
		if group.Name == "" && group.ID == "" {
			return
		}
		group.Models = normalizeModelNames(group.Models)
		group.RequestCapabilities = compactStrings(group.RequestCapabilities)
		key := strings.ToLower(firstNonEmptyUpstreamValue(group.ID, group.Name))
		if idx, ok := seen[key]; ok {
			existing := out[idx]
			existing.Models = normalizeModelNames(append(existing.Models, group.Models...))
			existing.RequestCapabilities = compactStrings(append(existing.RequestCapabilities, group.RequestCapabilities...))
			if existing.Description == "" {
				existing.Description = group.Description
			}
			if existing.Platform == "" {
				existing.Platform = group.Platform
			}
			if existing.Status == "" {
				existing.Status = group.Status
			}
			if existing.RateMultiplier == 0 {
				existing.RateMultiplier = group.RateMultiplier
			}
			if existing.Source == "" {
				existing.Source = group.Source
			}
			out[idx] = existing
			return
		}
		seen[key] = len(out)
		out = append(out, group)
	}
	for _, group := range current {
		add(group)
	}
	for _, group := range incoming {
		add(group)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out
}

func mergeUpstreamSubscriptions(current, incoming []model.UpstreamSubscription) []model.UpstreamSubscription {
	seen := make(map[string]struct{}, len(current)+len(incoming))
	out := make([]model.UpstreamSubscription, 0, len(current)+len(incoming))
	add := func(item model.UpstreamSubscription) {
		key := strings.ToLower(strings.Join([]string{item.Name, item.Plan, item.Status, item.ExpiresAt, item.Source}, "|"))
		if key == "||||" && item.Balance == 0 {
			return
		}
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	for _, item := range current {
		add(item)
	}
	for _, item := range incoming {
		add(item)
	}
	return out
}

func compactStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func maskSecret(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) <= 10 {
		return value[:2] + "..." + value[len(value)-2:]
	}
	return value[:6] + "..." + value[len(value)-4:]
}

func buildUpstreamPriceCandidates(models []string, providerType, sourceLabel string, upstreamPrices map[string]model.UpstreamPriceCandidate) []model.UpstreamPriceCandidate {
	out := make([]model.UpstreamPriceCandidate, 0, len(models))
	for _, modelName := range models {
		info := model.LLMInfo{Name: modelName, CanonicalName: llmname.CanonicalModelName(modelName), UpstreamProviderType: providerType, UpstreamSource: sourceLabel}
		supported, policy, reason := model.InferCacheSupport(info)
		candidate := model.UpstreamPriceCandidate{
			Name:           modelName,
			CanonicalName:  info.CanonicalName,
			CacheSupported: &supported,
			CachePolicy:    string(policy),
			CacheReason:    reason,
		}
		if upstream, ok := upstreamPrices[strings.ToLower(strings.TrimSpace(modelName))]; ok {
			candidate.LLMPrice = upstream.LLMPrice
			candidate.PriceSource = upstream.PriceSource
			candidate.PriceMatchedKey = upstream.PriceMatchedKey
			candidate.Sources = upstream.Sources
			info.LLMPrice = upstream.LLMPrice
			supported, policy, reason = model.InferCacheSupport(info)
			candidate.CacheSupported = &supported
			candidate.CachePolicy = string(policy)
			candidate.CacheReason = reason
		}
		out = append(out, candidate)
	}
	return out
}

func ApplyUpstreamGateway(ctx context.Context, req model.UpstreamApplyRequest) (model.UpstreamApplyResult, error) {
	inspect, err := InspectUpstreamGateway(ctx, req.Inspect)
	if err != nil {
		return model.UpstreamApplyResult{}, err
	}
	if inspect.ModelCount == 0 {
		return model.UpstreamApplyResult{}, fmt.Errorf("upstream has no importable models")
	}
	keys := upstreamKeysForApply(inspect, req.Inspect, req.UpstreamSiteID)
	if len(keys) == 0 {
		return model.UpstreamApplyResult{}, fmt.Errorf("upstream has no importable gateway key")
	}
	enabled := true
	if req.EnableChannel != nil {
		enabled = *req.EnableChannel
	}
	overwriteModels := true
	if req.OverwriteModels != nil {
		overwriteModels = *req.OverwriteModels
	}
	appendKeys := true
	if req.AppendKeys != nil {
		appendKeys = *req.AppendKeys
	}
	modelList := strings.Join(inspect.Models, ",")
	baseURLs := []model.BaseUrl{{URL: inspect.APIBaseURL, Delay: 0}}

	var channel *model.Channel
	created := false
	if req.TargetChannelID > 0 {
		update := model.ChannelUpdateRequest{
			ID:       req.TargetChannelID,
			BaseUrls: &baseURLs,
		}
		if overwriteModels {
			update.CustomModel = &modelList
		}
		if req.EnableChannel != nil {
			update.Enabled = &enabled
		}
		if appendKeys {
			update.KeysToAdd = keys
		}
		if req.UpstreamSiteID > 0 {
			update.UpstreamSiteID = &req.UpstreamSiteID
			source := inspect.SourceLabel
			update.UpstreamSource = &source
		}
		channel, err = ChannelUpdate(&update, ctx)
		if err != nil {
			return model.UpstreamApplyResult{}, err
		}
	} else {
		name := strings.TrimSpace(req.ChannelName)
		if name == "" {
			name = defaultUpstreamChannelName(inspect)
		}
		newChannel := model.Channel{
			Name:              name,
			Type:              transformerOutbound.OutboundTypeOpenAIChat,
			Enabled:           enabled,
			KeyManagementMode: model.KeyManagementModeClassified,
			KeyRoutingPolicy:  model.KeyRoutingPolicyRoundRobin,
			BaseUrls:          baseURLs,
			Keys:              channelKeysFromAddRequests(keys),
			CustomModel:       modelList,
			AutoSync:          true,
			UpstreamSiteID:    req.UpstreamSiteID,
			UpstreamSource:    inspect.SourceLabel,
		}
		if err := ChannelCreate(&newChannel, ctx); err != nil {
			return model.UpstreamApplyResult{}, err
		}
		channel = &newChannel
		created = true
	}
	if err := upsertUpstreamLLMInfos(upstreamLLMInfos(inspect), ctx); err != nil {
		return model.UpstreamApplyResult{}, err
	}
	return model.UpstreamApplyResult{Channel: *channel, Summary: appliedChannelSummary(*channel), Inspect: inspect, Created: created}, nil
}

func appliedChannelSummary(channel model.Channel) model.UpstreamAppliedChannel {
	return model.UpstreamAppliedChannel{
		ID:          channel.ID,
		Name:        channel.Name,
		Type:        int(channel.Type),
		Enabled:     channel.Enabled,
		BaseUrls:    channel.BaseUrls,
		CustomModel: channel.CustomModel,
		KeyCount:    len(channel.Keys),
	}
}

func upstreamKeysForApply(inspect model.UpstreamInspectResult, req model.UpstreamInspectRequest, upstreamSiteID int) []model.ChannelKeyAddRequest {
	out := make([]model.ChannelKeyAddRequest, 0)
	modelList := strings.Join(inspect.Models, ",")
	capabilities := strings.Join(inspect.RequestCapabilities, ",")
	for _, key := range inspect.Keys {
		raw := strings.TrimSpace(key.Key)
		if raw == "" {
			continue
		}
		allowedModels := strings.Join(key.AllowedModels, ",")
		if allowedModels == "" {
			allowedModels = modelList
		}
		requestCapabilities := strings.Join(key.RequestCapabilities, ",")
		out = append(out, model.ChannelKeyAddRequest{
			Enabled:             true,
			ChannelKey:          raw,
			SourceType:          firstNonEmptyUpstreamValue(key.SourceType, model.ChannelKeySourceTypePaidMetered),
			Remark:              upstreamKeyRemark(key, inspect.ProviderType),
			AllowedModels:       allowedModels,
			RequestCapabilities: requestCapabilities,
			UpstreamSiteID:      upstreamSiteID,
			UpstreamKeyName:     key.Name,
		})
	}
	if len(out) > 0 {
		return out
	}
	raw := strings.TrimSpace(req.AccessKey)
	if raw == "" {
		return nil
	}
	return []model.ChannelKeyAddRequest{{
		Enabled:             true,
		ChannelKey:          raw,
		SourceType:          model.ChannelKeySourceTypePaidMetered,
		Remark:              inspect.ProviderType,
		AllowedModels:       modelList,
		RequestCapabilities: capabilities,
		UpstreamSiteID:      upstreamSiteID,
		UpstreamKeyName:     inspect.ProviderType,
	}}
}

func upstreamKeyRemark(key model.UpstreamGatewayKey, fallback string) string {
	parts := make([]string, 0, 3)
	if strings.TrimSpace(key.Name) != "" {
		parts = append(parts, strings.TrimSpace(key.Name))
	}
	if len(key.Groups) > 0 {
		parts = append(parts, "groups:"+strings.Join(compactStrings(key.Groups), "/"))
	}
	if strings.TrimSpace(key.Status) != "" {
		parts = append(parts, "status:"+strings.TrimSpace(key.Status))
	}
	if len(parts) == 0 {
		return fallback
	}
	remark := strings.Join(parts, " · ")
	if len(remark) > 180 {
		return remark[:180]
	}
	return remark
}

func channelKeysFromAddRequests(keys []model.ChannelKeyAddRequest) []model.ChannelKey {
	out := make([]model.ChannelKey, 0, len(keys))
	for _, key := range keys {
		out = append(out, model.ChannelKey{
			Enabled:             key.Enabled,
			ChannelKey:          key.ChannelKey,
			SourceType:          key.SourceType,
			Remark:              key.Remark,
			AllowedModels:       key.AllowedModels,
			RequestCapabilities: key.RequestCapabilities,
			UpstreamSiteID:      key.UpstreamSiteID,
			UpstreamKeyName:     key.UpstreamKeyName,
		})
	}
	return out
}

func upstreamLLMInfos(inspect model.UpstreamInspectResult) []model.LLMInfo {
	out := make([]model.LLMInfo, 0, len(inspect.PriceCandidates))
	for _, candidate := range inspect.PriceCandidates {
		out = append(out, model.LLMInfo{
			Name:                 candidate.Name,
			CanonicalName:        candidate.CanonicalName,
			LLMPrice:             candidate.LLMPrice,
			CachePolicy:          model.CachePolicy(candidate.CachePolicy),
			CacheReason:          candidate.CacheReason,
			UpstreamProviderType: inspect.ProviderType,
			UpstreamSource:       inspect.SourceLabel,
		})
	}
	return out
}

func upsertUpstreamLLMInfos(infos []model.LLMInfo, ctx context.Context) error {
	toCreate := make([]model.LLMInfo, 0, len(infos))
	for _, info := range infos {
		existing, err := LLMGet(strings.ToLower(strings.TrimSpace(info.Name)))
		if err != nil {
			toCreate = append(toCreate, info)
			continue
		}
		if existing.CachePolicy == "" || existing.CachePolicy == model.CachePolicyUnknown {
			existing.CachePolicy = info.CachePolicy
		}
		if strings.TrimSpace(existing.CacheReason) == "" {
			existing.CacheReason = info.CacheReason
		}
		if existing.LLMPrice.IsZero() && !info.LLMPrice.IsZero() {
			existing.LLMPrice = info.LLMPrice
		}
		existing.UpstreamProviderType = info.UpstreamProviderType
		existing.UpstreamSource = info.UpstreamSource
		if err := LLMUpdate(existing, ctx); err != nil {
			return err
		}
	}
	return LLMBatchCreate(toCreate, ctx)
}

func defaultUpstreamChannelName(inspect model.UpstreamInspectResult) string {
	parsed, err := url.Parse(inspect.BaseURL)
	host := strings.TrimSpace(inspect.BaseURL)
	if err == nil && strings.TrimSpace(parsed.Hostname()) != "" {
		host = parsed.Hostname()
	}
	prefix := "Upstream"
	switch inspect.ProviderType {
	case model.UpstreamProviderSub2API:
		prefix = "sub2API"
	case model.UpstreamProviderNewAPI:
		prefix = "New API"
	}
	return strings.TrimSpace(prefix + " - " + host)
}

func firstNonEmptyUpstreamValue(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func numberFromAny(value any) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case int:
		return float64(typed), true
	case json.Number:
		if parsed, err := typed.Float64(); err == nil {
			return parsed, true
		}
	case string:
		if parsed, err := strconv.ParseFloat(strings.TrimSpace(typed), 64); err == nil {
			return parsed, true
		}
	}
	return 0, false
}
