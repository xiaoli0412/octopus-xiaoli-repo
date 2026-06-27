// test-upstream-mock is a local mock server covering the three upstream protocols
// used by Octopus: new API, sub2API and OpenAI Compatible. It is intended for
// regression tests in internal/op/upstream_gateway_test.go and can also be run
// standalone for manual spot checks.
//
// Usage:
//
//	go run ./scripts/test-upstream-mock
//
// The server prints its base URL on startup. It exposes three protocol prefixes:
//   - /newapi/*     – new API compatible endpoints
//   - /sub2api/*    – sub2API compatible endpoints
//   - /openai/*     – OpenAI Compatible endpoints
//
// In addition the root path (/) also serves the same endpoints so that callers
// can use the server URL directly as a base URL without a prefix.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	newAPIToken   = "newapi-management-token"
	sub2APIToken  = "sub2-management-token"
	openAIKey     = "sk-openai-compatible"
	cookieToken   = "cookie-session-token"
	adminUsername = "admin@example.com"
	adminPassword = "admin123"
)

func main() {
	port := flag.String("port", "18088", "listen port")
	flag.Parse()

	mux := http.NewServeMux()
	registerCommon(mux, "")
	registerCommon(mux, "/newapi")
	registerSub2API(mux, "")
	registerSub2API(mux, "/sub2api")
	registerOpenAI(mux, "")
	registerOpenAI(mux, "/openai")

	addr := ":" + *port
	fmt.Printf("test-upstream-mock listening on http://127.0.0.1%s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func registerCommon(mux *http.ServeMux, prefix string) {
	base := prefix
	mux.HandleFunc(base+"/api/user/me", newAPIUserMe)
	mux.HandleFunc(base+"/api/user/self", newAPIUserSelf)
	mux.HandleFunc(base+"/api/user/models", newAPIUserModels)
	mux.HandleFunc(base+"/api/user/self/models", newAPIUserModels)
	mux.HandleFunc(base+"/api/user/self/groups", newAPIUserGroups)
	mux.HandleFunc(base+"/api/token", newAPITokenList)
	mux.HandleFunc(base+"/api/token/", newAPITokenList)
	mux.HandleFunc(base+"/api/token/search", newAPITokenList)
	mux.HandleFunc(base+"/api/pricing", newAPIPricing)
	mux.HandleFunc(base+"/api/user/checkin", newAPICheckin)
	mux.HandleFunc(base+"/v1/models", openAIModels)
}

func registerSub2API(mux *http.ServeMux, prefix string) {
	base := prefix
	mux.HandleFunc(base+"/api/profile", sub2APIProfile)
	mux.HandleFunc(base+"/api/user/profile", sub2APIProfile)
	mux.HandleFunc(base+"/api/me", sub2APIProfile)
	mux.HandleFunc(base+"/api/v1/user/profile", sub2APIProfile)
	mux.HandleFunc(base+"/api/v1/auth/me", sub2APIProfile)
	mux.HandleFunc(base+"/api/v1/auth/login", sub2APILogin)
	mux.HandleFunc(base+"/api/keys", sub2APIKeys)
	mux.HandleFunc(base+"/api/api-keys", sub2APIKeys)
	mux.HandleFunc(base+"/api/v1/keys", sub2APIKeys)
	mux.HandleFunc(base+"/api/v1/api-keys", sub2APIKeys)
	mux.HandleFunc(base+"/api/subscriptions", sub2APISubscriptions)
	mux.HandleFunc(base+"/api/user/subscriptions", sub2APISubscriptions)
	mux.HandleFunc(base+"/api/v1/subscriptions", sub2APISubscriptions)
	mux.HandleFunc(base+"/api/v1/subscriptions/active", sub2APISubscriptions)
	mux.HandleFunc(base+"/api/groups", sub2APIGroups)
	mux.HandleFunc(base+"/api/user/groups", sub2APIGroups)
	mux.HandleFunc(base+"/api/v1/groups", sub2APIGroups)
	mux.HandleFunc(base+"/api/v1/groups/available", sub2APIGroups)
	mux.HandleFunc(base+"/api/v1/pricing", sub2APIPricing)
	mux.HandleFunc(base+"/api/v1/user/checkin", sub2APICheckin)
}

func registerOpenAI(mux *http.ServeMux, prefix string) {
	base := prefix
	mux.HandleFunc(base+"/v1/models", openAIModels)
	mux.HandleFunc(base+"/v1/chat/completions", openAIChatCompletions)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func bearer(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}

func xAPIKey(r *http.Request) string {
	return r.Header.Get("x-api-key")
}

func cookieSession(r *http.Request) string {
	for _, c := range strings.Split(r.Header.Get("Cookie"), ";") {
		parts := strings.SplitN(strings.TrimSpace(c), "=", 2)
		if len(parts) == 2 && parts[0] == "session" {
			return parts[1]
		}
	}
	return ""
}

func requireBearer(w http.ResponseWriter, r *http.Request, token string) bool {
	if bearer(r) == token {
		return true
	}
	writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
	return false
}

func requireAPIKey(w http.ResponseWriter, r *http.Request) bool {
	if bearer(r) == openAIKey || xAPIKey(r) == openAIKey {
		return true
	}
	writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
	return false
}

func newAPIUserMe(w http.ResponseWriter, r *http.Request) {
	if !requireBearer(w, r, newAPIToken) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"id":           "newapi-user-1",
			"username":     "newapi-owner",
			"display_name": "Owner",
			"role":         1,
			"status":       1,
		},
	})
}

