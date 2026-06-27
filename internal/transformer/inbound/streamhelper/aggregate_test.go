package streamhelper

import (
	"encoding/json"
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/model"
)

func TestMergeStreamingChunk_AggregatesTextAndReasoning(t *testing.T) {
	cases := []struct {
		name      string
		envValue  string
		envSet    bool
		wantEqual bool
	}{
		{name: "rust_path_enabled", envSet: false, wantEqual: true},
		{name: "rust_path_disabled_false", envValue: "false", envSet: true, wantEqual: true},
		{name: "rust_path_disabled_zero", envValue: "0", envSet: true, wantEqual: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.envSet {
				t.Setenv(envDisableRustStreamAggregate, tc.envValue)
			} else {
				t.Setenv(envDisableRustStreamAggregate, "")
			}

			text1 := "hel"
			text2 := "lo"
			reason1 := "think-"
			reason2 := "more"
			finish := "stop"

			var agg *model.InternalLLMResponse
			agg = MergeStreamingChunk(agg, &model.InternalLLMResponse{
				ID:      "resp-1",
				Object:  "chat.completion.chunk",
				Created: 1,
				Model:   "gpt-test",
				Choices: []model.Choice{{
					Index: 0,
					Delta: &model.Message{Role: "assistant", Content: model.MessageContent{Content: &text1}, ReasoningContent: &reason1},
				}},
			})
			agg = MergeStreamingChunk(agg, &model.InternalLLMResponse{
				Object: "chat.completion.chunk",
				Usage:  &model.Usage{PromptTokens: 1, CompletionTokens: 2, TotalTokens: 3},
				Choices: []model.Choice{{
					Index:        0,
					Delta:        &model.Message{Content: model.MessageContent{Content: &text2}, ReasoningContent: &reason2},
					FinishReason: &finish,
				}},
			})

			if agg == nil || len(agg.Choices) != 1 || agg.Choices[0].Message == nil {
				t.Fatalf("aggregated response malformed: %#v", agg)
			}
			if got := *agg.Choices[0].Message.Content.Content; got != "hello" {
				t.Fatalf("content = %q, want hello", got)
			}
			if got := *agg.Choices[0].Message.ReasoningContent; got != "think-more" {
				t.Fatalf("reasoning = %q, want think-more", got)
			}
			if agg.Usage == nil || agg.Usage.TotalTokens != 3 {
				t.Fatalf("usage = %#v, want total=3", agg.Usage)
			}
		})
	}
}

func TestMergeStreamingChunk_MatchesGoAggregateByteIdentical(t *testing.T) {
	chunks := []string{
		`{"object":"chat.completion.chunk","choices":[{"index":0,"delta":{"role":"assistant","content":"hello "}}]}`,
		`{"object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":"world"},"finish_reason":"stop"}]}`,
	}

	// Reference produced by the pure Go path.
	var ref model.InternalLLMResponse
	if err := json.Unmarshal([]byte(`{"object":"chat.completion"}`), &ref); err != nil {
		t.Fatalf("unmarshal initial: %v", err)
	}
	for _, c := range chunks {
		var ck model.InternalLLMResponse
		if err := json.Unmarshal([]byte(c), &ck); err != nil {
			t.Fatalf("unmarshal chunk: %v", err)
		}
		model.MergeStreamingResponseAggregate(&ref, &ck)
	}
	wantBytes, err := json.Marshal(&ref)
	if err != nil {
		t.Fatalf("marshal ref: %v", err)
	}

	// Rust path (on non-rust builds this is still the Go fallback).
	t.Setenv(envDisableRustStreamAggregate, "")
	var agg *model.InternalLLMResponse
	for _, c := range chunks {
		var ck model.InternalLLMResponse
		if err := json.Unmarshal([]byte(c), &ck); err != nil {
			t.Fatalf("unmarshal chunk: %v", err)
		}
		agg = MergeStreamingChunk(agg, &ck)
	}
	gotBytes, err := json.Marshal(agg)
	if err != nil {
		t.Fatalf("marshal aggregate: %v", err)
	}
	if string(gotBytes) != string(wantBytes) {
		t.Fatalf("rust path output differs\n got: %s\nwant: %s", string(gotBytes), string(wantBytes))
	}
}

