package model

import "testing"

func TestMergeStreamingResponseAggregate_MergesChoiceMessageAndDeltaFromSameChunk(t *testing.T) {
	t.Parallel()

	result := &InternalLLMResponse{Object: "chat.completion"}
	content := "answer"
	reasoning := "think"
	toolArgs := "{\"x\":1}"

	MergeStreamingResponseAggregate(result, &InternalLLMResponse{
		Object: "chat.completion.chunk",
		Choices: []Choice{{
			Index: 0,
			Message: &Message{
				Role:    "assistant",
				Content: MessageContent{Content: &content},
			},
			Delta: &Message{
				ReasoningContent: &reasoning,
				ToolCalls: []ToolCall{{
					Index: 0,
					ID:    "call-1",
					Type:  "function",
					Function: FunctionCall{
						Name:      "calc",
						Arguments: toolArgs,
					},
				}},
			},
		}},
	})

	if len(result.Choices) != 1 || result.Choices[0].Message == nil {
		t.Fatalf("aggregated response malformed: %#v", result)
	}
	message := result.Choices[0].Message
	if message.Content.Content == nil || *message.Content.Content != "answer" {
		t.Fatalf("content = %#v, want answer", message.Content.Content)
	}
	if message.ReasoningContent == nil || *message.ReasoningContent != "think" {
		t.Fatalf("reasoning = %#v, want think", message.ReasoningContent)
	}
	if len(message.ToolCalls) != 1 {
		t.Fatalf("tool call count = %d, want 1", len(message.ToolCalls))
	}
	if got := message.ToolCalls[0].Function.Arguments; got != toolArgs {
		t.Fatalf("tool arguments = %q, want %q", got, toolArgs)
	}
}

func TestMergeStreamingResponseAggregate_PromotesTextToMultipleContentWhenImageArrives(t *testing.T) {
	t.Parallel()

	result := &InternalLLMResponse{Object: "chat.completion"}
	text1 := "hello"
	text2 := "world"
	imageURL := "data:image/png;base64,ZmFrZQ=="

	MergeStreamingResponseAggregate(result, &InternalLLMResponse{
		Object: "chat.completion.chunk",
		Choices: []Choice{{
			Index: 0,
			Message: &Message{
				Role:    "assistant",
				Content: MessageContent{Content: &text1},
			},
		}},
	})
	MergeStreamingResponseAggregate(result, &InternalLLMResponse{
		Object: "chat.completion.chunk",
		Choices: []Choice{{
			Index: 0,
			Delta: &Message{
				Images: []MessageContentPart{{
					Type: "image_url",
					ImageURL: &ImageURL{
						URL: imageURL,
					},
				}},
			},
		}},
	})
	MergeStreamingResponseAggregate(result, &InternalLLMResponse{
		Object: "chat.completion.chunk",
		Choices: []Choice{{
			Index: 0,
			Message: &Message{
				Content: MessageContent{Content: &text2},
			},
		}},
	})

	if len(result.Choices) != 1 || result.Choices[0].Message == nil {
		t.Fatalf("aggregated response malformed: %#v", result)
	}
	message := result.Choices[0].Message
	if message.Content.Content != nil {
		t.Fatalf("content string = %#v, want nil after multipart promotion", message.Content.Content)
	}
	if len(message.Content.MultipleContent) != 3 {
		t.Fatalf("multiple content count = %d, want 3", len(message.Content.MultipleContent))
	}
	if message.Content.MultipleContent[0].Type != "text" || message.Content.MultipleContent[0].Text == nil || *message.Content.MultipleContent[0].Text != "hello" {
		t.Fatalf("first part = %#v, want hello text part", message.Content.MultipleContent[0])
	}
	if message.Content.MultipleContent[1].Type != "image_url" || message.Content.MultipleContent[1].ImageURL == nil || message.Content.MultipleContent[1].ImageURL.URL != imageURL {
		t.Fatalf("second part = %#v, want image part", message.Content.MultipleContent[1])
	}
	if message.Content.MultipleContent[2].Type != "text" || message.Content.MultipleContent[2].Text == nil || *message.Content.MultipleContent[2].Text != "world" {
		t.Fatalf("third part = %#v, want world text part", message.Content.MultipleContent[2])
	}
}

func TestMergeStreamingResponseAggregate_AppendsRefusalAcrossStreamChunks(t *testing.T) {
	t.Parallel()

	result := &InternalLLMResponse{Object: "chat.completion"}
	finish := "stop"

	MergeStreamingResponseAggregate(result, &InternalLLMResponse{
		Object: "chat.completion.chunk",
		Choices: []Choice{{
			Index: 0,
			Delta: &Message{Role: "assistant", Refusal: "cannot "},
		}},
	})
	MergeStreamingResponseAggregate(result, &InternalLLMResponse{
		Object: "chat.completion.chunk",
		Choices: []Choice{{
			Index:        0,
			Delta:        &Message{Refusal: "comply"},
			FinishReason: &finish,
		}},
	})

	if len(result.Choices) != 1 || result.Choices[0].Message == nil {
		t.Fatalf("aggregated response malformed: %#v", result)
	}
	message := result.Choices[0].Message
	if got := message.Refusal; got != "cannot comply" {
		t.Fatalf("refusal = %q, want cannot comply", got)
	}
	if result.Choices[0].FinishReason == nil || *result.Choices[0].FinishReason != "stop" {
		t.Fatalf("finish reason = %#v, want stop", result.Choices[0].FinishReason)
	}
}

