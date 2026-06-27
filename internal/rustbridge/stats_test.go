package rustbridge

import (
	"math"
	"testing"
)

func TestMergeStatsHourly(t *testing.T) {
	existing := StatsMetrics{
		InputToken:     10,
		OutputToken:    5,
		InputCost:      1.0,
		OutputCost:     0.5,
		WaitTime:       100,
		RequestSuccess: 1,
		RequestFailed:  0,
	}
	delta := StatsMetrics{
		InputToken:     3,
		OutputToken:    2,
		InputCost:      0.3,
		OutputCost:     0.2,
		WaitTime:       50,
		RequestSuccess: 0,
		RequestFailed:  1,
	}
	merged, err := MergeStatsHourly(existing, delta)
	if err != nil {
		t.Fatalf("merge error: %v", err)
	}
	if merged.InputToken != 13 {
		t.Fatalf("input_token = %d, want 13", merged.InputToken)
	}
	if merged.OutputToken != 7 {
		t.Fatalf("output_token = %d, want 7", merged.OutputToken)
	}
	if merged.RequestFailed != 1 {
		t.Fatalf("request_failed = %d, want 1", merged.RequestFailed)
	}
}

func TestStatsQuantileMedian(t *testing.T) {
	values := []float64{1.0, 3.0, 2.0, 4.0, 5.0}
	q, err := StatsQuantile(values, 0.5)
	if err != nil {
		t.Fatalf("quantile error: %v", err)
	}
	if math.Abs(q-3.0) > 1e-9 {
		t.Fatalf("median = %v, want 3.0", q)
	}
}

func TestStatsQuantileEmpty(t *testing.T) {
	_, err := StatsQuantile([]float64{}, 0.5)
	if err == nil {
		t.Fatal("expected error for empty values")
	}
}

func TestStatsQuantileOutOfRange(t *testing.T) {
	_, err := StatsQuantile([]float64{1.0}, 1.5)
	if err == nil {
		t.Fatal("expected error for quantile > 1")
	}
}

func TestStatsEnvSwitch(t *testing.T) {
	if !Enabled() {
		t.Skip("rust backend not enabled")
	}
	existing := StatsMetrics{InputToken: 10}
	delta := StatsMetrics{InputToken: 5}
	rustMerged, err := MergeStatsHourly(existing, delta)
	if err != nil {
		t.Fatalf("rust merge error: %v", err)
	}
	t.Setenv(envDisableRustStats, "0")
	goMerged, err := MergeStatsHourly(existing, delta)
	if err != nil {
		t.Fatalf("go merge error: %v", err)
	}
	if rustMerged != goMerged {
		t.Fatalf("env switch changed result: rust=%+v go=%+v", rustMerged, goMerged)
	}
}
