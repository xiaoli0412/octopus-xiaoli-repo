package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/middleware"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/resp"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/router"
)

type antigravityOAuthSession struct {
	Status      string
	AccessToken string
	TokenType   string
	Scope       string
	Error       string
	CreatedAt   time.Time
}

var (
	antigravityOAuthSessions = map[string]*antigravityOAuthSession{}
	antigravityOAuthLock     sync.Mutex
	antigravityHTTPClient    = &http.Client{Timeout: 15 * time.Second}
)

const (
	maxAntigravityTokenResponseBytes int64 = 64 << 10
	antigravitySessionTTL                  = 15 * time.Minute
	antigravitySessionMaxEntries           = 512
)

type antigravityTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	Error       string `json:"error"`
	ErrorDesc   string `json:"error_description"`
}

func init() {
	router.NewGroupRouter("/api/v1/channel/antigravity").
		Use(middleware.Auth()).
		AddRoute(
			router.NewRoute("/oauth/start", http.MethodPost).
				Use(middleware.RequireJSON()).
				Handle(startAntigravityOAuth),
		).
		AddRoute(
			router.NewRoute("/oauth/poll", http.MethodPost).
				Use(middleware.RequireJSON()).
				Handle(pollAntigravityOAuth),
		)

	router.NewGroupRouter("/api/v1/channel/antigravity").
		AddRoute(
			router.NewRoute("/oauth/callback", http.MethodGet).
				Handle(antigravityOAuthCallback),
		)
}

func cleanupAntigravitySessions() {
	deadline := time.Now().Add(-antigravitySessionTTL)
	for state, session := range antigravityOAuthSessions {
		if session == nil || session.CreatedAt.Before(deadline) {
			delete(antigravityOAuthSessions, state)
		}
	}
}

func trimAntigravitySessionsLocked(currentState string) {
	for len(antigravityOAuthSessions) > antigravitySessionMaxEntries {
		oldestState := ""
		var oldestCreatedAt time.Time
		for state, session := range antigravityOAuthSessions {
			if state == currentState && len(antigravityOAuthSessions) > 1 {
				continue
			}
			candidate := time.Time{}
			if session != nil {
				candidate = session.CreatedAt
			}
			if oldestState == "" || candidate.Before(oldestCreatedAt) {
				oldestState = state
				oldestCreatedAt = candidate
			}
		}
		if oldestState == "" {
			break
		}
		delete(antigravityOAuthSessions, oldestState)
	}
}

func antigravityConfig() (clientID, clientSecret, authorizeURL, tokenURL, scope string, err error) {
	clientID = strings.TrimSpace(os.Getenv("OCTOPUS_ANTIGRAVITY_CLIENT_ID"))
	if clientID == "" {
		clientID = strings.TrimSpace(os.Getenv("ANTIGRAVITY_CLIENT_ID"))
	}
	clientSecret = strings.TrimSpace(os.Getenv("OCTOPUS_ANTIGRAVITY_CLIENT_SECRET"))
	if clientSecret == "" {
		clientSecret = strings.TrimSpace(os.Getenv("ANTIGRAVITY_CLIENT_SECRET"))
	}
	authorizeURL = strings.TrimSpace(os.Getenv("OCTOPUS_ANTIGRAVITY_AUTHORIZE_URL"))
	if authorizeURL == "" {
		authorizeURL = strings.TrimSpace(os.Getenv("ANTIGRAVITY_AUTHORIZE_URL"))
	}
	if authorizeURL == "" {
		authorizeURL = "https://accounts.google.com/o/oauth2/v2/auth"
	}
	tokenURL = strings.TrimSpace(os.Getenv("OCTOPUS_ANTIGRAVITY_TOKEN_URL"))
	if tokenURL == "" {
		tokenURL = strings.TrimSpace(os.Getenv("ANTIGRAVITY_TOKEN_URL"))
	}
	if tokenURL == "" {
		tokenURL = "https://oauth2.googleapis.com/token"
	}
	scope = strings.TrimSpace(os.Getenv("OCTOPUS_ANTIGRAVITY_SCOPE"))
	if scope == "" {
		scope = strings.TrimSpace(os.Getenv("ANTIGRAVITY_SCOPE"))
	}
	if scope == "" {
		scope = "https://www.googleapis.com/auth/cloud-platform https://www.googleapis.com/auth/userinfo.email https://www.googleapis.com/auth/userinfo.profile"
	}
	if err = validateAbsoluteHTTPURL(authorizeURL, "antigravity authorize url"); err != nil {
		return "", "", "", "", "", err
	}
	if err = validateAbsoluteHTTPURL(tokenURL, "antigravity token url"); err != nil {
		return "", "", "", "", "", err
	}
	return
}

