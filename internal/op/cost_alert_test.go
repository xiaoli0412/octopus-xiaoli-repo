package op

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

// setCostAlertSettings writes the given webhook URL, thresholds and format
// into the setting cache (and backing DB) for tests.
func setCostAlertSettings(t *testing.T, webhookURL, thresholds, format string) {
	t.Helper()
	if err := SettingSetString(model.SettingKeyCostAlertWebhookURL, webhookURL); err != nil {
		t.Fatalf("set webhook URL: %v", err)
	}
	if err := SettingSetString(model.SettingKeyCostAlertThresholds, thresholds); err != nil {
		t.Fatalf("set thresholds: %v", err)
	}
	if err := SettingSetString(model.SettingKeyCostAlertFormat, format); err != nil {
		t.Fatalf("set format: %v", err)
	}
}

// newWebhookServer starts an httptest.Server that records every received body.
// Returns the server and a function to retrieve the received bodies.
func newWebhookServer(t *testing.T) (*httptest.Server, func() [][]byte) {
	t.Helper()
	var mu sync.Mutex
	var received [][]byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		mu.Lock()
		received = append(received, body)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	return srv, func() [][]byte {
		mu.Lock()
		defer mu.Unlock()
		out := make([][]byte, len(received))
		copy(out, received)
		return out
	}
}

func TestCheckCostAlert_ThresholdsFireOnce(t *testing.T) {
	ctx := SetupOpTestDB(t)
	_ = ctx
	ResetAllCostAlerts()

	srv, getBodies := newWebhookServer(t)
	setCostAlertSettings(t, srv.URL, "0.5,0.8,1.0", "generic")

	// 50% threshold
	CheckCostAlert(1, "test-key", 5.0, 10.0)
	// 80% threshold
	CheckCostAlert(1, "test-key", 8.0, 10.0)
	// 100% threshold
	CheckCostAlert(1, "test-key", 10.0, 10.0)
	// Repeat at 100% — should NOT fire again
	CheckCostAlert(1, "test-key", 12.0, 10.0)

	// Allow goroutine HTTP calls to complete
	time.Sleep(200 * time.Millisecond)

	bodies := getBodies()
	if len(bodies) != 3 {
		t.Fatalf("expected 3 webhook calls, got %d", len(bodies))
	}
}

func TestCheckCostAlert_DedupSameThreshold(t *testing.T) {
	SetupOpTestDB(t)
	ResetAllCostAlerts()

	var callCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	setCostAlertSettings(t, srv.URL, "0.5", "generic")

	// First call at 60% — should fire
	CheckCostAlert(2, "dedup-key", 6.0, 10.0)
	// Second call at 70% — should NOT fire (same threshold 0.5 already triggered)
	CheckCostAlert(2, "dedup-key", 7.0, 10.0)
	// Third call at 90% — should NOT fire
	CheckCostAlert(2, "dedup-key", 9.0, 10.0)

	time.Sleep(200 * time.Millisecond)

	if got := atomic.LoadInt32(&callCount); got != 1 {
		t.Fatalf("expected 1 webhook call, got %d", got)
	}
}

func TestResetCostAlerts_ClearsRecords(t *testing.T) {
	SetupOpTestDB(t)
	ResetAllCostAlerts()

	var callCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	setCostAlertSettings(t, srv.URL, "0.5", "generic")

	// Fire at 60%
	CheckCostAlert(3, "reset-key", 6.0, 10.0)
	time.Sleep(100 * time.Millisecond)
	if got := atomic.LoadInt32(&callCount); got != 1 {
		t.Fatalf("expected 1 call after first trigger, got %d", got)
	}

	// Reset — should allow re-trigger
	ResetCostAlerts(3)

	// Fire again at 60% — should fire once more
	CheckCostAlert(3, "reset-key", 6.0, 10.0)
	time.Sleep(100 * time.Millisecond)
	if got := atomic.LoadInt32(&callCount); got != 2 {
		t.Fatalf("expected 2 calls after reset, got %d", got)
	}
}

func TestCheckCostAlert_MaxCostZeroSkips(t *testing.T) {
	SetupOpTestDB(t)
	ResetAllCostAlerts()

	var callCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	setCostAlertSettings(t, srv.URL, "0.5", "generic")

	// maxCost = 0 → should skip entirely
	CheckCostAlert(4, "zero-max", 100.0, 0.0)
	time.Sleep(100 * time.Millisecond)

	if got := atomic.LoadInt32(&callCount); got != 0 {
		t.Fatalf("expected 0 calls for maxCost=0, got %d", got)
	}
}

