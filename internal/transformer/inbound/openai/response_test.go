package openai

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/model"
)

func TestResponseInboundTransformRequestPreservesJSONSchemaFormat(t *testing.T) {
	t.Parallel()

	body := []byte(`{
		"model":"gpt-4.1",
		"input":"hello",
		"text":{
			"format":{
				"type":"json_schema",
				"schema":{"type":"object","properties":{"answer":{"type":"string"}},"required":["answer"]}
			}
		}
	}`)

	inbound := &ResponseInbound{}
	request, err := inbound.TransformRequest(context.Background(), body)
	if err != nil {
		t.Fatalf("TransformRequest() error = %v", err)
	}
	if request.ResponseFormat == nil {
		t.Fatal("ResponseFormat = nil, want json_schema format")
	}
	if request.ResponseFormat.Type != "json_schema" {
		t.Fatalf("ResponseFormat.Type = %q, want %q", request.ResponseFormat.Type, "json_schema")
	}
	if string(request.ResponseFormat.JSONSchema) == "" {
		t.Fatal("ResponseFormat.JSONSchema is empty, want preserved schema")
	}
}

func TestResponseInboundTransformResponsePreservesMixedTextImageOutputOrder(t *testing.T) {
	t.Parallel()

	inbound := &ResponseInbound{}
	text1 := "hello"
	text2 := "world"
	imageURL := "data:image/png;base64,ZmFrZQ=="

	body, err := inbound.TransformResponse(context.Background(), &model.InternalLLMResponse{
		ID:      "resp-mixed",
		Object:  "chat.completion",
		Model:   "gpt-4.1",
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

	var resp ResponsesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if len(resp.Output) != 3 {
		t.Fatalf("len(Output) = %d, want 3", len(resp.Output))
	}
	if resp.Output[0].Type != "message" || resp.Output[0].Content == nil || len(resp.Output[0].Content.Items) != 1 || resp.Output[0].Content.Items[0].Text == nil || *resp.Output[0].Content.Items[0].Text != "hello" {
		t.Fatalf("first output item = %#v, want leading text message", resp.Output[0])
	}
	if resp.Output[1].Type != "image_generation_call" || resp.Output[1].Result == nil || *resp.Output[1].Result != "ZmFrZQ==" {
		t.Fatalf("second output item = %#v, want image generation item", resp.Output[1])
	}
	if resp.Output[1].OutputFormat == nil || *resp.Output[1].OutputFormat != "png" {
		t.Fatalf("image output format = %#v, want png", resp.Output[1].OutputFormat)
	}
	if resp.Output[2].Type != "message" || resp.Output[2].Content == nil || len(resp.Output[2].Content.Items) != 1 || resp.Output[2].Content.Items[0].Text == nil || *resp.Output[2].Content.Items[0].Text != "world" {
		t.Fatalf("third output item = %#v, want trailing text message", resp.Output[2])
	}
}

func TestResponseInboundTransformResponsePreservesRefusalOutput(t *testing.T) {
	t.Parallel()

	inbound := &ResponseInbound{}

	body, err := inbound.TransformResponse(context.Background(), &model.InternalLLMResponse{
		ID:      "resp-refusal",
		Object:  "chat.completion",
		Model:   "gpt-4.1",
		Created: 2,
		Choices: []model.Choice{{
			Index: 0,
			Message: &model.Message{
				Role:    "assistant",
				Refusal: "cannot comply",
			},
		}},
	})
	if err != nil {
		t.Fatalf("TransformResponse() error = %v", err)
	}

	var resp ResponsesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if len(resp.Output) != 1 {
		t.Fatalf("len(Output) = %d, want 1", len(resp.Output))
	}
	if resp.Output[0].Type != "message" || resp.Output[0].Refusal == nil || *resp.Output[0].Refusal != "cannot comply" {
		t.Fatalf("refusal output item = %#v, want assistant refusal message", resp.Output[0])
	}
}

func TestResponseInboundTransformResponsePreservesContentFilterIncompleteReason(t *testing.T) {
	t.Parallel()

	inbound := &ResponseInbound{}
	text := "filtered"
	finish := "content_filter"

	body, err := inbound.TransformResponse(context.Background(), &model.InternalLLMResponse{
		ID:      "resp-content-filter",
		Object:  "chat.completion",
		Model:   "gpt-4.1",
		Created: 3,
		Choices: []model.Choice{{
			Index: 0,
			Message: &model.Message{
				Role: "assistant",
				Content: model.MessageContent{Content: &text},
			},
			FinishReason: &finish,
		}},
	})
	if err != nil {
		t.Fatalf("TransformResponse() error = %v", err)
	}

	var resp ResponsesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if resp.Status == nil || *resp.Status != "incomplete" {
		t.Fatalf("Status = %#v, want incomplete", resp.Status)
	}
	if resp.IncompleteDetails == nil || resp.IncompleteDetails.Reason != "content_filter" {
		t.Fatalf("IncompleteDetails = %#v, want content_filter", resp.IncompleteDetails)
	}
	if len(resp.Output) != 1 || resp.Output[0].Type != "message" {
		t.Fatalf("Output = %#v, want one message item", resp.Output)
	}
}