func buildAPIPublicBase(c *gin.Context) string {
	scheme := "http"
	host := "localhost"
	if c != nil && c.Request != nil {
		if c.Request.TLS != nil {
			scheme = "https"
		}
		if requestHost := strings.TrimSpace(c.Request.Host); requestHost != "" {
			host = requestHost
		}
	}

	if apiBaseURL, err := op.SettingGetString(model.SettingKeyAPIBaseURL); err == nil {
		trimmed := strings.TrimSpace(apiBaseURL)
		if trimmed != "" {
			trimmed = strings.TrimRight(trimmed, "/")
			if validateAbsoluteHTTPURL(trimmed, "api base URL") == nil {
				parsed, _ := url.Parse(trimmed)
				if !hostIsLoopback(parsed.Host) || hostIsLoopback(host) {
					return trimmed
				}
			}
		}
	}

	return fmt.Sprintf("%s://%s", scheme, host)
}

func hostIsLoopback(host string) bool {
	host = strings.TrimSpace(host)
	if host == "" {
		return false
	}
	if extracted, _, err := net.SplitHostPort(host); err == nil {
		host = extracted
	}
	host = strings.Trim(host, "[]")
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func randomOAuthState() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func startAntigravityOAuth(c *gin.Context) {
	clientID, _, authorizeURL, _, scope, err := antigravityConfig()
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, "antigravity oauth configuration is invalid")
		return
	}
	if clientID == "" {
		resp.Error(c, http.StatusBadRequest, "antigravity oauth is not configured: missing client_id")
		return
	}
	state, err := randomOAuthState()
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, "failed to generate oauth state")
		return
	}
	redirectURI := buildAPIPublicBase(c) + "/api/v1/channel/antigravity/oauth/callback"

	u, err := url.Parse(authorizeURL)
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, "invalid antigravity authorize url")
		return
	}
	q := u.Query()
	q.Set("response_type", "code")
	q.Set("client_id", clientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("scope", scope)
	q.Set("state", state)
	u.RawQuery = q.Encode()

	antigravityOAuthLock.Lock()
	defer antigravityOAuthLock.Unlock()
	cleanupAntigravitySessions()
	antigravityOAuthSessions[state] = &antigravityOAuthSession{
		Status:    "pending",
		CreatedAt: time.Now(),
	}
	trimAntigravitySessionsLocked(state)

	resp.Success(c, gin.H{
		"state":    state,
		"auth_url": u.String(),
	})
}

type antigravityPollRequest struct {
	State string `json:"state"`
}

func pollAntigravityOAuth(c *gin.Context) {
	var req antigravityPollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	if strings.TrimSpace(req.State) == "" {
		resp.Error(c, http.StatusBadRequest, "state is required")
		return
	}

	antigravityOAuthLock.Lock()
	cleanupAntigravitySessions()
	session, ok := antigravityOAuthSessions[req.State]
	if !ok {
		antigravityOAuthLock.Unlock()
		resp.Error(c, http.StatusNotFound, "oauth session not found or expired")
		return
	}
	copySession := *session
	if session.Status == "authorized" || session.Status == "failed" {
		delete(antigravityOAuthSessions, req.State)
	}
	antigravityOAuthLock.Unlock()

	resp.Success(c, gin.H{
		"status":       copySession.Status,
		"access_token": copySession.AccessToken,
		"token_type":   copySession.TokenType,
		"scope":        copySession.Scope,
		"error":        copySession.Error,
	})
}

