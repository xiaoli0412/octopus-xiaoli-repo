package openai

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/model"
	outboundopenai "github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/outbound/openai"
)

func TestResponseInboundAggregatesStreamWithoutChunkRetention(t *testing.T) {
	t.Parallel()

	inbound := &ResponseInbound{}
	arg1 := "{\"a\":"
	arg2 := "1}"
	finish := "tool_calls"

	_, err := inbound.TransformStream(context.Background(), &model.InternalLLMResponse{
		ID:      "resp-2",
		Object:  "chat.completion.chunk",
		Created: 2,
		Model:   "gpt-resp",
		Choices: []model.Choice{{
			Index: 0,
			Delta: &model.Message{Role: "assistant", ToolCalls: []model.ToolCall{{Index: 0, ID: "call-1", Type: "function", Function: model.FunctionCall{Name: "sum", Arguments: arg1}}}},
		}},
	})
	if err != nil {
		t.Fatalf("TransformStream() first chunk error = %v", err)
	}
	_, err = inbound.TransformStream(context.Background(), &model.InternalLLMResponse{
		Object: "chat.completion.chunk",
		Usage:  &model.Usage{PromptTokens: 2, CompletionTokens: 3, TotalTokens: 5},
		Choices: []model.Choice{{
			Index:        0,
			Delta:        &model.Message{ToolCalls: []model.ToolCall{{Index: 0, Function: model.FunctionCall{Arguments: arg2}}}},
			FinishReason: &finish,
		}},
	})
	if err != nil {
		t.Fatalf("TransformStream() second chunk error = %v", err)
	}

	resp, err := inbound.GetInternalResponse(context.Background())
	if err != nil {
		t.Fatalf("GetInternalResponse() error = %v", err)
	}
	if resp == nil || len(resp.Choices) != 1 || resp.Choices[0].Message == nil {
		t.Fatalf("aggregated response malformed: %#v", resp)
	}
	if len(resp.Choices[0].Message.ToolCalls) != 1 {
		t.Fatalf("tool call count = %d, want 1", len(resp.Choices[0].Message.ToolCalls))
	}
	if got := resp.Choices[0].Message.ToolCalls[0].Function.Arguments; got != "{\"a\":1}" {
		t.Fatalf("arguments = %q, want merged json", got)
	}
	if inbound.streamResult == nil {
		t.Fatal("streamResult = nil, want aggregated state retained")
	}
}

func TestResponseInboundAggregatesRefusalAcrossStreamChunks(t *testing.T) {
	t.Parallel()

	inbound := &ResponseInbound{}
	finish := "stop"

	_, err := inbound.TransformStream(context.Background(), &model.InternalLLMResponse{
		ID:      "resp-refusal",
		Object:  "chat.completion.chunk",
		Created: 5,
		Model:   "gpt-resp",
		Choices: []model.Choice{{
			Index: 0,
			Delta: &model.Message{Role: "assistant", Refusal: "cannot "},
		}},
	})
	if err != nil {
		t.Fatalf("TransformStream() first chunk error = %v", err)
	}
	_, err = inbound.TransformStream(context.Background(), &model.InternalLLMResponse{
		Object: "chat.completion.chunk",
		Usage:  &model.Usage{PromptTokens: 2, CompletionTokens: 1, TotalTokens: 3},
		Choices: []model.Choice{{
			Index:        0,
			Delta:        &model.Message{Refusal: "comply"},
			FinishReason: &finish,
		}},
	})
	if err != nil {
		t.Fatalf("TransformStream() second chunk error = %v", err)
	}

	resp, err := inbound.GetInternalResponse(context.Background())
	if err != nil {
		t.Fatalf("GetInternalResponse() error = %v", err)
	}
	if resp == nil || len(resp.Choices) != 1 || resp.Choices[0].Message == nil {
		t.Fatalf("aggregated response malformed: %#v", resp)
	}
	if got := resp.Choices[0].Message.Refusal; got != "cannot comply" {
		t.Fatalf("refusal = %q, want merged refusal", got)
	}
}

