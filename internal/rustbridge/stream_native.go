//go:build !rust
// +build !rust

package rustbridge

import (
	"sync"
)

// StreamBuffer is a pure-Go SSE stream buffer.
type StreamBuffer struct {
	mu   sync.Mutex
	data string
}

// NewStreamBuffer creates a new stream buffer.
func NewStreamBuffer() *StreamBuffer {
	return &StreamBuffer{}
}

// Feed appends a chunk and returns all complete events extracted so far.
func (b *StreamBuffer) Feed(chunk string) ([]string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	var events []string
	b.data, events = FeedStreamBuffer(b.data, chunk)
	return events, nil
}

// Take returns all complete events currently buffered without feeding new data.
func (b *StreamBuffer) Take() ([]string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	_, events := FeedStreamBuffer("", b.data)
	b.data = ""
	return events, nil
}

// Close is a no-op for the pure-Go buffer.
func (b *StreamBuffer) Close() {}
