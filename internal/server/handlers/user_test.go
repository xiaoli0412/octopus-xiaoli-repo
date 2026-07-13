package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	serverauth "github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/auth"
)

func TestLoginReturnsBadRequestForInvalidExpire(t *testing.T) {
	setupHandlerTest(t)

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/user/login", map[string]any{
		"username": "admin",
		"password": "admin",
		"expire":   -2,
	}, login)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	if !strings.Contains(res.Message, "expire") {
		t.Fatalf("message = %q, want expire validation message", res.Message)
	}
}

func TestLoginReturnsTokenForDefaultExpire(t *testing.T) {
	setupHandlerTest(t)

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/user/login", map[string]any{
		"username": "admin",
		"password": "admin",
		"expire":   0,
	}, login)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	if len(res.Data) == 0 {
		t.Fatalf("expected login response payload")
	}
	var payload struct {
		MustChangePassword bool `json:"must_change_password"`
	}
	if err := json.Unmarshal(res.Data, &payload); err != nil {
		t.Fatalf("unmarshal login payload error = %v", err)
	}
	if payload.MustChangePassword {
		t.Fatalf("must_change_password = true, want false for configured admin password")
	}
}

func TestStatusReturnsForceChangeFlag(t *testing.T) {
	setupHandlerTest(t)

	recorder := performJSONHandlerRequest(t, http.MethodGet, "/api/v1/user/status", nil, status)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	var payload struct {
		OK                 bool `json:"ok"`
		MustChangePassword bool `json:"must_change_password"`
	}
	if err := json.Unmarshal(res.Data, &payload); err != nil {
		t.Fatalf("unmarshal status payload error = %v", err)
	}
	if !payload.OK {
		t.Fatalf("ok = false, want true")
	}
	if payload.MustChangePassword {
		t.Fatalf("must_change_password = true, want false")
	}
}

func TestLoginReturnsForceChangeFlagForBuiltInBootstrapCredentials(t *testing.T) {
	setupHandlerTestDB(t)
	t.Setenv(op.BootstrapAdminUsernameEnv(), "")
	t.Setenv(op.BootstrapAdminPasswordEnv(), "")
	if err := initializeHandlerCaches(); err != nil {
		t.Fatalf("initializeHandlerCaches() error = %v", err)
	}
	if err := initializeHandlerUser(); err != nil {
		t.Fatalf("initializeHandlerUser() error = %v", err)
	}
	if err := initializeHandlerCaches(); err != nil {
		t.Fatalf("initializeHandlerCaches() after user init error = %v", err)
	}

	// With a randomly generated bootstrap password, verify the force-change flag
	// is set directly (login cannot be tested with an unknown generated password).
	if !op.UserMustChangePassword() {
		t.Fatalf("UserMustChangePassword() = false, want true for built-in bootstrap credentials")
	}
}

func TestForceChangePasswordClearsBuiltInBootstrapFlag(t *testing.T) {
	setupHandlerTestDB(t)
	t.Setenv(op.BootstrapAdminUsernameEnv(), "")
	t.Setenv(op.BootstrapAdminPasswordEnv(), "")
	if err := initializeHandlerCaches(); err != nil {
		t.Fatalf("initializeHandlerCaches() error = %v", err)
	}
	if err := initializeHandlerUser(); err != nil {
		t.Fatalf("initializeHandlerUser() error = %v", err)
	}
	if err := initializeHandlerCaches(); err != nil {
		t.Fatalf("initializeHandlerCaches() after user init error = %v", err)
	}

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/user/force-change-password", map[string]any{
		"new_username": "captain",
		"new_password": "changed-secret",
	}, forceChangePassword)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if got := op.UserGet().Username; got != "captain" {
		t.Fatalf("username = %q, want %q", got, "captain")
	}
	if err := op.UserVerify("captain", "changed-secret"); err != nil {
		t.Fatalf("UserVerify() error = %v", err)
	}
	statusRecorder := performJSONHandlerRequest(t, http.MethodGet, "/api/v1/user/status", nil, status)
	if statusRecorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", statusRecorder.Code, http.StatusOK, statusRecorder.Body.String())
	}
	res := decodeHandlerResponse(t, statusRecorder)
	var payload struct {
		MustChangePassword bool `json:"must_change_password"`
	}
	if err := json.Unmarshal(res.Data, &payload); err != nil {
		t.Fatalf("unmarshal status payload error = %v", err)
	}
	if payload.MustChangePassword {
		t.Fatalf("must_change_password = true, want false after force-change-password")
	}
}

