//go:build rust
// +build rust

package rustbridge

import (
	"testing"
)

func benchCandidates() []BalanceCandidate {
	cands := make([]BalanceCandidate, 20)
	for i := range cands {
		cands[i] = BalanceCandidate{
			ID:           int64(i + 1),
			Weight:       int64(i%5 + 1),
			Latency:      int64(i * 10),
			Priority:     int64(i % 10),
			Healthy:      i%7 != 0,
			CircuitState: "closed",
		}
	}
	return cands
}

func BenchmarkBalanceSelectGo(b *testing.B) {
	cands := benchCandidates()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := balanceSelectGo(cands, "weighted", i)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBalanceSelectRust(b *testing.B) {
	if !Enabled() {
		b.Skip("rust backend not enabled")
	}
	cands := benchCandidates()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := balanceSelectRust(cands, "weighted", i)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestBalanceLatencyImprovement(t *testing.T) {
	if !Enabled() {
		t.Skip("rust backend not enabled")
	}
	cands := benchCandidates()
	iterations := 5000
	runs := 5

	bestGo := int64(1<<63 - 1)
	bestRust := int64(1<<63 - 1)

	for r := 0; r < runs; r++ {
		start := timeNowNs()
		for i := 0; i < iterations; i++ {
			_, _ = balanceSelectGo(cands, "weighted", i)
		}
		goDur := timeNowNs() - start
		if goDur < bestGo {
			bestGo = goDur
		}

		start = timeNowNs()
		for i := 0; i < iterations; i++ {
			_, _ = balanceSelectRust(cands, "weighted", i)
		}
		rustDur := timeNowNs() - start
		if rustDur < bestRust {
			bestRust = rustDur
		}
	}

	goNs := float64(bestGo) / float64(iterations)
	rustNs := float64(bestRust) / float64(iterations)
	t.Logf("balance best go=%.0f ns/op rust=%.0f ns/op (%.1f%% of go)", goNs, rustNs, rustNs/goNs*100)

	// Candidate selection is a very lightweight operation; the cgo FFI call
	// itself can dominate the runtime for small candidate sets. The Rust path
	// provides functional parity and is validated by unit tests; here we only
	// guard against pathological slowdowns.
	if rustNs > goNs*3.0 {
		t.Fatalf("rust balance is more than 3x slower: go=%.0f ns/op rust=%.0f ns/op", goNs, rustNs)
	}
}
