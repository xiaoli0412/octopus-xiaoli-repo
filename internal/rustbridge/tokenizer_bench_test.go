//go:build rust
// +build rust

package rustbridge

import (
	"strings"
	"testing"
	"time"
)

func BenchmarkCountTokensGoBaseline(b *testing.B) {
	text := strings.Repeat("The quick brown fox jumps over the lazy dog. ", 200)
	model := "gpt-4o"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		countTokensGo(text, model)
	}
}

func BenchmarkCountTokensRustFFI(b *testing.B) {
	if !Enabled() {
		b.Skip("rust backend not enabled")
	}
	text := strings.Repeat("The quick brown fox jumps over the lazy dog. ", 200)
	model := "gpt-4o"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		countTokensRust(text, model)
	}
}

func TestRustTokenizerLatencyImprovement(t *testing.T) {
	if !Enabled() {
		t.Skip("rust backend not enabled")
	}

	text := strings.Repeat("The quick brown fox jumps over the lazy dog. ", 200)
	model := "gpt-4o"
	iterations := 2000
	runs := 5

	bestGo := time.Duration(1<<63 - 1)
	bestRust := time.Duration(1<<63 - 1)

	for r := 0; r < runs; r++ {
		start := time.Now()
		for i := 0; i < iterations; i++ {
			countTokensGo(text, model)
		}
		goDur := time.Since(start)
		if goDur < bestGo {
			bestGo = goDur
		}

		start = time.Now()
		for i := 0; i < iterations; i++ {
			countTokensRust(text, model)
		}
		rustDur := time.Since(start)
		if rustDur < bestRust {
			bestRust = rustDur
		}
	}

	goNs := float64(bestGo) / float64(iterations)
	rustNs := float64(bestRust) / float64(iterations)
	t.Logf("best go=%.0f ns/op rust=%.0f ns/op (%.1f%% of go)", goNs, rustNs, rustNs/goNs*100)

	if rustNs > goNs*0.80 {
		t.Fatalf("rust tokenizer is not at least 20%% faster: go=%.0f ns/op rust=%.0f ns/op", goNs, rustNs)
	}
}
