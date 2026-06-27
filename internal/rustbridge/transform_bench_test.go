//go:build rust
// +build rust

package rustbridge

import (
	"strings"
	"testing"
)

func benchChatRequestJSON() string {
	messages := []string{}
	for i := 0; i < 20; i++ {
		messages = append(messages, `{"role":"user","content":"`+strings.Repeat("word ", 20)+`"}`)
	}
	return `{"model":"gpt-4o","messages":[` + strings.Join(messages, ",") + `],"temperature":0.7,"max_tokens":1024,"stream":true,"tools":[{"type":"function","function":{"name":"weather","description":"get weather","parameters":{"type":"object"}}}],"tool_choice":"auto"}`
}

func BenchmarkTransformOpenAIChatRequestGo(b *testing.B) {
	req := benchChatRequestJSON()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := transformOpenAIChatRequestGo(req, "openai")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTransformOpenAIChatRequestRust(b *testing.B) {
	if !Enabled() {
		b.Skip("rust backend not enabled")
	}
	req := benchChatRequestJSON()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := transformOpenAIChatRequestRust(req, "openai")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestTransformLatencyImprovement(t *testing.T) {
	if !Enabled() {
		t.Skip("rust backend not enabled")
	}
	req := benchChatRequestJSON()
	iterations := 2000
	runs := 5

	bestGo := int64(1<<63 - 1)
	bestRust := int64(1<<63 - 1)

	for r := 0; r < runs; r++ {
		start := timeNowNs()
		for i := 0; i < iterations; i++ {
			_, _ = transformOpenAIChatRequestGo(req, "openai")
		}
		goDur := timeNowNs() - start
		if goDur < bestGo {
			bestGo = goDur
		}

		start = timeNowNs()
		for i := 0; i < iterations; i++ {
			_, _ = transformOpenAIChatRequestRust(req, "openai")
		}
		rustDur := timeNowNs() - start
		if rustDur < bestRust {
			bestRust = rustDur
		}
	}

	goNs := float64(bestGo) / float64(iterations)
	rustNs := float64(bestRust) / float64(iterations)
	t.Logf("transform best go=%.0f ns/op rust=%.0f ns/op (%.1f%% of go)", goNs, rustNs, rustNs/goNs*100)

	if rustNs > goNs*1.05 {
		t.Fatalf("rust transform is more than 5%% slower: go=%.0f ns/op rust=%.0f ns/op", goNs, rustNs)
	}
}
