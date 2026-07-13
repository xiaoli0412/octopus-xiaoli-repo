package rustbridge

import (
	"testing"
)

func testCandidates() []BalanceCandidate {
	return []BalanceCandidate{
		{ID: 1, Weight: 2, Latency: 10, Priority: 1, Healthy: true, CircuitState: "closed"},
		{ID: 2, Weight: 1, Latency: 20, Priority: 2, Healthy: true, CircuitState: "closed"},
		{ID: 3, Weight: 1, Latency: 5, Priority: 0, Healthy: false, CircuitState: "open"},
	}
}

func TestBalanceSelectWeighted(t *testing.T) {
	cands := testCandidates()
	r0, err := BalanceSelect(cands, "weighted", 0)
	if err != nil {
		t.Fatalf("balance error: %v", err)
	}
	if r0.ID != 1 {
		t.Fatalf("expected id 1 at idx 0, got %d", r0.ID)
	}
	r1, err := BalanceSelect(cands, "weighted", 1)
	if err != nil {
		t.Fatalf("balance error: %v", err)
	}
	if r1.ID != 1 {
		t.Fatalf("expected id 1 at idx 1, got %d", r1.ID)
	}
	r2, err := BalanceSelect(cands, "weighted", 2)
	if err != nil {
		t.Fatalf("balance error: %v", err)
	}
	if r2.ID != 2 {
		t.Fatalf("expected id 2 at idx 2, got %d", r2.ID)
	}
}

func TestBalanceSelectRoundRobin(t *testing.T) {
	cands := testCandidates()
	r, err := BalanceSelect(cands, "round_robin", 0)
	if err != nil {
		t.Fatalf("balance error: %v", err)
	}
	if r.ID != 1 {
		t.Fatalf("expected id 1, got %d", r.ID)
	}
}

func TestBalanceSelectFailover(t *testing.T) {
	cands := testCandidates()
	r, err := BalanceSelect(cands, "failover", 0)
	if err != nil {
		t.Fatalf("balance error: %v", err)
	}
	if r.ID != 1 {
		t.Fatalf("expected id 1, got %d", r.ID)
	}
}

func TestBalanceSelectRandom(t *testing.T) {
	cands := testCandidates()
	r, err := BalanceSelect(cands, "random", 0)
	if err != nil {
		t.Fatalf("balance error: %v", err)
	}
	if r.ID != 1 && r.ID != 2 {
		t.Fatalf("unexpected id %d", r.ID)
	}
}

func TestBalanceSelectUnhealthyExcluded(t *testing.T) {
	cands := testCandidates()
	cands[0].Healthy = false
	r, err := BalanceSelect(cands, "failover", 0)
	if err != nil {
		t.Fatalf("balance error: %v", err)
	}
	if r.ID != 2 {
		t.Fatalf("expected id 2 after excluding unhealthy, got %d", r.ID)
	}
}

func TestBalanceSelectLeastLatency(t *testing.T) {
	cands := []BalanceCandidate{
		{ID: 1, Weight: 1, Latency: 100, Priority: 1, Healthy: true, CircuitState: "closed"},
		{ID: 2, Weight: 1, Latency: 50, Priority: 2, Healthy: true, CircuitState: "closed"},
		{ID: 3, Weight: 1, Latency: 200, Priority: 0, Healthy: true, CircuitState: "closed"},
	}
	r, err := BalanceSelect(cands, "least_latency", 0)
	if err != nil {
		t.Fatalf("balance error: %v", err)
	}
	if r.ID != 2 {
		t.Fatalf("expected id 2 (lowest latency 50ms), got %d", r.ID)
	}
}

func TestBalanceSelectLeastLatencyExcludesOpen(t *testing.T) {
	cands := []BalanceCandidate{
		{ID: 1, Weight: 1, Latency: 10, Priority: 1, Healthy: false, CircuitState: "open"},
		{ID: 2, Weight: 1, Latency: 50, Priority: 2, Healthy: true, CircuitState: "closed"},
	}
	r, err := BalanceSelect(cands, "least_latency", 0)
	if err != nil {
		t.Fatalf("balance error: %v", err)
	}
	if r.ID != 2 {
		t.Fatalf("expected id 2 (open circuit excluded), got %d", r.ID)
	}
}

func TestBalanceSelectHealthAware(t *testing.T) {
	cands := []BalanceCandidate{
		{ID: 1, Weight: 1, Latency: 100, Priority: 1, Healthy: true, CircuitState: "closed"},
		{ID: 2, Weight: 1, Latency: 50, Priority: 2, Healthy: true, CircuitState: "half-open"},
		{ID: 3, Weight: 1, Latency: 200, Priority: 0, Healthy: true, CircuitState: "closed"},
	}
	r, err := BalanceSelect(cands, "health_aware", 0)
	if err != nil {
		t.Fatalf("balance error: %v", err)
	}
	// closed (rank 0) beats half-open (rank 1); among closed, lowest latency wins
	if r.ID != 1 {
		t.Fatalf("expected id 1 (closed, latency 100ms), got %d", r.ID)
	}
}

func TestBalanceSelectHealthAwarePrefersLowerLatency(t *testing.T) {
	cands := []BalanceCandidate{
		{ID: 1, Weight: 1, Latency: 200, Priority: 1, Healthy: true, CircuitState: "closed"},
		{ID: 2, Weight: 1, Latency: 80, Priority: 2, Healthy: true, CircuitState: "closed"},
	}
	r, err := BalanceSelect(cands, "health_aware", 0)
	if err != nil {
		t.Fatalf("balance error: %v", err)
	}
	if r.ID != 2 {
		t.Fatalf("expected id 2 (lower latency 80ms), got %d", r.ID)
	}
}

func TestBalanceEnvSwitch(t *testing.T) {
	if !Enabled() {
		t.Skip("rust backend not enabled")
	}
	cands := testCandidates()
	rustResult, err := BalanceSelect(cands, "weighted", 0)
	if err != nil {
		t.Fatalf("rust balance error: %v", err)
	}
	t.Setenv(envDisableRustBalancer, "0")
	goResult, err := BalanceSelect(cands, "weighted", 0)
	if err != nil {
		t.Fatalf("go balance error: %v", err)
	}
	if rustResult != goResult {
		t.Fatalf("env switch changed result: rust=%+v go=%+v", rustResult, goResult)
	}
}
