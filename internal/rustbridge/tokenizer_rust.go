//go:build rust
// +build rust

package rustbridge

/*
#cgo windows LDFLAGS: -L${SRCDIR}/../../rust/core/target/release -loctopus_core -lws2_32 -luserenv -ladvapi32 -lntdll -lbcrypt
#cgo linux LDFLAGS: -L${SRCDIR}/../../rust/core/target/release -loctopus_core -lm
#cgo darwin LDFLAGS: -L${SRCDIR}/../../rust/core/target/release -loctopus_core -framework Security -framework CoreFoundation
#include <stdlib.h>
extern int octopus_tokenizer_count(const char* text, const char* model);
extern void octopus_free_string(char* s);
extern int octopus_json_extract_model(const char* json, char** out);
extern int octopus_json_extract_usage(const char* json, char** out);
extern int octopus_sse_aggregate(const char* aggregate, const char* chunk, char** out);
extern int octopus_transform_openai_chat_request(const char* json, const char* target, char** out);
extern int octopus_transform_openai_response(const char* json, const char* target, char** out);
extern int octopus_transform_embedding_request(const char* json, const char* target, char** out);
extern int octopus_balance_select(const char *candidates_json, const char *strategy, int current_idx, char **out);
typedef struct {
    long long id;
    long long weight;
    long long latency;
    long long priority;
    int healthy;
    int circuit_open;
} OctopusBalanceCandidate;
extern int octopus_balance_select_v2(const OctopusBalanceCandidate* candidates, int candidates_len, const char* strategy, int current_idx, long long* out_id, int* out_next_index);
extern void* octopus_stream_buffer_create(void);
extern int octopus_stream_buffer_feed(void* handle, const char* chunk, char** out);
extern int octopus_stream_buffer_take(void* handle, char** out);
extern void octopus_stream_buffer_free(void* handle);
extern int octopus_stream_extract_events(const char* data, char** out_events, int* out_consumed_bytes);
extern int octopus_stats_merge_hourly(const char *existing_json, const char *delta_json, char **out);
extern int octopus_stats_quantile(const char *values_json, double quantile, char **out);
*/
import "C"
import (
	"errors"
	"unsafe"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/tokenizer"
)

// Enabled reports whether the Rust FFI backend was compiled in.
func Enabled() bool { return true }

// CountTokens returns the token count for the given content and model.
// It uses the Rust tokenizer by default; set OCTOPUS_RUST_TOKENIZER=0 to fall
// back to the Go tiktoken-go implementation.
func CountTokens(content, model string) int {
	if !envEnabled(envDisableRustTokenizer) {
		return countTokensGo(content, model)
	}
	return countTokensRust(content, model)
}

func countTokensGo(content, model string) int {
	return tokenizer.CountTokensWithModel(content, model)
}

func countTokensRust(content, model string) int {
	ctext := C.CString(content)
	defer C.free(unsafe.Pointer(ctext))
	cmodel := C.CString(model)
	defer C.free(unsafe.Pointer(cmodel))
	return int(C.octopus_tokenizer_count(ctext, cmodel))
}

// ExtractModel returns a JSON object `{"model":"..."}` extracted from the
// request/response body.
func ExtractModel(json string) (string, error) {
	cjson := C.CString(json)
	defer C.free(unsafe.Pointer(cjson))
	var out *C.char
	if rc := C.octopus_json_extract_model(cjson, &out); rc != 0 {
		return "", errors.New("rust extract_model failed")
	}
	return takeCString(out), nil
}

// ExtractUsage returns a JSON usage object extracted from the response body.
func ExtractUsage(json string) (string, error) {
	cjson := C.CString(json)
	defer C.free(unsafe.Pointer(cjson))
	var out *C.char
	if rc := C.octopus_json_extract_usage(cjson, &out); rc != 0 {
		return "", errors.New("rust extract_usage failed")
	}
	return takeCString(out), nil
}

// SSEAggregate merges a streaming chunk into an aggregate response JSON.
func SSEAggregate(aggregate, chunk string) (string, error) {
	cagg := C.CString(aggregate)
	defer C.free(unsafe.Pointer(cagg))
	cchunk := C.CString(chunk)
	defer C.free(unsafe.Pointer(cchunk))
	var out *C.char
	if rc := C.octopus_sse_aggregate(cagg, cchunk, &out); rc != 0 {
		return "", errors.New("rust sse_aggregate failed")
	}
	return takeCString(out), nil
}

func takeCString(ptr *C.char) string {
	if ptr == nil {
		return ""
	}
	defer C.octopus_free_string(ptr)
	return C.GoString(ptr)
}
