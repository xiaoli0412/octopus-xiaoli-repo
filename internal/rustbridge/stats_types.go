package rustbridge

// StatsMetrics mirrors the Rust StatsMetrics structure.
type StatsMetrics struct {
	InputToken     int64   `json:"input_token"`
	OutputToken    int64   `json:"output_token"`
	InputCost      float64 `json:"input_cost"`
	OutputCost     float64 `json:"output_cost"`
	WaitTime       int64   `json:"wait_time"`
	RequestSuccess int64   `json:"request_success"`
	RequestFailed  int64   `json:"request_failed"`
}
