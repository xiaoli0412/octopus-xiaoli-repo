//go:build !rust
// +build !rust

package rustbridge

import (
	"errors"
	"sort"
)

// MergeStatsHourly merges delta into existing using pure Go.
func MergeStatsHourly(existing, delta StatsMetrics) (StatsMetrics, error) {
	existing.InputToken += delta.InputToken
	existing.OutputToken += delta.OutputToken
	existing.InputCost += delta.InputCost
	existing.OutputCost += delta.OutputCost
	existing.WaitTime += delta.WaitTime
	existing.RequestSuccess += delta.RequestSuccess
	existing.RequestFailed += delta.RequestFailed
	return existing, nil
}

// StatsQuantile computes a quantile over values using pure Go.
func StatsQuantile(values []float64, quantile float64) (float64, error) {
	if quantile < 0.0 || quantile > 1.0 {
		return 0, errors.New("quantile out of range")
	}
	if len(values) == 0 {
		return 0, errors.New("empty values")
	}
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)
	n := len(sorted)
	idxF := quantile * float64(n-1)
	lower := int(idxF)
	upper := lower
	if idxF != float64(lower) {
		upper = lower + 1
	}
	if upper >= n {
		upper = n - 1
	}
	frac := idxF - float64(lower)
	return sorted[lower]*(1.0-frac) + sorted[upper]*frac, nil
}