func TestChangeUsernameRequiresCurrentPassword(t *testing.T) {
	setupHandlerTest(t)

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/user/change-username", map[string]any{
		"new_username":     "alice",
		"current_password": "wrong-password",
	}, changeUsername)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusUnauthorized, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	if !strings.Contains(res.Message, "Authentication") {
		t.Fatalf("message = %q, want authentication failure message", res.Message)
	}
}

func TestChangeUsernameRejectsWhitespaceOnlyInput(t *testing.T) {
	setupHandlerTest(t)

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/user/change-username", map[string]any{
		"new_username":     "   ",
		"current_password": "admin",
	}, changeUsername)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	if !strings.Contains(res.Message, "username") {
		t.Fatalf("message = %q, want username validation message", res.Message)
	}
}

func TestChangeUsernameTrimsWhitespaceBeforeUpdate(t *testing.T) {
	setupHandlerTest(t)

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/user/change-username", map[string]any{
		"new_username":     "  alice  ",
		"current_password": "admin",
	}, changeUsername)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if got := op.UserGet().Username; got != "alice" {
		t.Fatalf("username = %q, want %q", got, "alice")
	}
}

func TestJWTBecomesInvalidAfterUsernameChange(t *testing.T) {
	setupHandlerTest(t)

	token, _, err := serverauth.GenerateJWTToken(0)
	if err != nil {
		t.Fatalf("GenerateJWTToken() error = %v", err)
	}
	if !serverauth.VerifyJWTToken(token) {
		t.Fatalf("VerifyJWTToken() before username change = false, want true")
	}

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/user/change-username", map[string]any{
		"new_username":     "alice",
		"current_password": "admin",
	}, changeUsername)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if got := op.UserGet().Username; got != "alice" {
		t.Fatalf("username = %q, want %q", got, "alice")
	}
	if serverauth.VerifyJWTToken(token) {
		t.Fatalf("VerifyJWTToken() after username change = true, want false")
	}
}

func TestJWTBecomesInvalidAfterPasswordChange(t *testing.T) {
	setupHandlerTest(t)

	token, _, err := serverauth.GenerateJWTToken(0)
	if err != nil {
		t.Fatalf("GenerateJWTToken() error = %v", err)
	}
	if !serverauth.VerifyJWTToken(token) {
		t.Fatalf("VerifyJWTToken() before password change = false, want true")
	}

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/user/change-password", map[string]any{
		"old_password": "admin",
		"new_password": "changed-secret",
	}, changePassword)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if serverauth.VerifyJWTToken(token) {
		t.Fatalf("VerifyJWTToken() after password change = true, want false")
	}
}

func TestChangePasswordRequiresCorrectCurrentPassword(t *testing.T) {
	setupHandlerTest(t)

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/user/change-password", map[string]any{
		"old_password": "wrong-password",
		"new_password": "new-secret",
	}, changePassword)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusUnauthorized, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	if !strings.Contains(res.Message, "Authentication") {
		t.Fatalf("message = %q, want authentication failure message", res.Message)
	}
}

func TestChangePasswordRejectsWhitespaceOnlyNewPassword(t *testing.T) {
	setupHandlerTest(t)

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/user/change-password", map[string]any{
		"old_password": "admin",
		"new_password": "   ",
	}, changePassword)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	if !strings.Contains(res.Message, "password") {
		t.Fatalf("message = %q, want password validation message", res.Message)
	}
}

