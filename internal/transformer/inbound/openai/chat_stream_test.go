package openai

import (
	"context"
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/model"
)

func TestChatInboundAggregatesStreamWithoutChunkRetention(t *testing.T) {
	t.Parallel()

	inbound := &ChatInbound{}
	text1 := "hel"
	text2 := "lo"
	reason1 := "think-"
	reason2 := "more"
	finish := "stop"

	_, err := inbound.TransformStream(context.Background(), &model.InternalLLMResponse{
		ID:      "resp-1",
		Object:  "chat.completion.chunk",
		Created: 1,
		Model:   "gpt-test",
		Choices: []model.Choice{{
			Index: 0,
			Delta: &model.Message{Role: "assistant", Content: model.MessageContent{Content: &text1}, ReasoningContent: &reason1},
		}},
	})
	if err != nil {
		t.Fatalf("TransformStream() first chunk error = %v", err)
	}
	_, err = inbound.TransformStream(context.Background(), &model.InternalLLMResponse{
		Object: "chat.completion.chunk",
		Usage:  &model.Usage{PromptTokens: 1, CompletionTokens: 2, TotalTokens: 3},
		Choices: []model.Choice{{
			Index:        0,
			Delta:        &model.Message{Content: model.MessageContent{Content: &text2}, ReasoningContent: &reason2},
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
	if got := *resp.Choices[0].Message.Content.Content; got != "hello" {
		t.Fatalf("content = %q, want hello", got)
	}
	if got := *resp.Choices[0].Message.ReasoningContent; got != "think-more" {
		t.Fatalf("reasoning = %q, want think-more", got)
	}
	if inbound.streamResult == nil {
		t.Fatal("streamResult = nil, want aggregated state retained")
	}
}
