package op

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	transformerOutbound "github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/outbound"
)

func TestHealthCheckProbeChannelSuccess(t *testing.T) {
	ctx := SetupOpTestDB(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// Use a sentinel delay of -1 so we can distinguish "not yet probed"
	// from a legitimately measured 0ms roundtrip on fast CI runners.
	channel := &model.Channel{
		Name:     "hc-success",
		Type:     transformerOutbound.OutboundTypeOpenAIChat,
		Enabled:  true,
		BaseUrls: []model.BaseUrl{{URL: srv.URL, Delay: -1}},
	}
	if err := ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}

	// Override circuit breaker shim to track invocation.
	var cbCalled int32
	oldCB := circuitBreakerResetForChannelFn
	circuitBreakerResetForChannelFn = func(channelID int) int {
		if channelID == channel.ID {
			atomic.StoreInt32(&cbCalled, 1)
		}
		return 0
	}
	defer func() { circuitBreakerResetForChannelFn = oldCB }()

	if err := probeChannel(*channel); err != nil {
		t.Fatalf("probeChannel() error = %v", err)
	}

	// Verify delay was updated in cache. A successful probe always records a
	// non-negative latency (>= 0ms); the sentinel -1 must have been replaced.
	updated, err := ChannelGet(channel.ID, ctx)
	if err != nil {
		t.Fatalf("ChannelGet() error = %v", err)
	}
	if len(updated.BaseUrls) != 1 {
		t.Fatalf("expected 1 base url, got %d", len(updated.BaseUrls))
	}
	if updated.BaseUrls[0].Delay < 0 {
		t.Fatalf("expected delay >= 0 after probe, got %d", updated.BaseUrls[0].Delay)
	}

	if atomic.LoadInt32(&cbCalled) == 0 {
		t.Fatal("circuit breaker reset shim was not called")
	}
}

func TestHealthCheckProbeChannelFailure(t *testing.T) {
	ctx := SetupOpTestDB(t)

	// Use a closed server URL to simulate connection failure.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	srv.Close()

	channel := &model.Channel{
		Name:     "hc-failure",
		Type:     transformerOutbound.OutboundTypeOpenAIChat,
		Enabled:  true,
		BaseUrls: []model.BaseUrl{{URL: srv.URL, Delay: 0}},
	}
	if err := ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}

	err := probeChannel(*channel)
	if err == nil {
		t.Fatal("probeChannel() expected error for closed server, got nil")
	}
	if !strings.Contains(err.Error(), "connectivity probe failed") {
		t.Fatalf("expected connectivity probe failed error, got: %v", err)
	}
}

func TestHealthCheckProbeChannelTimeout(t *testing.T) {
	ctx := SetupOpTestDB(t)

	// Set a very short probe timeout.
	if err := SettingSetString(model.SettingKeyHealthCheckTimeout, "50ms"); err != nil {
		t.Fatalf("SettingSetString(timeout) error = %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	channel := &model.Channel{
		Name:     "hc-timeout",
		Type:     transformerOutbound.OutboundTypeOpenAIChat,
		Enabled:  true,
		BaseUrls: []model.BaseUrl{{URL: srv.URL, Delay: 0}},
	}
	if err := ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}

	err := probeChannel(*channel)
	if err == nil {
		t.Fatal("probeChannel() expected timeout error, got nil")
	}
	// The error should mention context deadline exceeded or connectivity probe failed.
	lowerErr := strings.ToLower(err.Error())
	if !strings.Contains(lowerErr, "connectivity probe failed") && !strings.Contains(lowerErr, "deadline") {
		t.Fatalf("expected timeout-related error, got: %v", err)
	}
}

func TestHealthCheckProbeChannelNoBaseURL(t *testing.T) {
	ctx := SetupOpTestDB(t)

	channel := &model.Channel{
		Name:    "hc-no-url",
		Type:    transformerOutbound.OutboundTypeOpenAIChat,
		Enabled: true,
	}
	if err := ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}

	err := probeChannel(*channel)
	if err == nil {
		t.Fatal("probeChannel() expected error for no base URL, got nil")
	}
	if !strings.Contains(err.Error(), "no base URL") {
		t.Fatalf("expected no base URL error, got: %v", err)
	}
}

func TestHealthCheckCircuitBreakerRecovery(t *testing.T) {
	ctx := SetupOpTestDB(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	channel := &model.Channel{
		Name:     "hc-cb-recovery",
		Type:     transformerOutbound.OutboundTypeOpenAIChat,
		Enabled:  true,
		BaseUrls: []model.BaseUrl{{URL: srv.URL, Delay: 0}},
	}
	if err := ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}

	// Simulate tripped circuit breakers by using a mock that reports reset count.
	var resetCount int32
	oldCB := circuitBreakerResetForChannelFn
	circuitBreakerResetForChannelFn = func(channelID int) int {
		if channelID == channel.ID {
			atomic.StoreInt32(&resetCount, 2)
			return 2
		}
		return 0
	}
	defer func() { circuitBreakerResetForChannelFn = oldCB }()

	if err := probeChannel(*channel); err != nil {
		t.Fatalf("probeChannel() error = %v", err)
	}

	if got := atomic.LoadInt32(&resetCount); got != 2 {
		t.Fatalf("expected 2 circuit breakers reset, got %d", got)
	}
}

