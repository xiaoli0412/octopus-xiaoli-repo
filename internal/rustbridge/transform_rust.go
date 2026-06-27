//go:build rust
// +build rust

package rustbridge

/*
#cgo windows LDFLAGS: -L${SRCDIR}/../../rust/core/target/release -loctopus_core -lws2_32 -luserenv -ladvapi32 -lntdll -lbcrypt
#cgo linux LDFLAGS: -L${SRCDIR}/../../rust/core/target/release -loctopus_core -lm
#cgo darwin LDFLAGS: -L${SRCDIR}/../../rust/core/target/release -loctopus_core -framework Security -framework CoreFoundation
#include <stdlib.h>
extern int octopus_transform_openai_chat_request(const char* json, const char* target, char** out);
extern int octopus_transform_openai_response(const char* json, const char* target, char** out);
extern int octopus_transform_embedding_request(const char* json, const char* target, char** out);
extern void octopus_free_string(char* s);
*/
import "C"
import (
	"encoding/json"
	"errors"
	"unsafe"
)

// TransformOpenAIChatRequest transforms an OpenAI chat completion request JSON
// for the target provider. Set OCTOPUS_RUST_TRANSFORM=0 to use the Go fallback.
func TransformOpenAIChatRequest(json, target string) (string, error) {
	if !envEnabled(envDisableRustTransform) {
		return transformOpenAIChatRequestGo(json, target)
	}
	return transformOpenAIChatRequestRust(json, target)
}

// TransformOpenAIResponse transforms an OpenAI chat completion response JSON
// for the target provider.
func TransformOpenAIResponse(json, target string) (string, error) {
	if !envEnabled(envDisableRustTransform) {
		return transformOpenAIResponseGo(json, target)
	}
	return transformOpenAIResponseRust(json, target)
}

// TransformEmbeddingRequest transforms an OpenAI embedding request JSON for the
// target provider.
func TransformEmbeddingRequest(json, target string) (string, error) {
	if !envEnabled(envDisableRustTransform) {
		return transformEmbeddingRequestGo(json, target)
	}
	return transformEmbeddingRequestRust(json, target)
}

func transformOpenAIChatRequestRust(json, target string) (string, error) {
	cjson := C.CString(json)
	defer C.free(unsafe.Pointer(cjson))
	ctarget := C.CString(target)
	defer C.free(unsafe.Pointer(ctarget))
	var out *C.char
	if rc := C.octopus_transform_openai_chat_request(cjson, ctarget, &out); rc != 0 {
		return "", errors.New("rust transform_openai_chat_request failed")
	}
	return takeCString(out), nil
}

func transformOpenAIResponseRust(json, target string) (string, error) {
	cjson := C.CString(json)
	defer C.free(unsafe.Pointer(cjson))
	ctarget := C.CString(target)
	defer C.free(unsafe.Pointer(ctarget))
	var out *C.char
	if rc := C.octopus_transform_openai_response(cjson, ctarget, &out); rc != 0 {
		return "", errors.New("rust transform_openai_response failed")
	}
	return takeCString(out), nil
}

func transformEmbeddingRequestRust(json, target string) (string, error) {
	cjson := C.CString(json)
	defer C.free(unsafe.Pointer(cjson))
	ctarget := C.CString(target)
	defer C.free(unsafe.Pointer(ctarget))
	var out *C.char
	if rc := C.octopus_transform_embedding_request(cjson, ctarget, &out); rc != 0 {
		return "", errors.New("rust transform_embedding_request failed")
	}
	return takeCString(out), nil
}

func transformOpenAIChatRequestGo(body, target string) (string, error) {
	var req map[string]any
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		return "", err
	}
	// Minimal Go fallback: developer -> system, include usage for streams.
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

func transformOpenAIResponseGo(body, target string) (string, error) {
	_ = target
	return body, nil
}

func transformEmbeddingRequestGo(body, target string) (string, error) {
	_ = target
	return body, nil
}