func TestResponseInboundAggregatesLogprobsAcrossStreamChunks(t *testing.T) {
	t.Parallel()

	inbound := &ResponseInbound{}
	firstText := "hel"
	secondText := "lo"
	finish := "stop"

	firstChunk := &model.InternalLLMResponse{
		ID:      "resp-logprobs",
		Object:  "chat.completion.chunk",
		Created: 6,
		Model:   "gpt-resp",
		Choices: []model.Choice{{
			Index: 0,
			Delta: &model.Message{
				Role: "assistant",
				Content: model.MessageContent{
					Content: &firstText,
				},
			},
			Logprobs: &model.LogprobsContent{Content: []model.TokenLogprob{{
				Token:   "hel",
				Logprob: -0.11,
				Bytes:   []int{104, 101, 108},
			}}},
		}},
	}

	_, err := inbound.TransformStream(context.Background(), firstChunk)
	if err != nil {
		t.Fatalf("TransformStream() first chunk error = %v", err)
	}
	firstChunk.Choices[0].Logprobs.Content[0].Bytes[0] = 0

	_, err = inbound.TransformStream(context.Background(), &model.InternalLLMResponse{
		Object: "chat.completion.chunk",
		Usage:  &model.Usage{PromptTokens: 2, CompletionTokens: 2, TotalTokens: 4},
		Choices: []model.Choice{{
			Index:        0,
			Delta:        &model.Message{Content: model.MessageContent{Content: &secondText}},
			FinishReason: &finish,
			Logprobs: &model.LogprobsContent{Content: []model.TokenLogprob{{
				Token:   "lo",
				Logprob: -0.07,
			}}},
		}},
	})
	if err != nil {
		t.Fatalf("TransformStream() second chunk error = %v", err)
	}

	resp, err := inbound.GetInternalResponse(context.Background())
	if err != nil {
		t.Fatalf("GetInternalResponse() error = %v", err)
	}
	if resp == nil || len(resp.Choices) != 1 || resp.Choices[0].Message == nil {
		t.Fatalf("aggregated response malformed: %#v", resp)
	}
	if resp.Choices[0].Logprobs == nil {
		t.Fatal("logprobs = nil, want aggregated content")
	}
	if len(resp.Choices[0].Logprobs.Content) != 2 {
		t.Fatalf("logprobs token count = %d, want 2", len(resp.Choices[0].Logprobs.Content))
	}
	if got := resp.Choices[0].Logprobs.Content[0].Bytes; len(got) != 3 || got[0] != 104 {
		t.Fatalf("first token bytes = %#v, want preserved deep copy", got)
	}
	if got := resp.Choices[0].Logprobs.Content[1].Token; got != "lo" {
		t.Fatalf("second token = %q, want lo", got)
	}
}

func TestResponseInboundAggregatesToolCallNameAndArgumentsAcrossStreamChunks(t *testing.T) {
	t.Parallel()

	inbound := &ResponseInbound{}
	finish := "tool_calls"
	arg1 := "{\"city\":"
	arg2 := "\"Tokyo\"}"

	_, err := inbound.TransformStream(context.Background(), &model.InternalLLMResponse{
		ID:      "resp-tool-name",
		Object:  "chat.completion.chunk",
		Created: 7,
		Model:   "gpt-resp",
		Choices: []model.Choice{{
			Index: 0,
			Delta: &model.Message{Role: "assistant", ToolCalls: []model.ToolCall{{
				Index: 0,
				ID:    "call-1",
				Type:  "function",
				Function: model.FunctionCall{
					Name:      "weather_",
					Arguments: arg1,
				},
			}}},
		}},
	})
	if err != nil {
		t.Fatalf("TransformStream() first chunk error = %v", err)
	}
	_, err = inbound.TransformStream(context.Background(), &model.InternalLLMResponse{
		Object: "chat.completion.chunk",
		Choices: []model.Choice{{
			Index: 0,
			Delta: &model.Message{ToolCalls: []model.ToolCall{{
				Index: 0,
				Function: model.FunctionCall{
					Name:      "lookup",
					Arguments: arg2,
				},
			}}},
			FinishReason: &finish,
		}},
	})
	if err != nil {
		t.Fatalf("TransformStream() second chunk error = %v", err)
	}

	resp, err := inbound.GetInternalResponse(context.Background())
	if err != nil {
		t.Fatalf("GetInternalResponse() error = %v", err)
	}
	if resp == nil || len(resp.Choices) != 1 || resp.Choices[0].Message == nil {
		t.Fatalf("aggregated response malformed: %#v", resp)
	}
	if len(resp.Choices[0].Message.ToolCalls) != 1 {
		t.Fatalf("tool call count = %d, want 1", len(resp.Choices[0].Message.ToolCalls))
	}
	if got := resp.Choices[0].Message.ToolCalls[0].Function.Name; got != "weather_lookup" {
		t.Fatalf("tool name = %q, want weather_lookup", got)
	}
	if got := resp.Choices[0].Message.ToolCalls[0].Function.Arguments; got != "{\"city\":\"Tokyo\"}" {
		t.Fatalf("arguments = %q, want merged json", got)
	}
}