func TestHealthCheckCircuitBreakerShimNil(t *testing.T) {
	ctx := SetupOpTestDB(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	channel := &model.Channel{
		Name:     "hc-cb-nil",
		Type:     transformerOutbound.OutboundTypeOpenAIChat,
		Enabled:  true,
		BaseUrls: []model.BaseUrl{{URL: srv.URL, Delay: 0}},
	}
	if err := ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}

	// Set shim to nil to verify probeChannel handles missing shim gracefully.
	oldCB := circuitBreakerResetForChannelFn
	circuitBreakerResetForChannelFn = nil
	defer func() { circuitBreakerResetForChannelFn = oldCB }()

	if err := probeChannel(*channel); err != nil {
		t.Fatalf("probeChannel() with nil CB shim error = %v", err)
	}
}

func TestHealthCheckConfigInterval(t *testing.T) {
	SetupOpTestDB(t)

	// Default value.
	got := healthCheckInterval()
	if got != 5*time.Minute {
		t.Fatalf("default interval = %v, want 5m", got)
	}

	// Custom value.
	if err := SettingSetString(model.SettingKeyHealthCheckInterval, "3m"); err != nil {
		t.Fatalf("SettingSetString error = %v", err)
	}
	got = healthCheckInterval()
	if got != 3*time.Minute {
		t.Fatalf("custom interval = %v, want 3m", got)
	}

	// Invalid value falls back to default.
	if err := SettingSetString(model.SettingKeyHealthCheckInterval, "invalid"); err != nil {
		// SettingSetString validates, so "invalid" should fail validation.
		// That's expected — the runtime fallback handles unparseable values.
	}
	// Manually set cache to simulate an unparseable value bypassing validation.
	settingCache.Set(model.SettingKeyHealthCheckInterval, "not-a-duration")
	got = healthCheckInterval()
	if got != 5*time.Minute {
		t.Fatalf("invalid interval fallback = %v, want 5m", got)
	}

	// Empty value falls back to default.
	settingCache.Set(model.SettingKeyHealthCheckInterval, "")
	got = healthCheckInterval()
	if got != 5*time.Minute {
		t.Fatalf("empty interval fallback = %v, want 5m", got)
	}

	// Zero or negative duration falls back to default.
	settingCache.Set(model.SettingKeyHealthCheckInterval, "0s")
	got = healthCheckInterval()
	if got != 5*time.Minute {
		t.Fatalf("zero interval fallback = %v, want 5m", got)
	}
}

func TestHealthCheckConfigTimeout(t *testing.T) {
	SetupOpTestDB(t)

	// Default value.
	got := healthCheckTimeout()
	if got != 10*time.Second {
		t.Fatalf("default timeout = %v, want 10s", got)
	}

	// Custom value.
	if err := SettingSetString(model.SettingKeyHealthCheckTimeout, "30s"); err != nil {
		t.Fatalf("SettingSetString error = %v", err)
	}
	got = healthCheckTimeout()
	if got != 30*time.Second {
		t.Fatalf("custom timeout = %v, want 30s", got)
	}

	// Invalid value falls back to default.
	settingCache.Set(model.SettingKeyHealthCheckTimeout, "not-a-duration")
	got = healthCheckTimeout()
	if got != 10*time.Second {
		t.Fatalf("invalid timeout fallback = %v, want 10s", got)
	}
}

func TestHealthCheckConfigLLMProbe(t *testing.T) {
	SetupOpTestDB(t)

	// Default value: false.
	if healthCheckLLMProbeEnabled() {
		t.Fatal("default LLM probe should be false")
	}

	// Enable.
	if err := SettingSetString(model.SettingKeyHealthCheckLLMProbe, "true"); err != nil {
		t.Fatalf("SettingSetString error = %v", err)
	}
	if !healthCheckLLMProbeEnabled() {
		t.Fatal("LLM probe should be true after setting")
	}

	// Disable.
	if err := SettingSetString(model.SettingKeyHealthCheckLLMProbe, "false"); err != nil {
		t.Fatalf("SettingSetString error = %v", err)
	}
	if healthCheckLLMProbeEnabled() {
		t.Fatal("LLM probe should be false after setting")
	}
}

