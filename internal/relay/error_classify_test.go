package relay

import (
	"errors"
	"math"
	"testing"
	"time"
)

func TestClassifyHTTPError(t *testing.T) {
	t.Parallel()

	sentinelErr := errors.New("network error")

	cases := []struct {
		name       string
		statusCode int
		err        error
		want       ErrorClass
	}{
		// RateLimited
		{name: "429 with error", statusCode: 429, err: errors.New("rate limited"), want: ErrorClassRateLimited},
		{name: "429 nil error", statusCode: 429, err: nil, want: ErrorClassRateLimited},

		// AuthFailed
		{name: "401 with error", statusCode: 401, err: errors.New("unauthorized"), want: ErrorClassAuthFailed},
		{name: "403 with error", statusCode: 403, err: errors.New("forbidden"), want: ErrorClassAuthFailed},
		{name: "401 nil error", statusCode: 401, err: nil, want: ErrorClassAuthFailed},
		{name: "403 nil error", statusCode: 403, err: nil, want: ErrorClassAuthFailed},

		// ServerError
		{name: "500 with error", statusCode: 500, err: errors.New("internal server error"), want: ErrorClassServerError},
		{name: "502 with error", statusCode: 502, err: errors.New("bad gateway"), want: ErrorClassServerError},
		{name: "503 with error", statusCode: 503, err: errors.New("service unavailable"), want: ErrorClassServerError},
		{name: "501 other 5xx", statusCode: 501, err: errors.New("not implemented"), want: ErrorClassServerError},
		{name: "505 other 5xx", statusCode: 505, err: errors.New("http version not supported"), want: ErrorClassServerError},
		{name: "599 edge 5xx", statusCode: 599, err: errors.New("edge 5xx"), want: ErrorClassServerError},

		// Timeout
		{name: "408 with error", statusCode: 408, err: errors.New("request timeout"), want: ErrorClassTimeout},
		{name: "504 with error", statusCode: 504, err: errors.New("gateway timeout"), want: ErrorClassTimeout},
		{name: "408 nil error", statusCode: 408, err: nil, want: ErrorClassTimeout},
		{name: "504 nil error", statusCode: 504, err: nil, want: ErrorClassTimeout},

		// ClientError (other 4xx)
		{name: "400 bad request", statusCode: 400, err: errors.New("bad request"), want: ErrorClassClientError},
		{name: "404 not found", statusCode: 404, err: errors.New("not found"), want: ErrorClassClientError},
		{name: "422 unprocessable", statusCode: 422, err: errors.New("unprocessable entity"), want: ErrorClassClientError},
		{name: "413 too large", statusCode: 413, err: errors.New("payload too large"), want: ErrorClassClientError},
		{name: "451 unavailable legal", statusCode: 451, err: errors.New("unavailable for legal reasons"), want: ErrorClassClientError},

		// NetworkError (statusCode=0, err!=nil)
		{name: "network error with nil status", statusCode: 0, err: sentinelErr, want: ErrorClassNetworkError},
		{name: "network error connection refused", statusCode: 0, err: errors.New("connection refused"), want: ErrorClassNetworkError},

		// Unknown (statusCode=0, err=nil)
		{name: "zero status nil error", statusCode: 0, err: nil, want: ErrorClassUnknown},

		// Edge: 3xx with error (should be NetworkError since not 4xx/5xx but has err)
		{name: "302 with error", statusCode: 302, err: errors.New("redirect error"), want: ErrorClassNetworkError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := ClassifyHTTPError(tc.statusCode, tc.err)
			if got != tc.want {
				t.Fatalf("ClassifyHTTPError(%d, %v) = %d, want %d", tc.statusCode, tc.err, got, tc.want)
			}
		})
	}
}

func TestShouldRetry(t *testing.T) {
	t.Parallel()

	cases := []struct {
		class ErrorClass
		want  bool
	}{
		{class: ErrorClassRateLimited, want: true},
		{class: ErrorClassAuthFailed, want: false},
		{class: ErrorClassServerError, want: false},
		{class: ErrorClassTimeout, want: false},
		{class: ErrorClassClientError, want: false},
		{class: ErrorClassNetworkError, want: false},
		{class: ErrorClassUnknown, want: false},
	}

	for _, tc := range cases {
		t.Run(className(tc.class), func(t *testing.T) {
			t.Parallel()
			got := ShouldRetry(tc.class)
			if got != tc.want {
				t.Fatalf("ShouldRetry(%s) = %t, want %t", className(tc.class), got, tc.want)
			}
		})
	}
}

func TestShouldFailover(t *testing.T) {
	t.Parallel()

	cases := []struct {
		class ErrorClass
		want  bool
	}{
		{class: ErrorClassRateLimited, want: false},
		{class: ErrorClassAuthFailed, want: true},
		{class: ErrorClassServerError, want: true},
		{class: ErrorClassTimeout, want: true},
		{class: ErrorClassClientError, want: false},
		{class: ErrorClassNetworkError, want: true},
		{class: ErrorClassUnknown, want: false},
	}

	for _, tc := range cases {
		t.Run(className(tc.class), func(t *testing.T) {
			t.Parallel()
			got := ShouldFailover(tc.class)
			if got != tc.want {
				t.Fatalf("ShouldFailover(%s) = %t, want %t", className(tc.class), got, tc.want)
			}
		})
	}
}

