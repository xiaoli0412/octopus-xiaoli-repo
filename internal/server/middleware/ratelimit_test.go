package middleware

import (
	"context"
	"testing"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

func TestRateLimiterAllowRequestNoLimits(t *testing.T) {
	lim := newAPIKeyRateLimiter()
	key := model.APIKey{ID: 1}

	for i := 0; i < 100; i++ {
		if !lim.AllowRequest(key.ID, key) {
			t.Fatalf("request %d should be allowed when no limits set", i+1)
		}
	}
}

func TestRateLimiterAllowRequestRPM(t *testing.T) {
	lim := newAPIKeyRateLimiter()
	key := model.APIKey{ID: 2, RateLimitRPM: 3}

	for i := 0; i < 3; i++ {
		if !lim.AllowRequest(key.ID, key) {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
	if lim.AllowRequest(key.ID, key) {
		t.Fatal("4th request within the same minute should be rejected")
	}
}

func TestRateLimiterAllowRequestDaily(t *testing.T) {
	lim := newAPIKeyRateLimiter()
	key := model.APIKey{ID: 3, RateLimitDaily: 2}

	if !lim.AllowRequest(key.ID, key) {
		t.Fatal("1st daily request should be allowed")
	}
	if !lim.AllowRequest(key.ID, key) {
		t.Fatal("2nd daily request should be allowed")
	}
	if lim.AllowRequest(key.ID, key) {
		t.Fatal("3rd daily request should be rejected")
	}
}

func TestRateLimiterRecordTokensTPM(t *testing.T) {
	lim := newAPIKeyRateLimiter()
	key := model.APIKey{ID: 4, RateLimitTPM: 100}

	if !lim.RecordTokens(key.ID, 60, key) {
		t.Fatal("first token record should be allowed")
	}
	if !lim.RecordTokens(key.ID, 30, key) {
		t.Fatal("second token record within limit should be allowed")
	}
	if lim.RecordTokens(key.ID, 20, key) {
		t.Fatal("token record exceeding TPM should return false")
	}
}

func TestRateLimiterRecordTokensSlidingWindowExpires(t *testing.T) {
	lim := newAPIKeyRateLimiter()
	key := model.APIKey{ID: 5, RateLimitTPM: 50}

	if !lim.RecordTokens(key.ID, 50, key) {
		t.Fatal("first token record should be allowed")
	}
	if lim.RecordTokens(key.ID, 1, key) {
		t.Fatal("second token record should exceed limit")
	}

	// Manually backdate all hits so they fall outside the 1-minute window.
	lim.mu.Lock()
	window := lim.tpmWindows[key.ID]
	for i := range window.hits {
		window.hits[i].ts = time.Now().Add(-2 * time.Minute)
	}
	lim.mu.Unlock()

	if !lim.RecordTokens(key.ID, 50, key) {
		t.Fatal("after window expires, new tokens should be allowed")
	}
}

func TestRateLimiterSetRelayTokensInContext(t *testing.T) {
	ctx := context.Background()
	if totalTokensFromContext(ctx) != 0 {
		t.Fatal("empty context should return 0 tokens")
	}

	ctx = SetRelayTokensInContext(ctx, 150)
	if got := totalTokensFromContext(ctx); got != 150 {
		t.Fatalf("expected 150 tokens, got %d", got)
	}

	ctx = SetRelayTokensInContext(ctx, -10)
	if got := totalTokensFromContext(ctx); got != 150 {
		t.Fatalf("negative tokens should not overwrite context, got %d", got)
	}
}

func newAPIKeyRateLimiter() *apiKeyRateLimiter {
	return &apiKeyRateLimiter{
		rpmBuckets:    make(map[int]*tokenBucket),
		tpmWindows:    make(map[int]*slidingWindow),
		dailyCounters: make(map[int]*dailyWindow),
	}
}