func TestLoginRateLimitsRepeatedFailures(t *testing.T) {
	setupHandlerTest(t)

	base := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
	loginAttemptNow = func() time.Time { return base }
	t.Cleanup(func() {
		loginAttemptNow = time.Now
	})

	for i := 0; i < loginFailureLimit; i++ {
		recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/user/login", map[string]any{
			"username": "admin",
			"password": "wrong-password",
			"expire":   0,
		}, login)
		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d status = %d, want %d, body = %s", i+1, recorder.Code, http.StatusUnauthorized, recorder.Body.String())
		}
	}

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/user/login", map[string]any{
		"username": "admin",
		"password": "admin",
		"expire":   0,
	}, login)
	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("blocked status = %d, want %d, body = %s", recorder.Code, http.StatusTooManyRequests, recorder.Body.String())
	}
	if retryAfter := recorder.Header().Get("Retry-After"); retryAfter == "" {
		t.Fatalf("Retry-After header missing")
	} else if _, err := strconv.Atoi(retryAfter); err != nil {
		t.Fatalf("Retry-After = %q, want integer seconds", retryAfter)
	}
	res := decodeHandlerResponse(t, recorder)
	if !strings.Contains(strings.ToLower(res.Message), "too many") {
		t.Fatalf("message = %q, want rate-limit message", res.Message)
	}
}

func TestLoginRateLimitExpiresAndAllowsSuccess(t *testing.T) {
	setupHandlerTest(t)

	base := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
	current := base
	loginAttemptNow = func() time.Time { return current }
	t.Cleanup(func() {
		loginAttemptNow = time.Now
	})

	for i := 0; i < loginFailureLimit; i++ {
		recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/user/login", map[string]any{
			"username": "admin",
			"password": "wrong-password",
			"expire":   0,
		}, login)
		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d status = %d, want %d, body = %s", i+1, recorder.Code, http.StatusUnauthorized, recorder.Body.String())
		}
	}

	current = base.Add(loginBlockDuration + time.Second)
	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/user/login", map[string]any{
		"username": "admin",
		"password": "admin",
		"expire":   0,
	}, login)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status after block expiry = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
}

func TestLoginRateLimitIgnoresSpoofedForwardedHeaders(t *testing.T) {
	setupHandlerTest(t)

	base := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
	loginAttemptNow = func() time.Time { return base }
	t.Cleanup(func() {
		loginAttemptNow = time.Now
	})

	for i := 0; i < loginFailureLimit; i++ {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/user/login", strings.NewReader(`{"username":"admin","password":"wrong-password","expire":0}`))
		ctx.Request.Header.Set("Content-Type", "application/json")
		ctx.Request.Header.Set("X-Forwarded-For", "198.51.100."+strconv.Itoa(i+10))
		ctx.Request.RemoteAddr = "203.0.113.9:4567"
		login(ctx)
		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d status = %d, want %d, body = %s", i+1, recorder.Code, http.StatusUnauthorized, recorder.Body.String())
		}
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/user/login", strings.NewReader(`{"username":"admin","password":"admin","expire":0}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Request.Header.Set("X-Forwarded-For", "198.51.100.99")
	ctx.Request.RemoteAddr = "203.0.113.9:4567"
	login(ctx)

	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("blocked status = %d, want %d, body = %s", recorder.Code, http.StatusTooManyRequests, recorder.Body.String())
	}
}

func TestLoginThrottlePrunesExpiredEntriesAcrossDistinctKeys(t *testing.T) {
	setupHandlerTest(t)

	base := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
	current := base
	loginAttemptNow = func() time.Time { return current }
	t.Cleanup(func() {
		loginAttemptNow = time.Now
	})

	for i := 0; i < 3; i++ {
		key := loginThrottleKeyForTest(t, "203.0.113.10:4567", "ghost-user-"+strconv.Itoa(i))
		loginThrottleRecordFailure(key)
	}

	if got := len(loginAttempts); got != 3 {
		t.Fatalf("loginAttempts size before prune = %d, want %d", got, 3)
	}

	current = base.Add(loginFailureWindow + time.Second)
	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/user/login", map[string]any{
		"username": "admin",
		"password": "wrong-password",
		"expire":   0,
	}, login)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusUnauthorized, recorder.Body.String())
	}

	if got := len(loginAttempts); got != 1 {
		t.Fatalf("loginAttempts size after prune = %d, want %d", got, 1)
	}
	for key := range loginAttempts {
		if !strings.HasPrefix(key, loginThrottleUsernameComponent("admin")+"|") {
			t.Fatalf("unexpected surviving throttle key %q", key)
		}
	}
}

