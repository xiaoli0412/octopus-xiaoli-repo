package tokenizer

import (
	"strings"

	"github.com/tiktoken-go/tokenizer/codec"
)

func CountTokens(content, model string) int {
	// TODO 更多模型
	enc := codec.NewO200kBase()
	tc, err := enc.Count(content)
	if err != nil {
		return 0
	}
	return tc
}

// CountTokensWithModel selects the tiktoken encoding based on the model name.
// It is used by the Rust FFI A/B switch to provide a fair Go baseline.
func CountTokensWithModel(content, model string) int {
	enc := encodingForModel(model)
	tc, err := enc.Count(content)
	if err != nil {
		return 0
	}
	return tc
}

func encodingForModel(model string) *codec.Codec {
	m := strings.ToLower(model)
	if strings.Contains(m, "cl100k") ||
		strings.Contains(m, "text-embedding") ||
		(strings.Contains(m, "gpt-4") && !strings.Contains(m, "gpt-4o")) ||
		strings.Contains(m, "gpt-3.5") {
		return codec.NewCl100kBase()
	}
	return codec.NewO200kBase()
}
