package gemini

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	tmodel "github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/model"
)

func TestMessagesOutboundTransformResponseReturnsStructuredHTTPError(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusTooManyRequests,
		Body:       io.NopCloser(strings.NewReader(`{"error":{"code":429,"message":"quota exceeded","status":"RESOURCE_EXHAUSTED"}}`)),
	}

	_, err := (&MessagesOutbound{}).TransformResponse(context.Background(), resp)
	if err == nil {
		t.Fatal("TransformResponse() error = nil, want structured response error")
	}
	respErr, ok := err.(*tmodel.ResponseError)
	if !ok {
		t.Fatalf("TransformResponse() error type = %T, want *ResponseError", err)
	}
	if respErr.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("ResponseError.StatusCode = %d, want %d", respErr.StatusCode, http.StatusTooManyRequests)
	}
	if respErr.Detail.Message != "quota exceeded" {
		t.Fatalf("ResponseError.Detail.Message = %q, want %q", respErr.Detail.Message, "quota exceeded")
	}
	if respErr.Detail.Type != "RESOURCE_EXHAUSTED" {
		t.Fatalf("ResponseError.Detail.Type = %q, want %q", respErr.Detail.Type, "RESOURCE_EXHAUSTED")
	}
}

func TestConvertLLMToGeminiRequestBuildsResponseSchemaFromJSONSchema(t *testing.T) {
	t.Parallel()

	req := &tmodel.InternalLLMRequest{
		Model: "gemini-2.5-flash",
		Messages: []tmodel.Message{{
			Role:    "user",
			Content: tmodel.MessageContent{Content: ptrString("return json")},
		}},
		ResponseFormat: &tmodel.ResponseFormat{
			Type:       "json_schema",
			JSONSchema: json.RawMessage(`{"type":"object","properties":{"answer":{"type":"string"},"items":{"type":"array","items":{"type":"integer"}}},"required":["answer"]}`),
		},
	}

	geminiReq := ConvertLLMToGeminiRequest(req)
	if geminiReq.GenerationConfig == nil {
		t.Fatal("GenerationConfig = nil, want populated config")
	}
	if geminiReq.GenerationConfig.ResponseMimeType != "application/json" {
		t.Fatalf("ResponseMimeType = %q, want %q", geminiReq.GenerationConfig.ResponseMimeType, "application/json")
	}
	if geminiReq.GenerationConfig.ResponseSchema == nil {
		t.Fatal("ResponseSchema = nil, want converted schema")
	}
	if geminiReq.GenerationConfig.ResponseSchema.Type != "OBJECT" {
		t.Fatalf("ResponseSchema.Type = %q, want %q", geminiReq.GenerationConfig.ResponseSchema.Type, "OBJECT")
	}
	answer := geminiReq.GenerationConfig.ResponseSchema.Properties["answer"]
	if answer == nil || answer.Type != "STRING" {
		t.Fatalf("answer schema = %#v, want STRING property", answer)
	}
	items := geminiReq.GenerationConfig.ResponseSchema.Properties["items"]
	if items == nil || items.Type != "ARRAY" || items.Items == nil || items.Items.Type != "INTEGER" {
		t.Fatalf("items schema = %#v, want ARRAY of INTEGER", items)
	}
}

func ptrString(value string) *string {
	return &value
}
