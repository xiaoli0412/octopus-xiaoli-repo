package price

import (
	"strings"
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

func TestDecodeLLMPriceResponseRejectsOversizedBody(t *testing.T) {
	payload := strings.Repeat("a", int(maxLLMPriceResponseBytes)+1)

	err := decodeLLMPriceResponse(strings.NewReader(payload), &map[string]any{})
	if err == nil || !strings.Contains(err.Error(), "response body too large") {
		t.Fatalf("decodeLLMPriceResponse() error = %v, want size limit error", err)
	}
}

func TestDecodeLLMPriceResponseParsesValidPayload(t *testing.T) {
	type modelCostPayload struct {
		ID   string         `json:"id"`
		Cost model.LLMPrice `json:"cost"`
	}
	type providerPayload struct {
		Models map[string]modelCostPayload `json:"models"`
	}

	payload := map[string]providerPayload{}
	err := decodeLLMPriceResponse(strings.NewReader(`{"openai":{"models":{"gpt-4.1":{"id":"gpt-4.1","cost":{"input":1.23,"output":4.56}}}}}`), &payload)
	if err != nil {
		t.Fatalf("decodeLLMPriceResponse() error = %v", err)
	}
	if payload["openai"].Models["gpt-4.1"].Cost.Input != 1.23 {
		t.Fatalf("decoded input cost = %v, want %v", payload["openai"].Models["gpt-4.1"].Cost.Input, 1.23)
	}
	if payload["openai"].Models["gpt-4.1"].Cost.Output != 4.56 {
		t.Fatalf("decoded output cost = %v, want %v", payload["openai"].Models["gpt-4.1"].Cost.Output, 4.56)
	}
}
