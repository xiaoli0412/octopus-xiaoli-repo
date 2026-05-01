package relay

import (
	"context"
	"testing"
	"time"
)

func TestWaitForRetryDelayReturnsFalseWhenContextCanceled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	start := time.Now()
	if waitForRetryDelay(ctx, 50*time.Millisecond) {
		t.Fatal("waitForRetryDelay() = true, want false when context is canceled")
	}
	if elapsed := time.Since(start); elapsed >= 20*time.Millisecond {
		t.Fatalf("waitForRetryDelay() took %s after cancellation, want immediate return", elapsed)
	}
}

func TestWaitForRetryDelayReturnsTrueAfterDelay(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	start := time.Now()
	if !waitForRetryDelay(ctx, 10*time.Millisecond) {
		t.Fatal("waitForRetryDelay() = false, want true when delay elapses")
	}
	if elapsed := time.Since(start); elapsed < 10*time.Millisecond {
		t.Fatalf("waitForRetryDelay() returned too early after %s", elapsed)
	}
}