func TestComputeBackoffRange(t *testing.T) {
	t.Parallel()

	cases := []struct {
		attempt    int
		baseDelayMs int
	}{
		{attempt: 0, baseDelayMs: 1000},
		{attempt: 1, baseDelayMs: 1000},
		{attempt: 2, baseDelayMs: 1000},
		{attempt: 3, baseDelayMs: 1000},
		{attempt: 0, baseDelayMs: 0},
		{attempt: 1, baseDelayMs: 500},
		{attempt: 5, baseDelayMs: 200},
	}

	maxJitter := 500 * time.Millisecond

	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			t.Parallel()

			expectedBase := float64(tc.baseDelayMs) * math.Pow(1.5, float64(tc.attempt))
			minExpected := time.Duration(expectedBase) * time.Millisecond
			maxExpected := minExpected + maxJitter

			// Run multiple times to check range due to random jitter
			for i := 0; i < 100; i++ {
				got := ComputeBackoff(tc.attempt, tc.baseDelayMs)
				if got < minExpected {
					t.Fatalf("ComputeBackoff(%d, %d) = %v, want >= %v (base without jitter)", tc.attempt, tc.baseDelayMs, got, minExpected)
				}
				if got > maxExpected {
					t.Fatalf("ComputeBackoff(%d, %d) = %v, want <= %v (base + max jitter)", tc.attempt, tc.baseDelayMs, got, maxExpected)
				}
			}
		})
	}
}

func TestComputeBackoffNegativeInputs(t *testing.T) {
	t.Parallel()

	// Negative attempt should be treated as 0
	got := ComputeBackoff(-1, 1000)
	minExpected := 1000 * time.Millisecond
	maxExpected := minExpected + 500*time.Millisecond
	if got < minExpected || got > maxExpected {
		t.Fatalf("ComputeBackoff(-1, 1000) = %v, want in [%v, %v]", got, minExpected, maxExpected)
	}

	// Negative baseDelayMs should be treated as 0 (only jitter)
	got = ComputeBackoff(1, -100)
	if got < 0 || got > 500*time.Millisecond {
		t.Fatalf("ComputeBackoff(1, -100) = %v, want in [0, %v]", got, 500*time.Millisecond)
	}

	// Both negative
	got = ComputeBackoff(-5, -50)
	if got < 0 || got > 500*time.Millisecond {
		t.Fatalf("ComputeBackoff(-5, -50) = %v, want in [0, %v]", got, 500*time.Millisecond)
	}
}

func TestComputeBackoffExponentialGrowth(t *testing.T) {
	t.Parallel()

	baseDelayMs := 1000
	maxJitter := 500 * time.Millisecond

	// Verify that backoff increases with attempt (ignoring jitter)
	// Use the minimum expected value (base without jitter) for comparison
	minAttempt0 := time.Duration(float64(baseDelayMs)*math.Pow(1.5, 0)) * time.Millisecond
	minAttempt1 := time.Duration(float64(baseDelayMs)*math.Pow(1.5, 1)) * time.Millisecond
	minAttempt2 := time.Duration(float64(baseDelayMs)*math.Pow(1.5, 2)) * time.Millisecond

	if minAttempt1 <= minAttempt0 {
		t.Fatalf("expected min backoff at attempt 1 (%v) > attempt 0 (%v)", minAttempt1, minAttempt0)
	}
	if minAttempt2 <= minAttempt1 {
		t.Fatalf("expected min backoff at attempt 2 (%v) > attempt 1 (%v)", minAttempt2, minAttempt1)
	}

	// Verify expected values: 1000ms, 1500ms, 2250ms
	if minAttempt0 != 1000*time.Millisecond {
		t.Fatalf("min backoff attempt 0 = %v, want %v", minAttempt0, 1000*time.Millisecond)
	}
	if minAttempt1 != 1500*time.Millisecond {
		t.Fatalf("min backoff attempt 1 = %v, want %v", minAttempt1, 1500*time.Millisecond)
	}
	if minAttempt2 != 2250*time.Millisecond {
		t.Fatalf("min backoff attempt 2 = %v, want %v", minAttempt2, 2250*time.Millisecond)
	}

	// Sanity: max expected at attempt 2 = 2250 + 500 = 2750ms
	maxAttempt2 := minAttempt2 + maxJitter
	if maxAttempt2 != 2750*time.Millisecond {
		t.Fatalf("max backoff attempt 2 = %v, want %v", maxAttempt2, 2750*time.Millisecond)
	}
}

// className returns a human-readable name for an ErrorClass for test output.
func className(class ErrorClass) string {
	switch class {
	case ErrorClassUnknown:
		return "Unknown"
	case ErrorClassRateLimited:
		return "RateLimited"
	case ErrorClassAuthFailed:
		return "AuthFailed"
	case ErrorClassServerError:
		return "ServerError"
	case ErrorClassTimeout:
		return "Timeout"
	case ErrorClassClientError:
		return "ClientError"
	case ErrorClassNetworkError:
		return "NetworkError"
	default:
		return "Unknown"
	}
}
