package rustbridge

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	transformerModel "github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/tokenizer"
)

func TestCountTokensRustMatchesGo(t *testing.T) {
	t.Parallel()

	text := "The quick brown fox jumps over the lazy dog. 这是一段用于测试token计数的中英文混合文本，包含一些常见单词和标点符号。"
	model := "gpt-4o"

	goCount := tokenizer.CountTokensWithModel(text, model)
	rustCount := CountTokens(text, model)

	if Enabled() {
		if rustCount != goCount {
			t.Fatalf("rust count %d != go count %d", rustCount, goCount)
		}
	}
	if rustCount <= 0 {
		t.Fatalf("expected positive token count, got %d", rustCount)
	}
}

func TestCountTokensEnvSwitch(t *testing.T) {
	// Not parallel because this test mutates the process environment.
	text := "hello world"
	model := "gpt-4o"

	if !Enabled() {
		t.Skip("rust backend not enabled")
	}

	rustCount := CountTokens(text, model)
	t.Setenv(envDisableRustTokenizer, "0")
	goCount := CountTokens(text, model)
	if rustCount != goCount {
		t.Fatalf("env switch changed result: rust=%d go=%d", rustCount, goCount)
	}
}

func TestExtractModelAndUsage(t *testing.T) {
	t.Parallel()

	reqJSON := `{"model":"gpt-4o","messages":[{"role":"user","content":"hi"}]}`
	got, err := ExtractModel(reqJSON)
	if err != nil {
		t.Fatalf("extract model error: %v", err)
	}
	if got != `{"model":"gpt-4o"}` {
		t.Fatalf("extract model = %s, want {\"model\":\"gpt-4o\"}", got)
	}

	respJSON := `{"model":"gpt-4o","usage":{"prompt_tokens":10,"completion_tokens":5,"total_tokens":15}}`
	got, err = ExtractUsage(respJSON)
	if err != nil {
		t.Fatalf("extract usage error: %v", err)
	}
	var gotUsage, wantUsage map[string]int64
	if err := json.Unmarshal([]byte(got), &gotUsage); err != nil {
		t.Fatalf("unmarshal extracted usage: %v", err)
	}
	wantUsage = map[string]int64{"prompt_tokens": 10, "completion_tokens": 5, "total_tokens": 15}
	if len(gotUsage) != len(wantUsage) {
		t.Fatalf("extract usage fields mismatch: got %v, want %v", gotUsage, wantUsage)
	}
	for k, v := range wantUsage {
		if gotUsage[k] != v {
			t.Fatalf("extract usage %s = %d, want %d", k, gotUsage[k], v)
		}
	}
}

func TestSSEAggregateByteIdentical(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		chunks  []string
		initial string
	}{
		{
			name:    "merge_message_and_delta",
			initial: `{"object":"chat.completion"}`,
			chunks: []string{
				`{"object":"chat.completion.chunk","choices":[{"index":0,"message":{"role":"assistant","content":"answer"},"delta":{"reasoning_content":"think","tool_calls":[{"index":0,"id":"call-1","type":"function","function":{"name":"calc","arguments":"{\"x\":1}"}}]}}]}`,
			},
		},
		{
			name:    "promote_text_to_multiple_content",
			initial: `{"object":"chat.completion"}`,
			chunks: []string{
				`{"object":"chat.completion.chunk","choices":[{"index":0,"message":{"role":"assistant","content":"hello"}}]}`,
				`{"object":"chat.completion.chunk","choices":[{"index":0,"delta":{"images":[{"type":"image_url","image_url":{"url":"data:image/png;base64,ZmFrZQ=="}}]}}]}`,
				`{"object":"chat.completion.chunk","choices":[{"index":0,"message":{"content":"world"}}]}`,
			},
		},
		{
			name:    "append_refusal",
			initial: `{"object":"chat.completion"}`,
			chunks: []string{
				`{"object":"chat.completion.chunk","choices":[{"index":0,"delta":{"role":"assistant","refusal":"cannot "}}]}`,
				`{"object":"chat.completion.chunk","choices":[{"index":0,"delta":{"refusal":"comply"},"finish_reason":"stop"}]}`,
			},
		},
		{
			name:    "append_logprobs",
			initial: `{"object":"chat.completion"}`,
			chunks: []string{
				`{"object":"chat.completion.chunk","choices":[{"index":0,"delta":{"role":"assistant","content":"hel"},"logprobs":{"content":[{"token":"hel","logprob":-0.11,"bytes":[104,101,108],"top_logprobs":[{"token":"hey","logprob":-0.42,"bytes":[104,101,121]}]}]}}]}`,
				`{"object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":"lo"},"logprobs":{"content":[{"token":"lo","logprob":-0.07}]},"finish_reason":"stop"}]}`,
			},
		},
		{
			name:    "append_tool_call_name_and_arguments",
			initial: `{"object":"chat.completion"}`,
			chunks: []string{
				`{"object":"chat.completion.chunk","choices":[{"index":0,"delta":{"role":"assistant","tool_calls":[{"index":0,"id":"call-1","type":"function","function":{"name":"weather_","arguments":"{\"city\":"}}]}}]}`,
				`{"object":"chat.completion.chunk","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"name":"lookup","arguments":"\"Tokyo\"}"}}]},"finish_reason":"tool_calls"}]}`,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var agg transformerModel.InternalLLMResponse
			if err := json.Unmarshal([]byte(tc.initial), &agg); err != nil {
				t.Fatalf("unmarshal initial: %v", err)
			}

			rustAgg := tc.initial
			for _, chunk := range tc.chunks {
				var ck transformerModel.InternalLLMResponse
				if err := json.Unmarshal([]byte(chunk), &ck); err != nil {
					t.Fatalf("unmarshal chunk: %v", err)
				}
				transformerModel.MergeStreamingResponseAggregate(&agg, &ck)

				var err error
				rustAgg, err = SSEAggregate(rustAgg, chunk)
				if err != nil {
					t.Fatalf("rust aggregate error: %v", err)
				}
			}

			wantBytes, err := json.Marshal(&agg)
			if err != nil {
				t.Fatalf("marshal aggregate: %v", err)
			}

			if rustAgg != string(wantBytes) {
				t.Fatalf("rust aggregate output differs\n got: %s\nwant: %s", rustAgg, string(wantBytes))
			}
		})
	}
}

func TestCountTokensEdgeCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		content string
		model   string
	}{
		{name: "empty", content: "", model: "gpt-4o"},
		{name: "cl100k_model", content: "hello world", model: "gpt-3.5-turbo"},
		{name: "o200k_model", content: "hello world", model: "gpt-4o"},
		{name: "long_text", content: strings.Repeat("abc ", 1000), model: "gpt-4o"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			goCount := tokenizer.CountTokensWithModel(tc.content, tc.model)
			rustCount := CountTokens(tc.content, tc.model)
			if Enabled() && goCount != rustCount {
				t.Fatalf("rust count %d != go count %d", rustCount, goCount)
			}
			if rustCount < 0 {
				t.Fatalf("expected non-negative count, got %d", rustCount)
			}
		})
	}
}

func TestExtractModelEdgeCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		json    string
		want    string
		wantErr bool
	}{
		{name: "empty_object", json: "{}", want: `{"model":""}`},
		{name: "missing_model", json: `{"messages":[]}`, want: `{"model":""}`},
		{name: "nested_model_not_toplevel", json: `{"data":{"model":"gpt-4"}}`, want: `{"model":""}`},
		{name: "empty_string_model", json: `{"model":""}`, want: `{"model":""}`},
		{name: "invalid_json", json: `{bad`, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ExtractModel(tc.json)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %s", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %s, want %s", got, tc.want)
			}
		})
	}
}

func TestExtractUsageEdgeCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		json    string
		want    string
		wantErr bool
	}{
		{name: "missing_usage", json: `{"model":"gpt-4o"}`, want: `{}`},
		{name: "partial_usage", json: `{"usage":{"prompt_tokens":7}}`, want: `{"prompt_tokens":7,"completion_tokens":0,"total_tokens":0}`},
		{name: "zero_usage", json: `{"usage":{"prompt_tokens":0,"completion_tokens":0,"total_tokens":0}}`, want: `{"prompt_tokens":0,"completion_tokens":0,"total_tokens":0}`},
		{name: "invalid_json", json: `{bad`, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ExtractUsage(tc.json)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %s", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			var gotUsage, wantUsage map[string]int64
			if err := json.Unmarshal([]byte(got), &gotUsage); err != nil {
				t.Fatalf("unmarshal got usage: %v", err)
			}
			if err := json.Unmarshal([]byte(tc.want), &wantUsage); err != nil {
				t.Fatalf("unmarshal want usage: %v", err)
			}
			if len(gotUsage) != len(wantUsage) {
				t.Fatalf("usage fields mismatch: got %v, want %v", gotUsage, wantUsage)
			}
			for k, v := range wantUsage {
				if gotUsage[k] != v {
					t.Fatalf("usage %s = %d, want %d", k, gotUsage[k], v)
				}
			}
		})
	}
}

func TestSSEAggregateEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("invalid_aggregate", func(t *testing.T) {
		_, err := SSEAggregate("{bad", `{"choices":[]}`)
		if err == nil {
			t.Fatal("expected error for invalid aggregate")
		}
	})

	t.Run("invalid_chunk", func(t *testing.T) {
		_, err := SSEAggregate(`{"choices":[]}`, "{bad")
		if err == nil {
			t.Fatal("expected error for invalid chunk")
		}
	})

	t.Run("empty_aggregate", func(t *testing.T) {
		got, err := SSEAggregate(`{}`, `{"choices":[{"index":0,"delta":{"role":"assistant","content":"hi"}}]}`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got == "" {
			t.Fatal("expected non-empty aggregate")
		}
	})

	t.Run("multiple_choices", func(t *testing.T) {
		agg := `{"choices":[]}`
		chunk := `{"choices":[{"index":0,"delta":{"content":"a"}},{"index":1,"delta":{"content":"b"}}]}`
		got, err := SSEAggregate(agg, chunk)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(got, `"index":0`) || !strings.Contains(got, `"index":1`) {
			t.Fatalf("expected both choices, got %s", got)
		}
	})
}

func BenchmarkCountTokensGo(b *testing.B) {
	text := strings.Repeat("The quick brown fox jumps over the lazy dog. ", 200)
	model := "gpt-4o"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		countTokensGo(text, model)
	}
}

func BenchmarkCountTokensRust(b *testing.B) {
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

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