func TestResponsesStreamRoundTripPreservesToolCallNameFragments(t *testing.T) {
	t.Parallel()

	inbound := &ResponseInbound{}
	outbound := &outboundopenai.ResponseOutbound{}
	finish := "tool_calls"
	arg1 := "{\"city\":"
	arg2 := "\"Tokyo\"}"

	firstEvents, err := inbound.TransformStream(context.Background(), &model.InternalLLMResponse{
		ID:      "resp-tool-roundtrip",
		Object:  "chat.completion.chunk",
		Created: 8,
		Model:   "gpt-resp",
		Choices: []model.Choice{{
			Index: 0,
			Delta: &model.Message{Role: "assistant", ToolCalls: []model.ToolCall{{
				Index: 0,
				ID:    "call-1",
				Type:  "function",
				Function: model.FunctionCall{
					Name:      "weather_",
					Arguments: arg1,
				},
			}}},
		}},
	})
	if err != nil {
		t.Fatalf("TransformStream() first chunk error = %v", err)
	}
	secondEvents, err := inbound.TransformStream(context.Background(), &model.InternalLLMResponse{
		Object: "chat.completion.chunk",
		Choices: []model.Choice{{
			Index: 0,
			Delta: &model.Message{ToolCalls: []model.ToolCall{{
				Index: 0,
				Function: model.FunctionCall{
					Name:      "lookup",
					Arguments: arg2,
				},
			}}},
			FinishReason: &finish,
		}},
	})
	if err != nil {
		t.Fatalf("TransformStream() second chunk error = %v", err)
	}

	aggregated := &model.InternalLLMResponse{Object: "chat.completion"}
	for _, payload := range [][]byte{firstEvents, secondEvents} {
		for _, line := range strings.Split(string(payload), "\n") {
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			chunk, err := outbound.TransformStream(context.Background(), []byte(strings.TrimPrefix(line, "data: ")))
			if err != nil {
				t.Fatalf("TransformStream() outbound error = %v", err)
			}
			if chunk != nil {
				model.MergeStreamingResponseAggregate(aggregated, chunk)
			}
		}
	}

	if len(aggregated.Choices) != 1 || aggregated.Choices[0].Message == nil {
		t.Fatalf("round-trip aggregate malformed: %#v", aggregated)
	}
	if len(aggregated.Choices[0].Message.ToolCalls) != 1 {
		t.Fatalf("tool call count = %d, want 1", len(aggregated.Choices[0].Message.ToolCalls))
	}
	if got := aggregated.Choices[0].Message.ToolCalls[0].Function.Name; got != "weather_lookup" {
		t.Fatalf("tool name = %q, want weather_lookup", got)
	}
	if got := aggregated.Choices[0].Message.ToolCalls[0].Function.Arguments; got != "{\"city\":\"Tokyo\"}" {
		t.Fatalf("arguments = %q, want merged json", got)
	}
}

