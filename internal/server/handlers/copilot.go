package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/middleware"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/resp"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/router"
)

const maxCopilotOAuthResponseBytes int64 = 64 << 10

func copilotOAuthConfig() (clientID, scope, deviceCodeURL, accessTokenURL string, err error) {
	clientID = strings.TrimSpace(os.Getenv("OCTOPUS_COPILOT_CLIENT_ID"))
	if clientID == "" {
		clientID = strings.TrimSpace(os.Getenv("COPILOT_CLIENT_ID"))
	}
	if clientID == "" {
		clientID = "151ef1b1b0345b2351ca"
	}

	scope = strings.TrimSpace(os.Getenv("OCTOPUS_COPILOT_SCOPE"))
	if scope == "" {
		scope = strings.TrimSpace(os.Getenv("COPILOT_SCOPE"))
	}
	if scope == "" {
		scope = "copilot"
	}

	deviceCodeURL = strings.TrimSpace(os.Getenv("OCTOPUS_COPILOT_DEVICE_CODE_URL"))
	if deviceCodeURL == "" {
		deviceCodeURL = strings.TrimSpace(os.Getenv("COPILOT_DEVICE_CODE_URL"))
	}
	if deviceCodeURL == "" {
		deviceCodeURL = "https://github.com/login/device/code"
	}

	accessTokenURL = strings.TrimSpace(os.Getenv("OCTOPUS_COPILOT_ACCESS_TOKEN_URL"))
	if accessTokenURL == "" {
		accessTokenURL = strings.TrimSpace(os.Getenv("COPILOT_ACCESS_TOKEN_URL"))
	}
	if accessTokenURL == "" {
		accessTokenURL = "https://github.com/login/oauth/access_token"
	}
	if err = validateAbsoluteHTTPURL(deviceCodeURL, "copilot device code url"); err != nil {
		return "", "", "", "", err
	}
	if err = validateAbsoluteHTTPURL(accessTokenURL, "copilot access token url"); err != nil {
		return "", "", "", "", err
	}

	return
}

func init() {
	// device-code endpoint: no body needed
	router.NewGroupRouter("/api/v1/channel/copilot").
		Use(middleware.Auth()).
		AddRoute(
			router.NewRoute("/device-code", http.MethodPost).
				Use(middleware.RequireJSON()).
				Handle(copilotRequestDeviceCode),
		)
	// poll-token endpoint: requires JSON body
	router.NewGroupRouter("/api/v1/channel/copilot").
		Use(middleware.Auth()).
		Use(middleware.RequireJSON()).
		AddRoute(
			router.NewRoute("/poll-token", http.MethodPost).
				Handle(copilotPollToken),
		)
}

type copilotDeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

type copilotPollRequest struct {
	DeviceCode string `json:"device_code"`
}

type copilotPollResponse struct {
	AccessToken string `json:"access_token,omitempty"`
	TokenType   string `json:"token_type,omitempty"`
	Scope       string `json:"scope,omitempty"`
	Error       string `json:"error,omitempty"`
}

var copilotHTTPClient = &http.Client{Timeout: 15 * time.Second}

func copilotRequestDeviceCode(c *gin.Context) {
	clientID, scope, deviceCodeURL, _, err := copilotOAuthConfig()
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	bodyPayload, err := json.Marshal(map[string]string{
		"client_id": clientID,
		"scope":     scope,
	})
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	httpReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, deviceCodeURL, strings.NewReader(string(bodyPayload)))
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := copilotHTTPClient.Do(httpReq)
	if err != nil {
		resp.Error(c, http.StatusBadGateway, err.Error())
		return
	}
	defer httpResp.Body.Close()

	var result copilotDeviceCodeResponse
	if err := decodeCopilotOAuthResponse(httpResp.Body, &result); err != nil {
		resp.Error(c, http.StatusBadGateway, "failed to decode GitHub response: "+err.Error())
		return
	}

	if result.DeviceCode == "" {
		resp.Error(c, http.StatusBadGateway, "empty device_code from GitHub")
		return
	}

	resp.Success(c, result)
}

func copilotPollToken(c *gin.Context) {
	var req copilotPollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	if req.DeviceCode == "" {
		resp.Error(c, http.StatusBadRequest, "device_code is required")
		return
	}

	clientID, _, _, accessTokenURL, err := copilotOAuthConfig()
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	bodyPayload, err := json.Marshal(map[string]string{
		"client_id":   clientID,
		"device_code": req.DeviceCode,
		"grant_type":  "urn:ietf:params:oauth:grant-type:device_code",
	})
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	httpReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, accessTokenURL, strings.NewReader(string(bodyPayload)))
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := copilotHTTPClient.Do(httpReq)
	if err != nil {
		resp.Error(c, http.StatusBadGateway, err.Error())
		return
	}
	defer httpResp.Body.Close()

	var result copilotPollResponse
	if err := decodeCopilotOAuthResponse(httpResp.Body, &result); err != nil {
		resp.Error(c, http.StatusBadGateway, "failed to decode GitHub response: "+err.Error())
		return
	}

	resp.Success(c, result)
}

func decodeCopilotOAuthResponse(r io.Reader, target any) error {
	payload, err := io.ReadAll(io.LimitReader(r, maxCopilotOAuthResponseBytes+1))
	if err != nil {
		return err
	}
	if int64(len(payload)) > maxCopilotOAuthResponseBytes {
		return fmt.Errorf("copilot oauth response too large")
	}
	if err := json.Unmarshal(payload, target); err != nil {
		return fmt.Errorf("decode copilot oauth response: %w", err)
	}
	return nil
}