func TestHealthCheckStartHealthCheckTask(t *testing.T) {
	SetupOpTestDB(t)

	// When taskRegisterFn is nil, StartHealthCheckTask should not panic.
	oldFn := taskRegisterFn
	taskRegisterFn = nil
	defer func() { taskRegisterFn = oldFn }()

	// Should not panic.
	StartHealthCheckTask()

	// Now set a mock and verify registration.
	var registeredName string
	var registeredInterval time.Duration
	taskRegisterFn = func(name string, interval time.Duration, runOnStart bool, fn func()) {
		registeredName = name
		registeredInterval = interval
	}

	StartHealthCheckTask()

	if registeredName != healthCheckTaskName {
		t.Fatalf("registered name = %q, want %q", registeredName, healthCheckTaskName)
	}
	if registeredInterval != 5*time.Minute {
		t.Fatalf("registered interval = %v, want 5m", registeredInterval)
	}
}

func TestHealthCheckProbeChannelStatsUpdate(t *testing.T) {
	ctx := SetupOpTestDB(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// Disable circuit breaker shim to isolate stats testing.
	oldCB := circuitBreakerResetForChannelFn
	circuitBreakerResetForChannelFn = nil
	defer func() { circuitBreakerResetForChannelFn = oldCB }()

	channel := &model.Channel{
		Name:     "hc-stats",
		Type:     transformerOutbound.OutboundTypeOpenAIChat,
		Enabled:  true,
		BaseUrls: []model.BaseUrl{{URL: srv.URL, Delay: 0}},
	}
	if err := ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}

	// Successful probe should update stats.
	if err := probeChannel(*channel); err != nil {
		t.Fatalf("probeChannel() error = %v", err)
	}

	stats := StatsChannelGet(channel.ID)
	if stats.RequestSuccess != 1 {
		t.Fatalf("RequestSuccess = %d, want 1", stats.RequestSuccess)
	}
	if stats.RequestFailed != 0 {
		t.Fatalf("RequestFailed = %d, want 0", stats.RequestFailed)
	}

	// Failed probe should update failure stats.
	failedChannel := &model.Channel{
		Name:     "hc-stats-fail",
		Type:     transformerOutbound.OutboundTypeOpenAIChat,
		Enabled:  true,
		BaseUrls: []model.BaseUrl{{URL: "http://127.0.0.1:1", Delay: 0}},
	}
	if err := ChannelCreate(failedChannel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}

	_ = probeChannel(*failedChannel) // expected to fail

	failStats := StatsChannelGet(failedChannel.ID)
	if failStats.RequestFailed != 1 {
		t.Fatalf("RequestFailed = %d, want 1", failStats.RequestFailed)
	}
}

// TestHealthCheckUpdateChannelBaseUrlDelay verifies that only the probed URL's
// delay is updated, leaving other base URLs untouched.
func TestHealthCheckUpdateChannelBaseUrlDelay(t *testing.T) {
	ctx := SetupOpTestDB(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	otherURL := "http://other.example.com/v1"
	channel := &model.Channel{
		Name:     "hc-multi-url",
		Type:     transformerOutbound.OutboundTypeOpenAIChat,
		Enabled:  true,
		BaseUrls: []model.BaseUrl{{URL: srv.URL, Delay: 999}, {URL: otherURL, Delay: 888}},
	}
	if err := ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}

	// Update only the test server URL's delay.
	if err := updateChannelBaseUrlDelay(channel.ID, srv.URL, 42, ctx); err != nil {
		t.Fatalf("updateChannelBaseUrlDelay() error = %v", err)
	}

	updated, err := ChannelGet(channel.ID, ctx)
	if err != nil {
		t.Fatalf("ChannelGet() error = %v", err)
	}
	if len(updated.BaseUrls) != 2 {
		t.Fatalf("expected 2 base urls, got %d", len(updated.BaseUrls))
	}

	// Find the probed URL and verify delay.
	for _, bu := range updated.BaseUrls {
		if bu.URL == srv.URL {
			if bu.Delay != 42 {
				t.Fatalf("probed URL delay = %d, want 42", bu.Delay)
			}
		}
		if bu.URL == otherURL {
			if bu.Delay != 888 {
				t.Fatalf("other URL delay = %d, want 888 (should be untouched)", bu.Delay)
			}
		}
	}
}

// TestHealthCheckStartHealthCheckTaskWithCustomInterval ensures the task name
// constant is stable and registration works through the shim.
func TestHealthCheckStartHealthCheckTaskWithCustomInterval(t *testing.T) {
	SetupOpTestDB(t)

	if err := SettingSetString(model.SettingKeyHealthCheckInterval, "2m"); err != nil {
		t.Fatalf("SettingSetString error = %v", err)
	}

	var registeredInterval time.Duration
	oldFn := taskRegisterFn
	taskRegisterFn = func(name string, interval time.Duration, runOnStart bool, fn func()) {
		registeredInterval = interval
	}
	defer func() { taskRegisterFn = oldFn }()

	StartHealthCheckTask()

	if registeredInterval != 2*time.Minute {
		t.Fatalf("registered interval = %v, want 2m", registeredInterval)
	}
}