func TestCheckCostAlert_EmptyWebhookURLSkips(t *testing.T) {
	SetupOpTestDB(t)
	ResetAllCostAlerts()

	// Set empty webhook URL — alerts disabled
	setCostAlertSettings(t, "", "0.5", "generic")

	// Should not panic and should not fire any webhook
	CheckCostAlert(5, "no-webhook", 6.0, 10.0)
	// If no panic occurred, the test passes
}

func TestCheckCostAlert_SlackFormat(t *testing.T) {
	SetupOpTestDB(t)
	ResetAllCostAlerts()

	srv, getBodies := newWebhookServer(t)
	setCostAlertSettings(t, srv.URL, "0.5", "slack")

	CheckCostAlert(10, "slack-key", 6.0, 10.0)
	time.Sleep(200 * time.Millisecond)

	bodies := getBodies()
	if len(bodies) != 1 {
		t.Fatalf("expected 1 webhook call, got %d", len(bodies))
	}

	var payload map[string]string
	if err := json.Unmarshal(bodies[0], &payload); err != nil {
		t.Fatalf("failed to unmarshal slack body: %v", err)
	}
	if _, ok := payload["text"]; !ok {
		t.Fatalf("slack body missing 'text' field: %s", string(bodies[0]))
	}
	if payload["text"] == "" {
		t.Fatalf("slack text field is empty")
	}
}

func TestCheckCostAlert_FeishuFormat(t *testing.T) {
	SetupOpTestDB(t)
	ResetAllCostAlerts()

	srv, getBodies := newWebhookServer(t)
	setCostAlertSettings(t, srv.URL, "0.5", "feishu")

	CheckCostAlert(11, "feishu-key", 6.0, 10.0)
	time.Sleep(200 * time.Millisecond)

	bodies := getBodies()
	if len(bodies) != 1 {
		t.Fatalf("expected 1 webhook call, got %d", len(bodies))
	}

	var payload map[string]any
	if err := json.Unmarshal(bodies[0], &payload); err != nil {
		t.Fatalf("failed to unmarshal feishu body: %v", err)
	}
	if payload["msg_type"] != "text" {
		t.Fatalf("feishu msg_type = %v, want text", payload["msg_type"])
	}
	content, ok := payload["content"].(map[string]any)
	if !ok {
		t.Fatalf("feishu content is not an object: %s", string(bodies[0]))
	}
	if _, ok := content["text"]; !ok {
		t.Fatalf("feishu content missing 'text' field")
	}
}

func TestCheckCostAlert_DingTalkFormat(t *testing.T) {
	SetupOpTestDB(t)
	ResetAllCostAlerts()

	srv, getBodies := newWebhookServer(t)
	setCostAlertSettings(t, srv.URL, "0.5", "dingtalk")

	CheckCostAlert(12, "dingtalk-key", 6.0, 10.0)
	time.Sleep(200 * time.Millisecond)

	bodies := getBodies()
	if len(bodies) != 1 {
		t.Fatalf("expected 1 webhook call, got %d", len(bodies))
	}

	var payload map[string]any
	if err := json.Unmarshal(bodies[0], &payload); err != nil {
		t.Fatalf("failed to unmarshal dingtalk body: %v", err)
	}
	if payload["msgtype"] != "text" {
		t.Fatalf("dingtalk msgtype = %v, want text", payload["msgtype"])
	}
	text, ok := payload["text"].(map[string]any)
	if !ok {
		t.Fatalf("dingtalk text is not an object: %s", string(bodies[0]))
	}
	if _, ok := text["content"]; !ok {
		t.Fatalf("dingtalk text missing 'content' field")
	}
}

func TestCheckCostAlert_GenericFormat(t *testing.T) {
	SetupOpTestDB(t)
	ResetAllCostAlerts()

	srv, getBodies := newWebhookServer(t)
	setCostAlertSettings(t, srv.URL, "0.5", "generic")

	CheckCostAlert(13, "generic-key", 6.0, 10.0)
	time.Sleep(200 * time.Millisecond)

	bodies := getBodies()
	if len(bodies) != 1 {
		t.Fatalf("expected 1 webhook call, got %d", len(bodies))
	}

	var payload AlertPayload
	if err := json.Unmarshal(bodies[0], &payload); err != nil {
		t.Fatalf("failed to unmarshal generic body: %v", err)
	}
	if payload.APIKeyID != 13 {
		t.Fatalf("generic api_key_id = %d, want 13", payload.APIKeyID)
	}
	if payload.APIKeyName != "generic-key" {
		t.Fatalf("generic api_key_name = %q, want generic-key", payload.APIKeyName)
	}
	if payload.MaxCost != 10.0 {
		t.Fatalf("generic max_cost = %f, want 10.0", payload.MaxCost)
	}
	if payload.Threshold != 0.5 {
		t.Fatalf("generic threshold = %f, want 0.5", payload.Threshold)
	}
}

