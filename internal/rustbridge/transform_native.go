//go:build !rust
// +build !rust

package rustbridge

import "encoding/json"

// TransformOpenAIChatRequest is a pure-Go fallback for protocol transform.
func TransformOpenAIChatRequest(input, target string) (string, error) {
	var req map[string]any
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		return "", err
	}
	if msgs, ok := req["messages"].([]any); ok {
		for _, m := range msgs {
			if msg, ok := m.(map[string]any); ok {
				if role, _ := msg["role"].(string); role == "developer" {
					msg["role"] = "system"
				}
			}
		}
	}
	if stream, ok := req["stream"].(bool); ok && stream {
		opts, _ := req["stream_options"].(map[string]any)
		if opts == nil {
			opts = map[string]any{}
		}
		if _, ok := opts["include_usage"]; !ok {
			opts["include_usage"] = true
		}
		req["stream_options"] = opts
	}
	_ = target
	out, err := json.Marshal(req)
	return string(out), err
}

// TransformOpenAIResponse is a pure-Go fallback for response transform.
func TransformOpenAIResponse(input, target string) (string, error) {
	_ = target
	return input, nil
}

// TransformEmbeddingRequest is a pure-Go fallback for embedding transform.
func TransformEmbeddingRequest(input, target string) (string, error) {
	_ = target
	return input, nil
}
