package rustbridge

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTransformOpenAIChatRequestOpenAI(t *testing.T) {
	req := `{"model":"gpt-4o","messages":[{"role":"developer","content":"hi"}],"stream":true}`
	out, err := TransformOpenAIChatRequest(req, "openai")
	if err != nil {
		t.Fatalf("transform error: %v", err)
	}
	if !strings.Contains(out, `"role":"system"`) {
		t.Fatalf("expected developer -> system, got %s", out)
	}
	if !strings.Contains(out, `"include_usage":true`) {
		t.Fatalf("expected include_usage, got %s", out)
	}
}

func TestTransformOpenAIChatRequestAnthropic(t *testing.T) {
	if !Enabled() {
		t.Skip("rust backend not enabled")
	}
	req := `{"model":"claude-3-opus","messages":[{"role":"system","content":"sys"},{"role":"user","content":"hello"}],"temperature":0.5,"tools":[{"type":"function","function":{"name":"weather","description":"get weather","parameters":{"type":"object"}}}],"tool_choice":"auto"}`
	out, err := TransformOpenAIChatRequest(req, "anthropic")
	if err != nil {
		t.Fatalf("transform error: %v", err)
	}
	var v map[string]any
	if err := json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := v["system"]; !ok {
		t.Fatalf("expected system field, got %s", out)
	}
	if msgs, ok := v["messages"].([]any); !ok || len(msgs) != 1 {
		t.Fatalf("expected 1 non-system message, got %s", out)
	}
}

func TestTransformOpenAIChatRequestGemini(t *testing.T) {
	if !Enabled() {
		t.Skip("rust backend not enabled")
	}
	req := `{"model":"gemini-1.5","messages":[{"role":"user","content":"hello"}],"max_tokens":100,"temperature":0.7}`
	out, err := TransformOpenAIChatRequest(req, "gemini")
	if err != nil {
		t.Fatalf("transform error: %v", err)
	}
	var v map[string]any
	if err := json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := v["contents"]; !ok {
		t.Fatalf("expected contents field, got %s", out)
	}
	if _, ok := v["generationConfig"]; !ok {
		t.Fatalf("expected generationConfig field, got %s", out)
	}
}

func TestTransformEmbeddingRequestGemini(t *testing.T) {
	if !Enabled() {
		t.Skip("rust backend not enabled")
	}
	req := `{"model":"models/gemini-embedding","input":"hello world"}`
	out, err := TransformEmbeddingRequest(req, "gemini")
	if err != nil {
		t.Fatalf("transform error: %v", err)
	}
	var v map[string]any
	if err := json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := v["content"]; !ok {
		t.Fatalf("expected content field, got %s", out)
	}
}

func TestTransformOpenAIChatRequestInvalidProvider(t *testing.T) {
	req := `{"model":"gpt-4o","messages":[]}`
	if Enabled() {
		_, err := TransformOpenAIChatRequest(req, "unknown")
		if err == nil {
			t.Fatal("expected error for unknown provider")
		}
	}
}

func TestTransformEnvSwitch(t *testing.T) {
	if !Enabled() {
		t.Skip("rust backend not enabled")
	}
	req := `{"model":"gpt-4o","messages":[{"role":"developer","content":"hi"}],"stream":true}`
	rustOut, err := TransformOpenAIChatRequest(req, "openai")
	if err != nil {
		t.Fatalf("rust transform error: %v", err)
	}
	t.Setenv(envDisableRustTransform, "0")
	goOut, err := TransformOpenAIChatRequest(req, "openai")
	if err != nil {
		t.Fatalf("go fallback error: %v", err)
	}
	if rustOut != goOut {
		t.Fatalf("env switch changed output\nrust=%s\ngo=%s", rustOut, goOut)
	}
}