func TestParseCostAlertThresholds(t *testing.T) {
	tests := []struct {
		input string
		want  []float64
	}{
		{"0.5,0.8,1.0", []float64{0.5, 0.8, 1.0}},
		{" 0.5 , 0.8 , 1.0 ", []float64{0.5, 0.8, 1.0}},
		{"", nil},
		{"invalid", nil},
		{"0,3,-1", nil},
		{"1.5", []float64{1.5}},
	}
	for _, tt := range tests {
		got := parseCostAlertThresholds(tt.input)
		if len(got) != len(tt.want) {
			t.Errorf("parseCostAlertThresholds(%q) = %v, want %v", tt.input, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("parseCostAlertThresholds(%q)[%d] = %f, want %f", tt.input, i, got[i], tt.want[i])
			}
		}
	}
}

func TestBuildWebhookBody_Formats(t *testing.T) {
	payload := AlertPayload{
		APIKeyName:  "test",
		APIKeyID:    1,
		CurrentCost: 5.0,
		MaxCost:     10.0,
		Threshold:   0.5,
		Percentage:  0.5,
		Timestamp:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	// Slack
	slackBody, err := buildWebhookBody("slack", payload)
	if err != nil {
		t.Fatalf("buildWebhookBody(slack) error: %v", err)
	}
	var slack map[string]string
	if err := json.Unmarshal(slackBody, &slack); err != nil {
		t.Fatalf("slack unmarshal error: %v", err)
	}
	if slack["text"] == "" {
		t.Fatalf("slack text is empty")
	}

	// Feishu
	feishuBody, err := buildWebhookBody("feishu", payload)
	if err != nil {
		t.Fatalf("buildWebhookBody(feishu) error: %v", err)
	}
	var feishu map[string]any
	if err := json.Unmarshal(feishuBody, &feishu); err != nil {
		t.Fatalf("feishu unmarshal error: %v", err)
	}
	if feishu["msg_type"] != "text" {
		t.Fatalf("feishu msg_type = %v, want text", feishu["msg_type"])
	}

	// DingTalk
	dingBody, err := buildWebhookBody("dingtalk", payload)
	if err != nil {
		t.Fatalf("buildWebhookBody(dingtalk) error: %v", err)
	}
	var ding map[string]any
	if err := json.Unmarshal(dingBody, &ding); err != nil {
		t.Fatalf("dingtalk unmarshal error: %v", err)
	}
	if ding["msgtype"] != "text" {
		t.Fatalf("dingtalk msgtype = %v, want text", ding["msgtype"])
	}

	// Generic
	genericBody, err := buildWebhookBody("generic", payload)
	if err != nil {
		t.Fatalf("buildWebhookBody(generic) error: %v", err)
	}
	var generic AlertPayload
	if err := json.Unmarshal(genericBody, &generic); err != nil {
		t.Fatalf("generic unmarshal error: %v", err)
	}
	if generic.APIKeyID != 1 {
		t.Fatalf("generic api_key_id = %d, want 1", generic.APIKeyID)
	}
}

func TestResetCostAlerts_DoesNotAffectOtherKeys(t *testing.T) {
	SetupOpTestDB(t)
	ResetAllCostAlerts()

	var callCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	setCostAlertSettings(t, srv.URL, "0.5", "generic")

	// Fire for key 20 and key 21
	CheckCostAlert(20, "key-20", 6.0, 10.0)
	CheckCostAlert(21, "key-21", 6.0, 10.0)
	time.Sleep(100 * time.Millisecond)
	if got := atomic.LoadInt32(&callCount); got != 2 {
		t.Fatalf("expected 2 calls, got %d", got)
	}

	// Reset only key 20
	ResetCostAlerts(20)

	// Fire again — key 20 should fire, key 21 should NOT
	CheckCostAlert(20, "key-20", 6.0, 10.0)
	CheckCostAlert(21, "key-21", 6.0, 10.0)
	time.Sleep(100 * time.Millisecond)
	if got := atomic.LoadInt32(&callCount); got != 3 {
		t.Fatalf("expected 3 calls after selective reset, got %d", got)
	}
}
