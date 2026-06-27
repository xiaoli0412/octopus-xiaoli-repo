package rustbridge

// BalanceCandidate describes a single load balancer candidate.
type BalanceCandidate struct {
	ID           int64  `json:"id"`
	Weight       int64  `json:"weight"`
	Latency      int64  `json:"latency"`
	Priority     int64  `json:"priority"`
	Healthy      bool   `json:"healthy"`
	CircuitState string `json:"circuit_state"`
}

// BalanceResult is returned by BalanceSelect.
type BalanceResult struct {
	ID        int64 `json:"id"`
	NextIndex int   `json:"next_index"`
}