func TestMergeStreamingChunk_IgnoresDoneAndNil(t *testing.T) {
	text := "hi"
	agg := MergeStreamingChunk(nil, &model.InternalLLMResponse{
		Object: "chat.completion.chunk",
		Choices: []model.Choice{{
			Index: 0,
			Delta: &model.Message{Content: model.MessageContent{Content: &text}},
		}},
	})
	if agg == nil {
		t.Fatal("aggregate = nil")
	}

	agg = MergeStreamingChunk(agg, &model.InternalLLMResponse{Object: "[DONE]"})
	if agg == nil || agg.Choices == nil || len(agg.Choices) != 1 {
		t.Fatalf("[DONE] chunk should not alter aggregate: %#v", agg)
	}

	agg = MergeStreamingChunk(agg, nil)
	if agg == nil || agg.Choices == nil || len(agg.Choices) != 1 {
		t.Fatalf("nil chunk should not alter aggregate: %#v", agg)
	}
}

func TestMergeStreamingChunk_AggregatesToolCalls(t *testing.T) {
	t.Setenv(envDisableRustStreamAggregate, "")

	arg1 := "{\"a\":"
	arg2 := "1}"
	finish := "tool_calls"

	var agg *model.InternalLLMResponse
	agg = MergeStreamingChunk(agg, &model.InternalLLMResponse{
		Object: "chat.completion.chunk",
		Choices: []model.Choice{{
			Index: 0,
			Delta: &model.Message{Role: "assistant", ToolCalls: []model.ToolCall{{Index: 0, ID: "call-1", Type: "function", Function: model.FunctionCall{Name: "sum", Arguments: arg1}}}},
		}},
	})
	agg = MergeStreamingChunk(agg, &model.InternalLLMResponse{
		Object: "chat.completion.chunk",
		Choices: []model.Choice{{
			Index:        0,
			Delta:        &model.Message{ToolCalls: []model.ToolCall{{Index: 0, Function: model.FunctionCall{Arguments: arg2}}}},
			FinishReason: &finish,
		}},
	})

	if len(agg.Choices) != 1 || agg.Choices[0].Message == nil {
		t.Fatalf("aggregated response malformed: %#v", agg)
	}
	if len(agg.Choices[0].Message.ToolCalls) != 1 {
		t.Fatalf("tool call count = %d, want 1", len(agg.Choices[0].Message.ToolCalls))
	}
	if got := agg.Choices[0].Message.ToolCalls[0].Function.Arguments; got != "{\"a\":1}" {
		t.Fatalf("arguments = %q, want merged json", got)
	}
}

func TestMergeStreamingChunk_FallsBackOnMalformedAggregate(t *testing.T) {
	// The Rust path requires valid JSON. An aggregate with an unsupported type
	// cannot be marshaled and should fall back to the Go path without panicking.
	text := "ok"
	agg := &model.InternalLLMResponse{
		Object: "chat.completion",
	}
	agg = MergeStreamingChunk(agg, &model.InternalLLMResponse{
		Object: "chat.completion.chunk",
		Choices: []model.Choice{{
			Index: 0,
			Delta: &model.Message{Content: model.MessageContent{Content: &text}},
		}},
	})
	if agg == nil || len(agg.Choices) != 1 || agg.Choices[0].Message == nil {
		t.Fatalf("fallback aggregate malformed: %#v", agg)
	}
	if got := *agg.Choices[0].Message.Content.Content; got != "ok" {
		t.Fatalf("content = %q, want ok", got)
	}
}