func TestMergeStreamingResponseAggregate_AppendsLogprobsAcrossStreamChunks(t *testing.T) {
	t.Parallel()

	result := &InternalLLMResponse{Object: "chat.completion"}
	firstText := "hel"
	secondText := "lo"
	finish := "stop"

	firstChunk := &InternalLLMResponse{
		Object: "chat.completion.chunk",
		Choices: []Choice{{
			Index: 0,
			Delta: &Message{
				Role: "assistant",
				Content: MessageContent{
					Content: &firstText,
				},
			},
			Logprobs: &LogprobsContent{Content: []TokenLogprob{{
				Token:   "hel",
				Logprob: -0.11,
				Bytes:   []int{104, 101, 108},
				TopLogprobs: []TopLogprob{{
					Token:   "hey",
					Logprob: -0.42,
					Bytes:   []int{104, 101, 121},
				}},
			}}},
		}},
	}

	MergeStreamingResponseAggregate(result, firstChunk)
	firstChunk.Choices[0].Logprobs.Content[0].Bytes[0] = 0
	firstChunk.Choices[0].Logprobs.Content[0].TopLogprobs[0].Bytes[0] = 1

	MergeStreamingResponseAggregate(result, &InternalLLMResponse{
		Object: "chat.completion.chunk",
		Choices: []Choice{{
			Index: 0,
			Delta: &Message{Content: MessageContent{Content: &secondText}},
			Logprobs: &LogprobsContent{Content: []TokenLogprob{{
				Token:   "lo",
				Logprob: -0.07,
			}}},
			FinishReason: &finish,
		}},
	})

	if len(result.Choices) != 1 || result.Choices[0].Message == nil {
		t.Fatalf("aggregated response malformed: %#v", result)
	}
	if result.Choices[0].Logprobs == nil {
		t.Fatal("logprobs = nil, want aggregated content")
	}
	if len(result.Choices[0].Logprobs.Content) != 2 {
		t.Fatalf("logprobs token count = %d, want 2", len(result.Choices[0].Logprobs.Content))
	}
	if got := result.Choices[0].Logprobs.Content[0].Token; got != "hel" {
		t.Fatalf("first token = %q, want hel", got)
	}
	if got := result.Choices[0].Logprobs.Content[0].Bytes; len(got) != 3 || got[0] != 104 {
		t.Fatalf("first token bytes = %#v, want preserved deep copy", got)
	}
	if got := result.Choices[0].Logprobs.Content[0].TopLogprobs; len(got) != 1 || len(got[0].Bytes) != 3 || got[0].Bytes[0] != 104 {
		t.Fatalf("top logprobs = %#v, want preserved deep copy", got)
	}
	if got := result.Choices[0].Logprobs.Content[1].Token; got != "lo" {
		t.Fatalf("second token = %q, want lo", got)
	}
	if result.Choices[0].FinishReason == nil || *result.Choices[0].FinishReason != "stop" {
		t.Fatalf("finish reason = %#v, want stop", result.Choices[0].FinishReason)
	}
	if result.Choices[0].Message.Content.Content == nil || *result.Choices[0].Message.Content.Content != "hello" {
		t.Fatalf("content = %#v, want hello", result.Choices[0].Message.Content.Content)
	}
}

func TestMergeStreamingResponseAggregate_AppendsToolCallNameAndArgumentsAcrossStreamChunks(t *testing.T) {
	t.Parallel()

	result := &InternalLLMResponse{Object: "chat.completion"}
	finish := "tool_calls"
	arg1 := "{\"city\":"
	arg2 := "\"Tokyo\"}"

	MergeStreamingResponseAggregate(result, &InternalLLMResponse{
		Object: "chat.completion.chunk",
		Choices: []Choice{{
			Index: 0,
			Delta: &Message{Role: "assistant", ToolCalls: []ToolCall{{
				Index: 0,
				ID:    "call-1",
				Type:  "function",
				Function: FunctionCall{
					Name:      "weather_",
					Arguments: arg1,
				},
			}}},
		}},
	})
	MergeStreamingResponseAggregate(result, &InternalLLMResponse{
		Object: "chat.completion.chunk",
		Choices: []Choice{{
			Index: 0,
			Delta: &Message{ToolCalls: []ToolCall{{
				Index: 0,
				Function: FunctionCall{
					Name:      "lookup",
					Arguments: arg2,
				},
			}}},
			FinishReason: &finish,
		}},
	})

	if len(result.Choices) != 1 || result.Choices[0].Message == nil {
		t.Fatalf("aggregated response malformed: %#v", result)
	}
	message := result.Choices[0].Message
	if len(message.ToolCalls) != 1 {
		t.Fatalf("tool call count = %d, want 1", len(message.ToolCalls))
	}
	if got := message.ToolCalls[0].Function.Name; got != "weather_lookup" {
		t.Fatalf("tool name = %q, want weather_lookup", got)
	}
	if got := message.ToolCalls[0].Function.Arguments; got != "{\"city\":\"Tokyo\"}" {
		t.Fatalf("tool arguments = %q, want merged json", got)
	}
	if result.Choices[0].FinishReason == nil || *result.Choices[0].FinishReason != "tool_calls" {
		t.Fatalf("finish reason = %#v, want tool_calls", result.Choices[0].FinishReason)
	}
}