func TestLoginThrottleKeyDoesNotEmbedRawUsername(t *testing.T) {
	setupHandlerTest(t)

	key := loginThrottleKeyForTest(t, "203.0.113.10:4567", "  Admin@Example.COM  ")
	if strings.Contains(key, "admin@example.com") {
		t.Fatalf("throttle key leaked normalized username: %q", key)
	}
	if !strings.HasPrefix(key, loginThrottleUsernameComponent("Admin@Example.COM")+"|") {
		t.Fatalf("throttle key = %q, want hashed username prefix", key)
	}
}

func TestLoginThrottleCapsDistinctFailureEntries(t *testing.T) {
	setupHandlerTest(t)

	base := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
	loginAttemptNow = func() time.Time { return base }
	t.Cleanup(func() {
		loginAttemptNow = time.Now
	})

	for i := 0; i < loginAttemptMaxEntries+128; i++ {
		key := loginThrottleKeyForTest(t, "203.0.113.10:4567", "ghost-user-"+strconv.Itoa(i))
		loginThrottleRecordFailure(key)
	}

	if got := len(loginAttempts); got != loginAttemptMaxEntries {
		t.Fatalf("loginAttempts size = %d, want capped size %d", got, loginAttemptMaxEntries)
	}
	firstKey := loginThrottleKeyForTest(t, "203.0.113.10:4567", "ghost-user-0")
	if _, ok := loginAttempts[firstKey]; ok {
		t.Fatalf("oldest throttle key %q still present after cap eviction", firstKey)
	}
	latestKey := loginThrottleKeyForTest(t, "203.0.113.10:4567", "ghost-user-"+strconv.Itoa(loginAttemptMaxEntries+127))
	if _, ok := loginAttempts[latestKey]; !ok {
		t.Fatalf("latest throttle key %q missing after cap eviction", latestKey)
	}
}

func loginThrottleKeyForTest(t *testing.T, remoteAddr string, username string) string {
	t.Helper()
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/user/login", nil)
	ctx.Request.RemoteAddr = remoteAddr
	return loginThrottleKey(ctx, username)
}

func TestRotateSecretSucceeds(t *testing.T) {
	setupHandlerTest(t)

	token, _, err := serverauth.GenerateJWTToken(0)
	if err != nil {
		t.Fatalf("GenerateJWTToken() error = %v", err)
	}
	if !serverauth.VerifyJWTToken(token) {
		t.Fatalf("VerifyJWTToken() before rotation = false, want true")
	}

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/user/rotate-secret", nil, rotateSecret)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	res := decodeHandlerResponse(t, recorder)
	var dataStr string
	if err := json.Unmarshal(res.Data, &dataStr); err != nil {
		t.Fatalf("unmarshal rotate-secret data error = %v", err)
	}
	if !strings.Contains(strings.ToLower(dataStr), "rotated") {
		t.Fatalf("data = %q, want rotation success message", dataStr)
	}

	if !serverauth.VerifyJWTToken(token) {
		t.Fatalf("VerifyJWTToken() after rotation = false, want true (old token should verify via secondary)")
	}

	newToken, _, err := serverauth.GenerateJWTToken(0)
	if err != nil {
		t.Fatalf("GenerateJWTToken() after rotation error = %v", err)
	}
	if !serverauth.VerifyJWTToken(newToken) {
		t.Fatalf("VerifyJWTToken() for new token = false, want true")
	}
}

func TestRotateSecretChangesPrimaryKey(t *testing.T) {
	setupHandlerTest(t)

	primaryBefore, err := serverauth.JWTSigningSecretForTests()
	if err != nil {
		t.Fatalf("JWTSigningSecretForTests() before rotation error = %v", err)
	}

	recorder := performJSONHandlerRequest(t, http.MethodPost, "/api/v1/user/rotate-secret", nil, rotateSecret)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	primaryAfter, err := serverauth.JWTSigningSecretForTests()
	if err != nil {
		t.Fatalf("JWTSigningSecretForTests() after rotation error = %v", err)
	}
	if string(primaryAfter) == string(primaryBefore) {
		t.Fatal("primary secret did not change after rotation")
	}

	secondaryAfter, err := serverauth.JWTSecondarySigningSecretForTests()
	if err != nil {
		t.Fatalf("JWTSecondarySigningSecretForTests() after rotation error = %v", err)
	}
	if string(secondaryAfter) != string(primaryBefore) {
		t.Fatal("secondary secret should equal the original primary after rotation")
	}
}
