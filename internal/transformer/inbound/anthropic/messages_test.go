package anthropic

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/model"
)

func TestMessagesInboundTransformRequestPreservesImageToolResultParts(t *testing.T) {
	t.Parallel()

	body := []byte(`{
		"model":"claude-3-5-sonnet",
		"max_tokens":128,
		"messages":[
			{
				"role":"user",
				"content":[
					{
						"type":"tool_result",
						"tool_use_id":"tool-1",
						"content":[
							{"type":"text","text":"first result"},
							{"type":"image","source":{"type":"base64","media_type":"image/png","data":"ZmFrZQ=="}}
						]
					}
				]
			}
		]
	}`)

	inbound := &MessagesInbound{}
	request, err := inbound.TransformRequest(context.Background(), body)
	if err != nil {
		t.Fatalf("TransformRequest() error = %v", err)
	}
	if len(request.Messages) != 1 {
		t.Fatalf("len(Messages) = %d, want 1", len(request.Messages))
	}
	msg := request.Messages[0]
	if msg.Role != "tool" {
		t.Fatalf("Role = %q, want %q", msg.Role, "tool")
	}
	if msg.ToolCallID == nil || *msg.ToolCallID != "tool-1" {
		t.Fatalf("ToolCallID = %#v, want tool-1", msg.ToolCallID)
	}
	if len(msg.Content.MultipleContent) != 2 {
		t.Fatalf("len(MultipleContent) = %d, want 2", len(msg.Content.MultipleContent))
	}
	if msg.Content.MultipleContent[0].Type != "text" {
		t.Fatalf("first content type = %q, want text", msg.Content.MultipleContent[0].Type)
	}
	if msg.Content.MultipleContent[1].Type != "image_url" {
		t.Fatalf("second content type = %q, want image_url", msg.Content.MultipleContent[1].Type)
	}
	if msg.Content.MultipleContent[1].ImageURL == nil || msg.Content.MultipleContent[1].ImageURL.URL != "data:image/png;base64,ZmFrZQ==" {
		t.Fatalf("image url = %#v, want data url", msg.Content.MultipleContent[1].ImageURL)
	}
}

func TestMessagesInboundTransformResponsePreservesMixedTextImageOrder(t *testing.T) {
	t.Parallel()

	inbound := &MessagesInbound{}
	text1 := "hello"
	text2 := "world"
	imageURL := "data:image/png;base64,ZmFrZQ=="

	body, err := inbound.TransformResponse(context.Background(), &model.InternalLLMResponse{
		ID:      "msg-mixed",
		Object:  "chat.completion",
		Model:   "claude-3-7-sonnet",
		Created: 1,
		Choices: []model.Choice{{
			Index: 0,
			Message: &model.Message{
				Role: "assistant",
				Content: model.MessageContent{MultipleContent: []model.MessageContentPart{
					{Type: "text", Text: &text1},
					{Type: "image_url", ImageURL: &model.ImageURL{URL: imageURL}},
					{Type: "text", Text: &text2},
				}},
			},
		}},
	})
	if err != nil {
		t.Fatalf("TransformResponse() error = %v", err)
	}

	var resp Message
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if len(resp.Content) != 3 {
		t.Fatalf("len(Content) = %d, want 3", len(resp.Content))
	}
	if resp.Content[0].Type != "text" || resp.Content[0].Text == nil || *resp.Content[0].Text != "hello" {
		t.Fatalf("first content block = %#v, want leading text", resp.Content[0])
	}
	if resp.Content[1].Type != "image" || resp.Content[1].Source == nil || resp.Content[1].Source.Type != "base64" || resp.Content[1].Source.MediaType != "image/png" || resp.Content[1].Source.Data != "ZmFrZQ==" {
		t.Fatalf("second content block = %#v, want base64 image", resp.Content[1])
	}
	if resp.Content[2].Type != "text" || resp.Content[2].Text == nil || *resp.Content[2].Text != "world" {
		t.Fatalf("third content block = %#v, want trailing text", resp.Content[2])
	}
}
