//go:build rust
// +build rust

package rustbridge

/*
#cgo windows LDFLAGS: -L${SRCDIR}/../../rust/core/target/release -loctopus_core -lws2_32 -luserenv -ladvapi32 -lntdll -lbcrypt
#cgo linux LDFLAGS: -L${SRCDIR}/../../rust/core/target/release -loctopus_core -lm
#cgo darwin LDFLAGS: -L${SRCDIR}/../../rust/core/target/release -loctopus_core -framework Security -framework CoreFoundation
#include <stdlib.h>
extern int octopus_stats_merge_hourly(const char* existing_json, const char* delta_json, char** out);
extern int octopus_stats_quantile(const char* values_json, double quantile, char** out);
extern void octopus_free_string(char* s);
*/
import "C"
import (
	"encoding/json"
	"errors"
	"sort"
	"unsafe"
)

// MergeStatsHourly merges delta into existing using Rust FFI.
// Set OCTOPUS_RUST_STATS=0 to use the Go fallback.
func MergeStatsHourly(existing, delta StatsMetrics) (StatsMetrics, error) {
	if !envEnabled(envDisableRustStats) {
		return mergeStatsHourlyGo(existing, delta)
	}
	return mergeStatsHourlyRust(existing, delta)
}

// StatsQuantile computes a quantile over values using Rust FFI.
func StatsQuantile(values []float64, quantile float64) (float64, error) {
	if !envEnabled(envDisableRustStats) {
		return statsQuantileGo(values, quantile)
	}
	return statsQuantileRust(values, quantile)
}

func mergeStatsHourlyRust(existing, delta StatsMetrics) (StatsMetrics, error) {
	ejson, err := json.Marshal(existing)
	if err != nil {
		return StatsMetrics{}, err
	}
	djson, err := json.Marshal(delta)
	if err != nil {
		return StatsMetrics{}, err
	}
	cejson := C.CString(string(ejson))
	defer C.free(unsafe.Pointer(cejson))
	cdjson := C.CString(string(djson))
	defer C.free(unsafe.Pointer(cdjson))
	var out *C.char
	if rc := C.octopus_stats_merge_hourly(cejson, cdjson, &out); rc != 0 {
		return StatsMetrics{}, errors.New("rust stats_merge_hourly failed")
	}
	outStr := takeCString(out)
	var merged StatsMetrics
	if err := json.Unmarshal([]byte(outStr), &merged); err != nil {
		return StatsMetrics{}, err
	}
	return merged, nil
}

func statsQuantileRust(values []float64, quantile float64) (float64, error) {
	vjson, err := json.Marshal(values)
	if err != nil {
		return 0, err
	}
	cvjson := C.CString(string(vjson))
	defer C.free(unsafe.Pointer(cvjson))
	var out *C.char
	if rc := C.octopus_stats_quantile(cvjson, C.double(quantile), &out); rc != 0 {
		return 0, errors.New("rust stats_quantile failed")
	}
	outStr := takeCString(out)
	var result struct {
		Quantile float64  `json:"quantile"`
		Value    *float64 `json:"value"`
	}
	if err := json.Unmarshal([]byte(outStr), &result); err != nil {
		return 0, err
	}
	if result.Value == nil {
		return 0, errors.New("empty values")
	}
	return *result.Value, nil
}

func mergeStatsHourlyGo(existing, delta StatsMetrics) (StatsMetrics, error) {
	existing.InputToken += delta.InputToken
	existing.OutputToken += delta.OutputToken
	existing.InputCost += delta.InputCost
	existing.OutputCost += delta.OutputCost
	existing.WaitTime += delta.WaitTime
	existing.RequestSuccess += delta.RequestSuccess
	existing.RequestFailed += delta.RequestFailed
	return existing, nil
}

func statsQuantileGo(values []float64, quantile float64) (float64, error) {
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