func TestResponsesStreamRoundTripKeepsToolCallIndexesStableAcrossOutputItems(t *testing.T) {
	t.Parallel()

	inbound := &ResponseInbound{}
	outbound := &outboundopenai.ResponseOutbound{}
	finish := "tool_calls"
	preface := "Working on it."
	weatherArgs1 := "{\"city\":"
	weatherArgs2 := "\"Tokyo\"}"
	timeArgs1 := "{\"zone\":"
	timeArgs2 := "\"Asia/Tokyo\"}"

	firstEvents, err := inbound.TransformStream(context.Background(), &model.InternalLLMResponse{
		ID:      "resp-tool-index",
		Object:  "chat.completion.chunk",
		Created: 9,
		Model:   "gpt-resp",
		Choices: []model.Choice{{
			Index: 0,
			Delta: &model.Message{Role: "assistant", Content: model.MessageContent{Content: &preface}},
		}},
	})
	if err != nil {
		t.Fatalf("TransformStream() preface chunk error = %v", err)
	}
	secondEvents, err := inbound.TransformStream(context.Background(), &model.InternalLLMResponse{
		Object: "chat.completion.chunk",
		Choices: []model.Choice{{
			Index: 0,
			Delta: &model.Message{ToolCalls: []model.ToolCall{
				{
					Index: 0,
					ID:    "call-weather",
					Type:  "function",
					Function: model.FunctionCall{
						Name:      "weather_",
						Arguments: weatherArgs1,
					},
				},
				{
					Index: 1,
					ID:    "call-time",
					Type:  "function",
					Function: model.FunctionCall{
						Name:      "time_",
						Arguments: timeArgs1,
					},
				},
			}},
		}},
	})
	if err != nil {
		t.Fatalf("TransformStream() first tool chunk error = %v", err)
	}
	thirdEvents, err := inbound.TransformStream(context.Background(), &model.InternalLLMResponse{
		Object: "chat.completion.chunk",
		Choices: []model.Choice{{
			Index: 0,
			Delta: &model.Message{ToolCalls: []model.ToolCall{
				{
					Index: 0,
					Function: model.FunctionCall{
						Name:      "lookup",
						Arguments: weatherArgs2,
					},
				},
				{
					Index: 1,
					Function: model.FunctionCall{
						Name:      "lookup",
						Arguments: timeArgs2,
					},
				},
			}},
			FinishReason: &finish,
		}},
	})
	if err != nil {
		t.Fatalf("TransformStream() second tool chunk error = %v", err)
	}

	aggregated := &model.InternalLLMResponse{Object: "chat.completion"}
	for _, payload := range [][]byte{firstEvents, secondEvents, thirdEvents} {
		for _, line := range strings.Split(string(payload), "\n") {
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			chunk, err := outbound.TransformStream(context.Background(), []byte(strings.TrimPrefix(line, "data: ")))
			if err != nil {
				t.Fatalf("TransformStream() outbound error = %v", err)
			}
			if chunk != nil {
				model.MergeStreamingResponseAggregate(aggregated, chunk)
			}
		}
	}

	if len(aggregated.Choices) != 1 || aggregated.Choices[0].Message == nil {
		t.Fatalf("round-trip aggregate malformed: %#v", aggregated)
	}
	message := aggregated.Choices[0].Message
	if message.Content.Content == nil || *message.Content.Content != "Working on it." {
		t.Fatalf("content = %#v, want preserved preface text", message.Content.Content)
	}
	if len(message.ToolCalls) != 2 {
		t.Fatalf("tool call count = %d, want 2", len(message.ToolCalls))
	}
	if message.ToolCalls[0].Index != 0 || message.ToolCalls[1].Index != 1 {
		t.Fatalf("tool call indexes = [%d %d], want [0 1]", message.ToolCalls[0].Index, message.ToolCalls[1].Index)
	}
	if got := message.ToolCalls[0].Function.Name; got != "weather_lookup" {
		t.Fatalf("first tool name = %q, want weather_lookup", got)
	}
	if got := message.ToolCalls[1].Function.Name; got != "time_lookup" {
		t.Fatalf("second tool name = %q, want time_lookup", got)
	}
	if got := message.ToolCalls[0].Function.Arguments; got != "{\"city\":\"Tokyo\"}" {
		t.Fatalf("first tool arguments = %q, want merged weather args", got)
	}
	if got := message.ToolCalls[1].Function.Arguments; got != "{\"zone\":\"Asia/Tokyo\"}" {
		t.Fatalf("second tool arguments = %q, want merged time args", got)
	}
}

