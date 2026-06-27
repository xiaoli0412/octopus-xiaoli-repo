//go:build !rust
// +build !rust

package rustbridge

import (
	"encoding/json"
	"errors"

	transformerModel "github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/tokenizer"
)

// Enabled reports whether the Rust FFI backend was compiled in.
func Enabled() bool { return false }

// CountTokens uses the Go tiktoken-go implementation.
func CountTokens(content, model string) int {
	return tokenizer.CountTokensWithModel(content, model)
}

func countTokensGo(content, model string) int {
	return tokenizer.CountTokensWithModel(content, model)
}

func countTokensRust(content, model string) int { return 0 }

// ExtractModel returns the model name from a request/response JSON body.
func ExtractModel(body string) (string, error) {
	var req struct {
		Model string `json:"model"`
	}
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		return "", err
	}
	out, err := json.Marshal(map[string]string{"model": req.Model})
	return string(out), err
}

// ExtractUsage returns a usage JSON object from a response body.
func ExtractUsage(body string) (string, error) {
	var resp struct {
		Usage *struct {
			PromptTokens     int64 `json:"prompt_tokens"`
			CompletionTokens int64 `json:"completion_tokens"`
			TotalTokens      int64 `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		return "", err
	}
	if resp.Usage == nil {
		return "{}", nil
	}
	out, err := json.Marshal(struct {
		PromptTokens     int64 `json:"prompt_tokens"`
		CompletionTokens int64 `json:"completion_tokens"`
		TotalTokens      int64 `json:"total_tokens"`
	}{
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		TotalTokens:      resp.Usage.TotalTokens,
	})
	return string(out), err
}

// SSEAggregate merges a streaming chunk into an aggregate response JSON using
// the existing Go implementation.
func SSEAggregate(aggregate, chunk string) (string, error) {
	var agg, ck transformerModel.InternalLLMResponse
	if err := json.Unmarshal([]byte(aggregate), &agg); err != nil {
		return "", err
	}
	if err := json.Unmarshal([]byte(chunk), &ck); err != nil {
		return "", err
	}
	transformerModel.MergeStreamingResponseAggregate(&agg, &ck)
	out, err := json.Marshal(&agg)
	return string(out), err
}

// ErrRustNotEnabled is returned when Rust-only helpers are invoked without the
// rust build tag.
var ErrRustNotEnabled = errors.New("rust backend not enabled")
