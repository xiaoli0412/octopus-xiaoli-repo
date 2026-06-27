package streamhelper

import (
	"encoding/json"
	"os"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/rustbridge"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/model"
)

const envDisableRustStreamAggregate = "OCTOPUS_RUST_STREAM_AGGREGATE"

// rustStreamAggregateEnabled reports whether the Rust-accelerated SSE aggregate
// path is enabled. It is enabled unless the env var is explicitly set to "0",
// "false", "FALSE" or "False".
func rustStreamAggregateEnabled() bool {
	v := os.Getenv(envDisableRustStreamAggregate)
	return v != "0" && v != "false" && v != "FALSE" && v != "False"
}

// MergeStreamingChunk folds chunk into aggregate using rustbridge.SSEAggregate
// when enabled, falling back to model.MergeStreamingResponseAggregate when
// disabled or when the Rust path fails.
//
// The chunk must be a decoded InternalLLMResponse (not an SSE "data:" line).
// Either aggregate or chunk may be nil; [DONE] chunks are ignored.
func MergeStreamingChunk(aggregate, chunk *model.InternalLLMResponse) *model.InternalLLMResponse {
	if chunk == nil || chunk.Object == "[DONE]" {
		return aggregate
	}
	if aggregate == nil {
		aggregate = &model.InternalLLMResponse{Object: "chat.completion"}
	}

	if rustStreamAggregateEnabled() {
		merged, ok := mergeWithRust(aggregate, chunk)
		if ok {
			return merged
		}
	}

	model.MergeStreamingResponseAggregate(aggregate, chunk)
	return aggregate
}

func mergeWithRust(aggregate, chunk *model.InternalLLMResponse) (*model.InternalLLMResponse, bool) {
	aggJSON, err := json.Marshal(aggregate)
	if err != nil {
		return nil, false
	}
	chunkJSON, err := json.Marshal(chunk)
	if err != nil {
		return nil, false
	}

	out, err := rustbridge.SSEAggregate(string(aggJSON), string(chunkJSON))
	if err != nil {
		return nil, false
	}

	var merged model.InternalLLMResponse
	if err := json.Unmarshal([]byte(out), &merged); err != nil {
		return nil, false
	}
	return &merged, true
}