func antigravityOAuthCallback(c *gin.Context) {
	state, err := parseRequiredTrimmedStringQuery(c, "state")
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	antigravityOAuthLock.Lock()
	cleanupAntigravitySessions()
	session, ok := antigravityOAuthSessions[state]
	antigravityOAuthLock.Unlock()
	if !ok {
		c.String(http.StatusBadRequest, "session not found or expired")
		return
	}

	if errText, hasError := parseOptionalTrimmedStringQuery(c, "error"); hasError {
		msg, _ := parseOptionalTrimmedStringQuery(c, "error_description")
		if msg == "" {
			msg = errText
		}
		antigravityOAuthLock.Lock()
		session.Status = "failed"
		session.Error = msg
		antigravityOAuthLock.Unlock()
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<html><body><h3>Authorization failed</h3><p>You can close this window.</p></body></html>"))
		return
	}

	code, err := parseRequiredTrimmedStringQuery(c, "code")
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	clientID, clientSecret, _, tokenURL, _, err := antigravityConfig()
	if err != nil {
		antigravityOAuthLock.Lock()
		session.Status = "failed"
		session.Error = err.Error()
		antigravityOAuthLock.Unlock()
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<html><body><h3>Authorization failed</h3><p>Server oauth config is invalid.</p></body></html>"))
		return
	}
	if clientID == "" || clientSecret == "" {
		antigravityOAuthLock.Lock()
		session.Status = "failed"
		session.Error = "antigravity oauth is not configured: missing client_id/client_secret"
		antigravityOAuthLock.Unlock()
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<html><body><h3>Authorization failed</h3><p>Server oauth config is missing.</p></body></html>"))
		return
	}

	redirectURI := buildAPIPublicBase(c) + "/api/v1/channel/antigravity/oauth/callback"
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)

	httpReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		antigravityOAuthLock.Lock()
		session.Status = "failed"
		session.Error = err.Error()
		antigravityOAuthLock.Unlock()
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<html><body><h3>Authorization failed</h3><p>You can close this window.</p></body></html>"))
		return
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := antigravityHTTPClient.Do(httpReq)
	if err != nil {
		antigravityOAuthLock.Lock()
		session.Status = "failed"
		session.Error = err.Error()
		antigravityOAuthLock.Unlock()
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<html><body><h3>Authorization failed</h3><p>You can close this window.</p></body></html>"))
		return
	}
	defer httpResp.Body.Close()

	payload, tokenResp, err := readAntigravityTokenResponse(httpResp.Body)
	if err != nil {
		antigravityOAuthLock.Lock()
		session.Status = "failed"
		session.Error = err.Error()
		antigravityOAuthLock.Unlock()
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<html><body><h3>Authorization failed</h3><p>You can close this window.</p></body></html>"))
		return
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 || tokenResp.AccessToken == "" {
		errMsg := tokenResp.ErrorDesc
		if errMsg == "" {
			errMsg = tokenResp.Error
		}
		if errMsg == "" {
			errMsg = fmt.Sprintf("token exchange failed: %s", string(payload))
		}
		antigravityOAuthLock.Lock()
		session.Status = "failed"
		session.Error = errMsg
		antigravityOAuthLock.Unlock()
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<html><body><h3>Authorization failed</h3><p>You can close this window.</p></body></html>"))
		return
	}

	antigravityOAuthLock.Lock()
	session.Status = "authorized"
	session.AccessToken = tokenResp.AccessToken
	session.TokenType = tokenResp.TokenType
	session.Scope = tokenResp.Scope
	antigravityOAuthLock.Unlock()

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<html><body><h3>Authorization successful</h3><p>You can close this window and return to Octopus.</p><script>setTimeout(function(){window.close();}, 800);</script></body></html>"))
}

func readAntigravityTokenResponse(r io.Reader) ([]byte, antigravityTokenResponse, error) {
	var tokenResp antigravityTokenResponse
	payload, err := io.ReadAll(io.LimitReader(r, maxAntigravityTokenResponseBytes+1))
	if err != nil {
		return nil, tokenResp, err
	}
	if int64(len(payload)) > maxAntigravityTokenResponseBytes {
		return nil, tokenResp, fmt.Errorf("antigravity token response too large")
	}
	if len(payload) == 0 {
		return payload, tokenResp, nil
	}
	if err := json.Unmarshal(payload, &tokenResp); err != nil {
		return payload, tokenResp, fmt.Errorf("decode antigravity token response: %w", err)
	}
	return payload, tokenResp, nil
}
