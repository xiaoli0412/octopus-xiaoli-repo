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