func TestResponsesStreamRoundTripKeepsDistinctItemIDsForCallIDLessToolCalls(t *testing.T) {
	t.Parallel()

	inbound := &ResponseInbound{}
	outbound := &outboundopenai.ResponseOutbound{}
	finish := "tool_calls"
	weatherArgs1 := "{\"city\":"
	weatherArgs2 := "\"Tokyo\"}"
	timeArgs1 := "{\"zone\":"
	timeArgs2 := "\"Asia/Tokyo\"}"

	firstEvents, err := inbound.TransformStream(context.Background(), &model.InternalLLMResponse{
		ID:      "resp-tool-no-callid",
		Object:  "chat.completion.chunk",
		Created: 10,
		Model:   "gpt-resp",
		Choices: []model.Choice{{
			Index: 0,
			Delta: &model.Message{Role: "assistant", ToolCalls: []model.ToolCall{
				{
					Index: 0,
					Type:  "function",
					Function: model.FunctionCall{
						Name:      "weather_",
						Arguments: weatherArgs1,
					},
				},
				{
					Index: 1,
					Type:  "function",
					Function: model.FunctionCall{
						Name:      "time_",
						Arguments: timeArgs1,
					},
				},
			}},
		}},
	})
	if err != nil {
		t.Fatalf("TransformStream() first tool chunk error = %v", err)
	}
	secondEvents, err := inbound.TransformStream(context.Background(), &model.InternalLLMResponse{
		Object: "chat.completion.chunk",
		Choices: []model.Choice{{
			Index: 0,
			Delta: &model.Message{ToolCalls: []model.ToolCall{
				{
					Index: 0,
					Function: model.FunctionCall{
						Name:      "lookup",
						Arguments: weatherArgs2,
					},
				},
				{
					Index: 1,
					Function: model.FunctionCall{
						Name:      "lookup",
						Arguments: timeArgs2,
					},
				},
			}},
			FinishReason: &finish,
		}},
	})
	if err != nil {
		t.Fatalf("TransformStream() second tool chunk error = %v", err)
	}

	var doneItemIDs []string
	for _, payload := range [][]byte{firstEvents, secondEvents} {
		for _, line := range strings.Split(string(payload), "\n") {
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			var event outboundopenai.ResponsesStreamEvent
			if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &event); err != nil {
				t.Fatalf("json.Unmarshal() event error = %v", err)
			}
			if event.Type == "response.output_item.done" && event.Item != nil && event.Item.Type == "function_call" {
				doneItemIDs = append(doneItemIDs, event.Item.ID)
			}
		}
	}
	if len(doneItemIDs) != 2 {
		t.Fatalf("done item count = %d, want 2", len(doneItemIDs))
	}
	if doneItemIDs[0] == "" || doneItemIDs[1] == "" {
		t.Fatalf("done item ids = %#v, want both non-empty", doneItemIDs)
	}
	if doneItemIDs[0] == doneItemIDs[1] {
		t.Fatalf("done item ids = %#v, want distinct generated ids", doneItemIDs)
	}

	aggregated := &model.InternalLLMResponse{Object: "chat.completion"}
	for _, payload := range [][]byte{firstEvents, secondEvents} {
		for _, line := range strings.Split(string(payload), "\n") {
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			chunk, err := outbound.TransformStream(context.Background(), []byte(strings.TrimPrefix(line, "data: ")))
			if err != nil {
				t.Fatalf("TransformStream() outbound error = %v", err)
			}
			if chunk != nil {
				model.MergeStreamingResponseAggregate(aggregated, chunk)
			}
		}
	}

	if len(aggregated.Choices) != 1 || aggregated.Choices[0].Message == nil {
		t.Fatalf("round-trip aggregate malformed: %#v", aggregated)
	}
	toolCalls := aggregated.Choices[0].Message.ToolCalls
	if len(toolCalls) != 2 {
		t.Fatalf("tool call count = %d, want 2", len(toolCalls))
	}
	if toolCalls[0].Index != 0 || toolCalls[1].Index != 1 {
		t.Fatalf("tool call indexes = [%d %d], want [0 1]", toolCalls[0].Index, toolCalls[1].Index)
	}
	if toolCalls[0].Function.Name != "weather_lookup" || toolCalls[1].Function.Name != "time_lookup" {
		t.Fatalf("tool call names = [%q %q], want [weather_lookup time_lookup]", toolCalls[0].Function.Name, toolCalls[1].Function.Name)
	}
	if toolCalls[0].Function.Arguments != "{\"city\":\"Tokyo\"}" || toolCalls[1].Function.Arguments != "{\"zone\":\"Asia/Tokyo\"}" {
		t.Fatalf("tool call arguments = [%q %q], want merged payloads", toolCalls[0].Function.Arguments, toolCalls[1].Function.Arguments)
	}
}

