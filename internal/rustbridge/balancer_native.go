//go:build !rust
// +build !rust

package rustbridge

import (
	"errors"
	"math/rand"
	"sort"
)

// BalanceSelect chooses a candidate using the pure-Go implementation.
func BalanceSelect(candidates []BalanceCandidate, strategy string, currentIdx int) (BalanceResult, error) {
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
	default:
		return BalanceResult{}, errors.New("unknown strategy")
	}
}
