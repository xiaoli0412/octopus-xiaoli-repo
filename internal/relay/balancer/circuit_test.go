package balancer

import (
	"sync"
	"testing"
	"time"
)

func resetCircuitBreakerState() {
	globalBreaker = sync.Map{}
}

func TestGetCooldownUsesDefaultExponentialBackoffWithCap(t *testing.T) {
	cases := []struct {
		tripCount int
		want      time.Duration
	}{
		{tripCount: 1, want: 60 * time.Second},
		{tripCount: 2, want: 120 * time.Second},
		{tripCount: 3, want: 240 * time.Second},
		{tripCount: 4, want: 480 * time.Second},
		{tripCount: 5, want: 600 * time.Second},
	}

	for _, tc := range cases {
		if got := GetCooldown(tc.tripCount); got != tc.want {
			t.Fatalf("GetCooldown(%d) = %v, want %v", tc.tripCount, got, tc.want)
		}
	}
}

func TestCircuitBreakerTripsRecoversAndResetsOnSuccess(t *testing.T) {
	resetCircuitBreakerState()

	const channelID = 10
	const keyID = 20
	const modelName = "gpt-4o"
	key := circuitKey(channelID, keyID, modelName)

	for i := 0; i < 4; i++ {
		RecordFailure(channelID, keyID, modelName)
		if tripped, _ := IsTripped(channelID, keyID, modelName); tripped {
			t.Fatalf("circuit should not trip before default threshold, failure #%d", i+1)
		}
	}

	RecordFailure(channelID, keyID, modelName)
	tripped, remaining := IsTripped(channelID, keyID, modelName)
	if !tripped {
		t.Fatal("circuit should be open after reaching default threshold")
	}
	if remaining <= 0 {
		t.Fatalf("remaining cooldown = %v, want > 0", remaining)
	}

	entryAny, ok := globalBreaker.Load(key)
	if !ok {
		t.Fatal("expected breaker entry to exist")
	}
	entry := entryAny.(*circuitEntry)
	entry.mu.Lock()
	entry.LastFailureTime = time.Now().Add(-GetCooldown(entry.TripCount) - time.Second)
	entry.mu.Unlock()

	tripped, remaining = IsTripped(channelID, keyID, modelName)
	if tripped {
		t.Fatalf("circuit should move to half-open after cooldown, remaining=%v", remaining)
	}

	tripped, _ = IsTripped(channelID, keyID, modelName)
	if !tripped {
		t.Fatal("half-open circuit should reject parallel probes")
	}

	RecordSuccess(channelID, keyID, modelName)
	tripped, remaining = IsTripped(channelID, keyID, modelName)
	if tripped || remaining != 0 {
		t.Fatalf("circuit should reset after success, got tripped=%t remaining=%v", tripped, remaining)
	}
}

func TestCircuitBreakerHalfOpenFailureReopensWithLongerCooldown(t *testing.T) {
	resetCircuitBreakerState()

	const channelID = 11
	const keyID = 21
	const modelName = "claude-3-5-sonnet"
	key := circuitKey(channelID, keyID, modelName)

	for i := 0; i < 5; i++ {
		RecordFailure(channelID, keyID, modelName)
	}

	entryAny, ok := globalBreaker.Load(key)
	if !ok {
		t.Fatal("expected breaker entry to exist")
	}
	entry := entryAny.(*circuitEntry)
	entry.mu.Lock()
	firstCooldown := GetCooldown(entry.TripCount)
	entry.LastFailureTime = time.Now().Add(-firstCooldown - time.Second)
	entry.mu.Unlock()

	if tripped, _ := IsTripped(channelID, keyID, modelName); tripped {
		t.Fatal("circuit should allow half-open probe after first cooldown")
	}

	RecordFailure(channelID, keyID, modelName)
	tripped, remaining := IsTripped(channelID, keyID, modelName)
	if !tripped {
		t.Fatal("circuit should reopen after half-open failure")
	}
	if remaining <= firstCooldown {
		t.Fatalf("remaining cooldown = %v, want > %v after second trip", remaining, firstCooldown)
	}
}

func TestSnapshotSummaryReportsOpenHalfOpenAndClosedBreakers(t *testing.T) {
	resetCircuitBreakerState()

	for i := 0; i < 5; i++ {
		RecordFailure(101, 201, "model-open")
		RecordFailure(102, 202, "model-half-open")
	}
	RecordFailure(103, 203, "model-closed")

	entryAny, ok := globalBreaker.Load(circuitKey(102, 202, "model-half-open"))
	if !ok {
		t.Fatal("expected half-open breaker entry to exist")
	}
	entry := entryAny.(*circuitEntry)
	entry.mu.Lock()
	entry.LastFailureTime = time.Now().Add(-GetCooldown(entry.TripCount) - time.Second)
	entry.mu.Unlock()

	if tripped, _ := IsTripped(102, 202, "model-half-open"); tripped {
		t.Fatal("breaker should move to half-open before snapshot")
	}

	summary := SnapshotSummary(time.Now())
	if summary.TrackedCount != 3 {
		t.Fatalf("TrackedCount = %d, want 3", summary.TrackedCount)
	}
	if summary.OpenCount != 1 {
		t.Fatalf("OpenCount = %d, want 1", summary.OpenCount)
	}
	if summary.HalfOpenCount != 1 {
		t.Fatalf("HalfOpenCount = %d, want 1", summary.HalfOpenCount)
	}
	if summary.ClosedCount != 1 {
		t.Fatalf("ClosedCount = %d, want 1", summary.ClosedCount)
	}
	if summary.MaxRemainingCooldownSec <= 0 {
		t.Fatalf("MaxRemainingCooldownSec = %d, want > 0", summary.MaxRemainingCooldownSec)
	}
}