func TestResponseInboundStreamDefersIncompleteTerminalEventUntilDoneMarker(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		responseID           string
		finishReason         string
		wantIncompleteReason string
		wantFinishReason     string
	}{
		{name: "length", responseID: "resp-length", finishReason: "length", wantIncompleteReason: "length", wantFinishReason: "length"},
		{name: "content_filter", responseID: "resp-content-filter", finishReason: "content_filter", wantIncompleteReason: "content_filter", wantFinishReason: "content_filter"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			inbound := &ResponseInbound{}
			outbound := &outboundopenai.ResponseOutbound{}
			text := "partial"

			firstEvents, err := inbound.TransformStream(context.Background(), &model.InternalLLMResponse{
				ID:      tt.responseID,
				Object:  "chat.completion.chunk",
				Created: 10,
				Model:   "gpt-resp",
				Choices: []model.Choice{{
					Index: 0,
					Delta: &model.Message{Role: "assistant", Content: model.MessageContent{Content: &text}},
				}},
			})
			if err != nil {
				t.Fatalf("TransformStream() first chunk error = %v", err)
			}
			if len(firstEvents) == 0 {
				t.Fatal("first events empty, want response.created/in_progress payload")
			}

			terminalEvents, err := inbound.TransformStream(context.Background(), &model.InternalLLMResponse{
				Object: "chat.completion.chunk",
				Choices: []model.Choice{{
					Index:        0,
					FinishReason: &tt.finishReason,
				}},
			})
			if err != nil {
				t.Fatalf("TransformStream() terminal chunk error = %v", err)
			}
			for _, line := range strings.Split(string(terminalEvents), "\n") {
				line = strings.TrimSpace(line)
				if !strings.HasPrefix(line, "data: ") {
					continue
				}
				var event ResponsesStreamEvent
				if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &event); err != nil {
					t.Fatalf("json.Unmarshal() terminal chunk event error = %v", err)
				}
				if event.Type == "response.incomplete" {
					t.Fatalf("terminal chunk events = %s, want terminal response deferred until usage/[DONE]", string(terminalEvents))
				}
			}

			doneEvents, err := inbound.TransformStream(context.Background(), &model.InternalLLMResponse{Object: "[DONE]"})
			if err != nil {
				t.Fatalf("TransformStream() done marker error = %v", err)
			}

			roundTripAggregated := &model.InternalLLMResponse{Object: "chat.completion"}
			var found bool
			for _, payload := range [][]byte{firstEvents, doneEvents} {
				for _, line := range strings.Split(string(payload), "\n") {
					line = strings.TrimSpace(line)
					if !strings.HasPrefix(line, "data: ") {
						continue
					}
					if line == "data: [DONE]" {
						continue
					}

					raw := []byte(strings.TrimPrefix(line, "data: "))
					var event ResponsesStreamEvent
					if err := json.Unmarshal(raw, &event); err != nil {
						t.Fatalf("json.Unmarshal() event error = %v", err)
					}
					if event.Type == "response.incomplete" {
						found = true
						if event.Response == nil || event.Response.Status == nil || *event.Response.Status != "incomplete" {
							t.Fatalf("terminal response status = %#v, want incomplete", event.Response)
						}
						if event.Response.Usage != nil {
							t.Fatalf("terminal response usage = %#v, want nil before usage chunk", event.Response.Usage)
						}
						if event.Response.IncompleteDetails == nil || event.Response.IncompleteDetails.Reason != tt.wantIncompleteReason {
							t.Fatalf("IncompleteDetails = %#v, want %s", event.Response.IncompleteDetails, tt.wantIncompleteReason)
						}
					}

					chunk, err := outbound.TransformStream(context.Background(), raw)
					if err != nil {
						t.Fatalf("TransformStream() outbound error = %v", err)
					}
					if chunk != nil {
						model.MergeStreamingResponseAggregate(roundTripAggregated, chunk)
					}
				}
			}
			if !found {
				t.Fatalf("done events = %s, want response.incomplete", string(doneEvents))
			}

			aggregated, err := inbound.GetInternalResponse(context.Background())
			if err != nil {
				t.Fatalf("GetInternalResponse() error = %v", err)
			}
			if aggregated == nil || len(aggregated.Choices) != 1 {
				t.Fatalf("aggregated response malformed: %#v", aggregated)
			}
			if aggregated.Choices[0].FinishReason == nil || *aggregated.Choices[0].FinishReason != tt.wantFinishReason {
				t.Fatalf("finish reason = %#v, want %s", aggregated.Choices[0].FinishReason, tt.wantFinishReason)
			}

			if len(roundTripAggregated.Choices) != 1 {
				t.Fatalf("round-trip aggregated choices = %d, want 1", len(roundTripAggregated.Choices))
			}
			if roundTripAggregated.Choices[0].FinishReason == nil || *roundTripAggregated.Choices[0].FinishReason != tt.wantFinishReason {
				t.Fatalf("round-trip finish reason = %#v, want %s", roundTripAggregated.Choices[0].FinishReason, tt.wantFinishReason)
			}
			if roundTripAggregated.Usage != nil {
				t.Fatalf("round-trip usage = %#v, want nil before usage chunk", roundTripAggregated.Usage)
			}
		})
	}
}