func newAPIUserSelf(w http.ResponseWriter, r *http.Request) {
	if !requireBearer(w, r, newAPIToken) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"username":     "newapi-owner",
			"group":        "default",
			"used_quota":   12.5,
			"remain_quota": 987.5,
		},
	})
}

func newAPIUserModels(w http.ResponseWriter, r *http.Request) {
	if !requireBearer(w, r, newAPIToken) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": []map[string]any{
			{"id": "gpt-4o"},
			{"id": "gpt-4o-mini"},
			{"id": "text-embedding-3-small"},
		},
	})
}

func newAPIUserGroups(w http.ResponseWriter, r *http.Request) {
	if !requireBearer(w, r, newAPIToken) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": []map[string]any{
			{"id": 1, "name": "default", "models": []string{"gpt-4o"}, "rate_multiplier": 1.0, "status": "active"},
		},
	})
}

func newAPITokenList(w http.ResponseWriter, r *http.Request) {
	if !requireBearer(w, r, newAPIToken) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": []map[string]any{
			{"id": 1, "name": "default", "key": "sk-newapi-importable", "models": "gpt-4o", "status": 1},
		},
	})
}

func newAPIPricing(w http.ResponseWriter, r *http.Request) {
	if !requireBearer(w, r, newAPIToken) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"model_ratio": map[string]float64{
			"gpt-4o":      0.5,
			"gpt-4o-mini": 0.05,
		},
		"completion_ratio": map[string]float64{
			"gpt-4o":      2.0,
			"gpt-4o-mini": 2.0,
		},
	})
}

func newAPICheckin(w http.ResponseWriter, r *http.Request) {
	if bearer(r) != newAPIToken && cookieSession(r) != cookieToken {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    map[string]any{"quota": 3.3},
		"message": "签到成功",
	})
}

func sub2APILogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		return
	}
	var body map[string]string
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body["username"] == adminUsername && body["password"] == adminPassword {
		writeJSON(w, http.StatusOK, map[string]any{
			"data": map[string]any{
				"access_token": sub2APIToken,
				"user_id":      "sub2-user-1",
			},
		})
		return
	}
	writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "invalid credentials"})
}

func sub2APIProfile(w http.ResponseWriter, r *http.Request) {
	if !requireBearer(w, r, sub2APIToken) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"id":           "sub2-user-1",
			"email":        adminUsername,
			"quota":        500.0,
			"used":         50.0,
			"subscription": "pro",
		},
	})
}

func sub2APIKeys(w http.ResponseWriter, r *http.Request) {
	if !requireBearer(w, r, sub2APIToken) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": []map[string]any{
			{"id": 1, "name": "main", "key": "sk-sub2-importable", "models": []string{"gpt-4o"}, "groups": []string{"vip"}, "status": "enabled"},
		},
	})
}

func sub2APISubscriptions(w http.ResponseWriter, r *http.Request) {
	if !requireBearer(w, r, sub2APIToken) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": []map[string]any{
			{"name": "pro", "plan": "monthly", "status": "active", "balance": 450.0, "expires_at": "2026-12-31"},
		},
	})
}

func sub2APIGroups(w http.ResponseWriter, r *http.Request) {
	if !requireBearer(w, r, sub2APIToken) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": []map[string]any{
			{"id": 1, "name": "vip", "models": []string{"gpt-4o"}, "rate_multiplier": 0.8, "status": "active"},
		},
	})
}

func sub2APIPricing(w http.ResponseWriter, r *http.Request) {
	if !requireBearer(w, r, sub2APIToken) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": []map[string]any{
			{"model": "gpt-4o", "input_price": 0.6, "output_price": 1.2},
		},
	})
}

func sub2APICheckin(w http.ResponseWriter, r *http.Request) {
	if bearer(r) != sub2APIToken && cookieSession(r) != cookieToken {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"amount":  7.7,
		"message": "sub2 checkin reward",
	})
}

func openAIModels(w http.ResponseWriter, r *http.Request) {
	if !requireAPIKey(w, r) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"object": "list",
		"data": []map[string]any{
			{"id": "gpt-3.5-turbo", "object": "model"},
			{"id": "gpt-4o", "object": "model"},
		},
	})
}

func openAIChatCompletions(w http.ResponseWriter, r *http.Request) {
	if !requireAPIKey(w, r) {
		return
	}
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		return
	}
	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)
	modelName := "gpt-3.5-turbo"
	if m, ok := body["model"].(string); ok && m != "" {
		modelName = m
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":      "chatcmpl-test",
		"object":  "chat.completion",
		"model":   modelName,
		"choices": []map[string]any{{"index": 0, "message": map[string]string{"role": "assistant", "content": "hi"}, "finish_reason": "stop"}},
		"usage":   map[string]int{"prompt_tokens": 2, "completion_tokens": 1, "total_tokens": 3},
	})
}

// Test helpers used by upstream_gateway_test.go.
// They are no-ops at runtime but kept here so the file is self-contained.
func init() {
	if v := os.Getenv("TEST_UPSTREAM_MOCK_PORT"); v != "" {
		_ = v
	}
}
