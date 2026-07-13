//go:build rust
// +build rust

package rustbridge

/*
#cgo windows LDFLAGS: -L${SRCDIR}/../../rust/core/target/release -loctopus_core -lws2_32 -luserenv -ladvapi32 -lntdll -lbcrypt
#cgo linux LDFLAGS: -L${SRCDIR}/../../rust/core/target/release -loctopus_core -lm
#cgo darwin LDFLAGS: -L${SRCDIR}/../../rust/core/target/release -loctopus_core -framework Security -framework CoreFoundation
#include <stdlib.h>
typedef struct {
    long long id;
    long long weight;
    long long latency;
    long long priority;
    int healthy;
    int circuit_open;
} OctopusBalanceCandidate;
extern int octopus_balance_select(const char* candidates_json, const char* strategy, int current_idx, char** out);
extern int octopus_balance_select_v2(const OctopusBalanceCandidate* candidates, int candidates_len, const char* strategy, int current_idx, long long* out_id, int* out_next_index);
extern void octopus_free_string(char* s);
*/
import "C"
import (
	"errors"
	"math/rand"
	"sort"
	"unsafe"
)

// BalanceSelect chooses a candidate using the Rust FFI accelerator.
// Set OCTOPUS_RUST_BALANCER=0 to use the Go fallback.
func BalanceSelect(candidates []BalanceCandidate, strategy string, currentIdx int) (BalanceResult, error) {
	if !envEnabled(envDisableRustBalancer) {
		return balanceSelectGo(candidates, strategy, currentIdx)
	}
	return balanceSelectRust(candidates, strategy, currentIdx)
}

func balanceSelectRust(candidates []BalanceCandidate, strategy string, currentIdx int) (BalanceResult, error) {
	ccands := make([]C.OctopusBalanceCandidate, len(candidates))
	for i, c := range candidates {
		ccands[i] = C.OctopusBalanceCandidate{
			id:           C.longlong(c.ID),
			weight:       C.longlong(c.Weight),
			latency:      C.longlong(c.Latency),
			priority:     C.longlong(c.Priority),
			healthy:      cbool(c.Healthy),
			circuit_open: cbool(c.CircuitState == "open"),
		}
	}
	cstrategy := C.CString(strategy)
	defer C.free(unsafe.Pointer(cstrategy))
	var outID C.longlong
	var outNext C.int
	if rc := C.octopus_balance_select_v2(
		(*C.OctopusBalanceCandidate)(unsafe.Pointer(&ccands[0])),
		C.int(len(ccands)),
		cstrategy,
		C.int(currentIdx),
		&outID,
		&outNext,
	); rc != 0 {
		return BalanceResult{}, errors.New("rust balance_select failed")
	}
	return BalanceResult{ID: int64(outID), NextIndex: int(outNext)}, nil
}

func cbool(v bool) C.int {
	if v {
		return 1
	}
	return 0
}

// balanceSelectGo is the pure-Go implementation kept in the Rust build for
// benchmarks and parity checks.
func balanceSelectGo(candidates []BalanceCandidate, strategy string, currentIdx int) (BalanceResult, error) {
	available := make([]BalanceCandidate, 0, len(candidates))
	for _, c := range candidates {
		if c.Healthy && c.CircuitState != "open" {
			available = append(available, c)
		}
	}
	if len(available) == 0 {
		return BalanceResult{}, errors.New("no available candidates")
	}

	switch strategy {
	case "weighted":
		var seq []int64
		for _, c := range available {
			w := c.Weight
			if w <= 0 {
				w = 1
			}
			for i := int64(0); i < w; i++ {
				seq = append(seq, c.ID)
			}
		}
		idx := currentIdx % len(seq)
		if idx < 0 {
			idx = 0
		}
		return BalanceResult{ID: seq[idx], NextIndex: idx + 1}, nil
	case "round_robin":
		sort.Slice(available, func(i, j int) bool {
			if available[i].Priority != available[j].Priority {
				return available[i].Priority < available[j].Priority
			}
			if available[i].Latency != available[j].Latency {
				return available[i].Latency < available[j].Latency
			}
			return available[i].ID < available[j].ID
		})
		idx := currentIdx % len(available)
		if idx < 0 {
			idx = 0
		}
		return BalanceResult{ID: available[idx].ID, NextIndex: idx + 1}, nil
	case "random":
		c := available[rand.Intn(len(available))]
		return BalanceResult{ID: c.ID}, nil
	case "failover":
		sort.Slice(available, func(i, j int) bool {
			if available[i].Priority != available[j].Priority {
				return available[i].Priority < available[j].Priority
			}
			if available[i].Latency != available[j].Latency {
				return available[i].Latency < available[j].Latency
			}
			return available[i].ID < available[j].ID
		})
		return BalanceResult{ID: available[0].ID}, nil
	case "least_latency":
		sort.Slice(available, func(i, j int) bool {
			if available[i].Latency != available[j].Latency {
				return available[i].Latency < available[j].Latency
			}
			return available[i].ID < available[j].ID
		})
		return BalanceResult{ID: available[0].ID}, nil
	case "health_aware":
		sort.Slice(available, func(i, j int) bool {
			ci := circuitSeverityRank(available[i].CircuitState)
			cj := circuitSeverityRank(available[j].CircuitState)
			if ci != cj {
				return ci < cj
			}
			if available[i].Latency != available[j].Latency {
				return available[i].Latency < available[j].Latency
			}
			return available[i].ID < available[j].ID
		})
		return BalanceResult{ID: available[0].ID}, nil
	default:
		return BalanceResult{}, errors.New("unknown strategy")
	}
}

// circuitSeverityRank ranks circuit states for health-aware sorting:
// closed (0) < half-open (1) < open (2). Open candidates are already
// filtered out by the available check, but the rank is kept for safety.
func circuitSeverityRank(state string) int {
	switch state {
	case "open":
		return 2
	case "half-open":
		return 1
	default:
		return 0
	}
}