func TestResponsesStreamRoundTripMapsIncompleteAndFailedTerminalEvents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		finishReason         string
		wantTerminalType     string
		wantFinishReason     string
		wantIncompleteReason string
	}{
		{name: "length", finishReason: "length", wantTerminalType: "response.incomplete", wantFinishReason: "length"},
		{name: "content_filter", finishReason: "content_filter", wantTerminalType: "response.incomplete", wantFinishReason: "content_filter", wantIncompleteReason: "content_filter"},
		{name: "error", finishReason: "error", wantTerminalType: "response.failed", wantFinishReason: "error"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			inbound := &ResponseInbound{}
			outbound := &outboundopenai.ResponseOutbound{}
			text := "terminal"

			firstEvents, err := inbound.TransformStream(context.Background(), &model.InternalLLMResponse{
				ID:      "resp-" + tt.finishReason,
				Object:  "chat.completion.chunk",
				Created: 11,
				Model:   "gpt-resp",
				Choices: []model.Choice{{
					Index: 0,
					Delta: &model.Message{Role: "assistant", Content: model.MessageContent{Content: &text}},
				}},
			})
			if err != nil {
				t.Fatalf("TransformStream() first chunk error = %v", err)
			}

			secondEvents, err := inbound.TransformStream(context.Background(), &model.InternalLLMResponse{
				Object: "chat.completion.chunk",
				Usage:  &model.Usage{PromptTokens: 2, CompletionTokens: 1, TotalTokens: 3},
				Choices: []model.Choice{{
					Index:        0,
					FinishReason: &tt.finishReason,
				}},
			})
			if err != nil {
				t.Fatalf("TransformStream() second chunk error = %v", err)
			}

			var sawTerminal bool
			aggregated := &model.InternalLLMResponse{Object: "chat.completion"}
			for _, payload := range [][]byte{firstEvents, secondEvents} {
				for _, line := range strings.Split(string(payload), "\n") {
					line = strings.TrimSpace(line)
					if !strings.HasPrefix(line, "data: ") {
						continue
					}
					raw := []byte(strings.TrimPrefix(line, "data: "))

					var event ResponsesStreamEvent
					if err := json.Unmarshal(raw, &event); err == nil && event.Type == tt.wantTerminalType {
						sawTerminal = true
						if tt.wantIncompleteReason != "" {
							if event.Response == nil || event.Response.IncompleteDetails == nil || event.Response.IncompleteDetails.Reason != tt.wantIncompleteReason {
								t.Fatalf("incomplete details = %#v, want %s", event.Response, tt.wantIncompleteReason)
							}
						}
					}

					chunk, err := outbound.TransformStream(context.Background(), raw)
					if err != nil {
						t.Fatalf("TransformStream() outbound error = %v", err)
					}
					if chunk != nil {
						model.MergeStreamingResponseAggregate(aggregated, chunk)
					}
				}
			}

			if !sawTerminal {
				t.Fatalf("did not observe %s in stream", tt.wantTerminalType)
			}
			if len(aggregated.Choices) != 1 {
				t.Fatalf("aggregated choices = %d, want 1", len(aggregated.Choices))
			}
			if aggregated.Choices[0].FinishReason == nil || *aggregated.Choices[0].FinishReason != tt.wantFinishReason {
				t.Fatalf("finish reason = %#v, want %s", aggregated.Choices[0].FinishReason, tt.wantFinishReason)
			}
			if aggregated.Usage == nil || aggregated.Usage.TotalTokens != 3 {
				t.Fatalf("usage = %#v, want total tokens 3", aggregated.Usage)
			}
		})
	}
}
