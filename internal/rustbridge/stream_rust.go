//go:build rust
// +build rust

package rustbridge

/*
#cgo windows LDFLAGS: -L${SRCDIR}/../../rust/core/target/release -loctopus_core -lws2_32 -luserenv -ladvapi32 -lntdll -lbcrypt
#cgo linux LDFLAGS: -L${SRCDIR}/../../rust/core/target/release -loctopus_core -lm
#cgo darwin LDFLAGS: -L${SRCDIR}/../../rust/core/target/release -loctopus_core -framework Security -framework CoreFoundation
#include <stdlib.h>
typedef struct {
	int* starts;
	int* ends;
	int count;
	int consumed_bytes;
} OctopusEventBoundaries;
extern int octopus_stream_find_event_boundaries(const char* data, int len, OctopusEventBoundaries** out_boundaries);
extern void octopus_stream_boundaries_free(OctopusEventBoundaries* boundaries);
extern int octopus_stream_find_event_boundaries_ex(const char* data, int len, int* starts, int* ends, int max_events, int* out_count, int* out_consumed);
extern void* octopus_stream_buffer_create(void);
extern int octopus_stream_buffer_feed(void* handle, const char* chunk, char** out);
extern int octopus_stream_buffer_take(void* handle, char** out);
extern void octopus_stream_buffer_free(void* handle);
extern int octopus_stream_extract_events(const char* data, char** out_events, int* out_consumed_bytes);
extern void octopus_free_string(char* s);
*/
import "C"
import (
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"unsafe"
)

const eventBufInitialCap = 512

// StreamBuffer wraps an SSE stream buffer backed by either Rust FFI or pure Go.
type StreamBuffer struct {
	mu      sync.Mutex
	useRust bool
	data    string
	starts  []C.int
	ends    []C.int
}

// NewStreamBuffer creates a new stream buffer. Set OCTOPUS_RUST_STREAM=0 to use
// the pure-Go implementation even when the Rust build tag is enabled.
func NewStreamBuffer() *StreamBuffer {
	return &StreamBuffer{useRust: envEnabled(envDisableRustStream)}
}

// Feed appends a chunk and returns all complete events extracted so far.
func (b *StreamBuffer) Feed(chunk string) ([]string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.useRust {
		var events []string
		b.data, events = FeedStreamBuffer(b.data, chunk)
		return events, nil
	}
	// Fast path: if the buffer is empty, process the incoming chunk directly
	// without paying for a string copy.
	if b.data == "" {
		if chunk == "" {
			return nil, nil
		}
		return b.feedBytes(chunk)
	}

	b.data += chunk
	return b.feedBytes(b.data)
}

func (b *StreamBuffer) feedBytes(src string) ([]string, error) {
	n := len(src)
	cdata := (*C.char)(unsafe.Pointer(unsafe.StringData(src)))

	needed := eventBufInitialCap
	if cap(b.starts) < needed {
		b.starts = make([]C.int, needed)
		b.ends = make([]C.int, needed)
	}
	starts := b.starts[:needed]
	ends := b.ends[:needed]

	var count C.int
	var consumed C.int
	rc := C.octopus_stream_find_event_boundaries_ex(cdata, C.int(n), &starts[0], &ends[0], C.int(needed), &count, &consumed)
	if rc == -15 {
		needed = n
		b.starts = make([]C.int, needed)
		b.ends = make([]C.int, needed)
		starts = b.starts
		ends = b.ends
		rc = C.octopus_stream_find_event_boundaries_ex(cdata, C.int(n), &starts[0], &ends[0], C.int(needed), &count, &consumed)
	}
	if rc != 0 {
		return nil, errors.New("rust stream_find_event_boundaries_ex failed")
	}

	events := make([]string, 0, int(count))
	for i := 0; i < int(count); i++ {
		s := int(starts[i])
		e := int(ends[i])
		if e > s {
			events = append(events, src[s:e])
		}
	}

	if int(consumed) >= n {
		b.data = ""
	} else {
		b.data = src[int(consumed):]
	}
	return events, nil
}

// Take returns all complete events currently buffered without feeding new data.
func (b *StreamBuffer) Take() ([]string, error) {
	return b.Feed("")
}

// Close is a no-op for the stateless Rust extractor.
func (b *StreamBuffer) Close() {}

func parseStringArray(s string) ([]string, error) {
	var arr []string
	if s == "" {
		return arr, nil
	}
	if err := json.Unmarshal([]byte(s), &arr); err != nil {
		return nil, err
	}
	return arr, nil
}

// FeedStreamBuffer is a pure-function helper that feeds a chunk into the
// provided partial buffer and returns the updated partial buffer plus complete
// events. It works as a Go fallback.
func FeedStreamBuffer(partial, chunk string) (string, []string) {
	data := partial + chunk
	var events []string
	for {
		pos := strings.Index(data, "\n\n")
		if pos < 0 {
			break
		}
		event := strings.TrimSpace(data[:pos])
		if event != "" {
			events = append(events, event)
		}
		if pos+2 >= len(data) {
			data = ""
			break
		}
		data = data[pos+2:]
	}
	return data, events
}
