package copilot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/httpx"
)

const (
	copilotTokenURL                       = "https://api.github.com/copilot_internal/v2/token"
	copilotTokenRefreshBuffer             = 30 * time.Second
	maxCopilotTokenResponseBytes    int64 = 64 << 10
	maxCopilotChatResponseBodyBytes int64 = 32 << 20
)

type cachedToken struct {
	Token     string
	ExpiresAt time.Time
}

type copilotTokenResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
	Message   string `json:"message,omitempty"`
}

var (
	tokenCache             = make(map[string]*cachedToken)
	tokenCacheMu           sync.RWMutex
	copilotTokenHTTPClient = &http.Client{Timeout: 15 * time.Second}
)

// ExchangeToken exchanges a GitHub OAuth access_token for a short-lived Copilot API token.
// Results are cached until near expiry.
func ExchangeToken(ctx context.Context, githubAccessToken string) (string, error) {
	tokenCacheMu.RLock()
	if cached, ok := tokenCache[githubAccessToken]; ok {
		if time.Now().Before(cached.ExpiresAt.Add(-copilotTokenRefreshBuffer)) {
			tokenCacheMu.RUnlock()
			return cached.Token, nil
		}
	}
	tokenCacheMu.RUnlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, copilotTokenURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create token exchange request: %w", err)
	}
	req.Header.Set("Authorization", "token "+githubAccessToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Editor-Version", "vscode/1.100.0")
	req.Header.Set("Editor-Plugin-Version", "copilot/1.300.0")

	resp, err := copilotTokenHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("copilot token exchange failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		payload, _ := httpx.ReadLimitedBody(resp.Body, maxCopilotTokenResponseBytes, "copilot token response too large")
		var errResp copilotTokenResponse
		_ = json.Unmarshal(payload, &errResp)
		if errResp.Message != "" {
			return "", fmt.Errorf("copilot token exchange failed (%d): %s", resp.StatusCode, errResp.Message)
		}
		return "", fmt.Errorf("copilot token exchange failed with status %d", resp.StatusCode)
	}

	payload, err := httpx.ReadLimitedBody(resp.Body, maxCopilotTokenResponseBytes, "copilot token response too large")
	if err != nil {
		return "", fmt.Errorf("failed to read copilot token response: %w", err)
	}

	var tokenResp copilotTokenResponse
	if err := json.Unmarshal(payload, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode copilot token response: %w", err)
	}

	if tokenResp.Token == "" {
		return "", fmt.Errorf("copilot token exchange returned empty token")
	}

	expiresAt := time.Unix(tokenResp.ExpiresAt, 0)
	tokenCacheMu.Lock()
	tokenCache[githubAccessToken] = &cachedToken{
		Token:     tokenResp.Token,
		ExpiresAt: expiresAt,
	}
	tokenCacheMu.Unlock()

	return tokenResp.Token, nil
}

// ChatOutbound wraps OpenAI Chat format but exchanges the GitHub OAuth token
// for a short-lived Copilot API token before each request.
type ChatOutbound struct{}

func (o *ChatOutbound) TransformRequest(ctx context.Context, request *model.InternalLLMRequest, baseUrl, key string) (*http.Request, error) {
	copilotToken, err := ExchangeToken(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("copilot token exchange failed: %w", err)
	}

	request.ClearHelpFields()

	for i := range request.Messages {
		if request.Messages[i].Role == "developer" {
			request.Messages[i].Role = "system"
		}
	}

	if request.Stream != nil && *request.Stream {
		if request.StreamOptions == nil {
			request.StreamOptions = &model.StreamOptions{IncludeUsage: true}
		} else if !request.StreamOptions.IncludeUsage {
			request.StreamOptions.IncludeUsage = true
		}
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+copilotToken)
	httpReq.Header.Set("Editor-Version", "vscode/1.100.0")
	httpReq.Header.Set("Editor-Plugin-Version", "copilot/1.300.0")
	httpReq.Header.Set("Copilot-Integration-Id", "vscode-chat")

	parsedUrl, err := url.Parse(strings.TrimSuffix(baseUrl, "/"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse base url: %w", err)
	}
	parsedUrl.Path = parsedUrl.Path + "/chat/completions"
	httpReq.URL = parsedUrl
	httpReq.Method = http.MethodPost
	return httpReq, nil
}

func (o *ChatOutbound) TransformResponse(ctx context.Context, response *http.Response) (*model.InternalLLMResponse, error) {
	body, err := httpx.ReadLimitedBody(response.Body, maxCopilotChatResponseBodyBytes, "copilot response body too large")
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if len(body) == 0 {
		return nil, fmt.Errorf("response body is empty")
	}

	if response.StatusCode >= http.StatusBadRequest {
		var errResp struct {
			Error model.ErrorDetail `json:"error"`
		}
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error.Message != "" {
			return nil, &model.ResponseError{
				StatusCode: response.StatusCode,
				Detail:     errResp.Error,
			}
		}
		return nil, fmt.Errorf("HTTP error %d: %s", response.StatusCode, string(body))
	}

	var resp model.InternalLLMResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return &resp, nil
}

func (o *ChatOutbound) TransformStream(ctx context.Context, eventData []byte) (*model.InternalLLMResponse, error) {
	if bytes.HasPrefix(eventData, []byte("[DONE]")) {
		return &model.InternalLLMResponse{
			Object: "[DONE]",
		}, nil
	}

	var errCheck struct {
		Error *model.ErrorDetail `json:"error"`
	}
	if err := json.Unmarshal(eventData, &errCheck); err == nil && errCheck.Error != nil {
		return nil, &model.ResponseError{
			Detail: *errCheck.Error,
		}
	}

	var resp model.InternalLLMResponse
	if err := json.Unmarshal(eventData, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal stream chunk: %w", err)
	}
	return &resp, nil
}
