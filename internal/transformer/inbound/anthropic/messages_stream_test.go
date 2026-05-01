package anthropic

import (
	"context"
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/model"
)

func TestMessagesInboundAggregatesStreamWithoutChunkRetention(t *testing.T) {
	t.Parallel()

	inbound := &MessagesInbound{}
	text1 := "foo"
	text2 := "bar"
	finish := "stop"

	_, err := inbound.TransformStream(context.Background(), &model.InternalLLMResponse{
		ID:      "resp-3",
		Object:  "chat.completion.chunk",
		Created: 3,
		Model:   "claude-test",
		Choices: []model.Choice{{
			Index: 0,
			Delta: &model.Message{Role: "assistant", Content: model.MessageContent{Content: &text1}},
		}},
	})
	if err != nil {
		t.Fatalf("TransformStream() first chunk error = %v", err)
	}
	_, err = inbound.TransformStream(context.Background(), &model.InternalLLMResponse{
		Object: "chat.completion.chunk",
		Usage:  &model.Usage{PromptTokens: 4, CompletionTokens: 5, TotalTokens: 9},
		Choices: []model.Choice{{
			Index:        0,
			Delta:        &model.Message{Content: model.MessageContent{Content: &text2}},
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
	if got := *resp.Choices[0].Message.Content.Content; got != "foobar" {
		t.Fatalf("content = %q, want foobar", got)
	}
	if inbound.streamResult == nil {
		t.Fatal("streamResult = nil, want aggregated state retained")
	}
}

func TestMessagesInboundAggregatesReasoningSignatureAcrossStreamChunks(t *testing.T) {
	t.Parallel()

	inbound := &MessagesInbound{}
	reason1 := "think-"
	reason2 := "more"
	sig1 := "sig-"
	sig2 := "done"
	finish := "stop"

	_, err := inbound.TransformStream(context.Background(), &model.InternalLLMResponse{
		ID:      "resp-4",
		Object:  "chat.completion.chunk",
		Created: 4,
		Model:   "claude-test",
		Choices: []model.Choice{{
			Index: 0,
			Delta: &model.Message{Role: "assistant", ReasoningContent: &reason1, ReasoningSignature: &sig1},
		}},
	})
	if err != nil {
		t.Fatalf("TransformStream() first chunk error = %v", err)
	}
	_, err = inbound.TransformStream(context.Background(), &model.InternalLLMResponse{
		Object: "chat.completion.chunk",
		Usage:  &model.Usage{PromptTokens: 1, CompletionTokens: 1, TotalTokens: 2},
		Choices: []model.Choice{{
			Index:        0,
			Delta:        &model.Message{ReasoningContent: &reason2, ReasoningSignature: &sig2},
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
	if got := *resp.Choices[0].Message.ReasoningContent; got != "think-more" {
		t.Fatalf("reasoning = %q, want think-more", got)
	}
	if resp.Choices[0].Message.ReasoningSignature == nil {
		t.Fatal("reasoning signature = nil, want aggregated signature")
	}
	if got := *resp.Choices[0].Message.ReasoningSignature; got != "sig-done" {
		t.Fatalf("reasoning signature = %q, want sig-done", got)
	}
}
