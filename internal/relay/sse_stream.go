package relay

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/rustbridge"
)

// useGoSSE reports whether the legacy github.com/tmaxmax/go-sse path should be
// used instead of rustbridge.StreamBuffer.
func useGoSSE() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("OCTOPUS_USE_GO_SSE")))
	return v == "1" || v == "true"
}

// extractSSEData extracts the joined "data:" payload from a single SSE event.
// Other fields (event:/id:/retry:) are ignored. Multiple data: lines are joined
// with "\n" and empty data lines are skipped. The [DONE] marker is returned as
// the literal string "[DONE]". If the event contains no data lines, an empty
// string is returned.
func extractSSEData(event string) string {
	var data []string
	for _, line := range strings.Split(event, "\n") {
		line = strings.TrimSuffix(line, "\r")
		colon := strings.Index(line, ":")
		if colon < 0 {
			continue
		}
		field := line[:colon]
		// SSE field names are case-insensitive; ignore malformed lines with
		// whitespace around the field name.
		if strings.TrimSpace(field) != field {
			continue
		}
		if !strings.EqualFold(field, "data") {
			continue
		}
		value := line[colon+1:]
		if len(value) > 0 && value[0] == ' ' {
			value = value[1:]
		}
		if value == "" {
			continue
		}
		data = append(data, value)
	}
	if len(data) == 0 {
		return ""
	}
	return strings.Join(data, "\n")
}

// sseReadResult carries one extracted SSE data payload or a read error from the
// streaming goroutine to handleStreamResponse.
type sseReadResult struct {
	data string
	err  error
}

// readStreamWithBuffer reads response.Body using rustbridge.StreamBuffer and
// emits extracted SSE data payloads on results. It closes results when the body
// is fully read or an error occurs.
func readStreamWithBuffer(body io.Reader, results chan<- sseReadResult, stopReading <-chan struct{}) {
	defer close(results)

	buf := rustbridge.NewStreamBuffer()
	defer buf.Close()

	reader := bufio.NewReader(body)
	readBuf := make([]byte, 32*1024)
	for {
		n, err := reader.Read(readBuf)
		if n > 0 {
			events, feedErr := buf.Feed(string(readBuf[:n]))
			if feedErr != nil {
				select {
				case results <- sseReadResult{err: feedErr}:
				case <-stopReading:
				}
				return
			}
			for _, ev := range events {
				data := extractSSEData(ev)
				if data == "" {
					continue
				}
				select {
				case results <- sseReadResult{data: data}:
				case <-stopReading:
					return
				}
			}
		}
		if err != nil {
			if err != io.EOF {
				select {
				case results <- sseReadResult{err: err}:
				case <-stopReading:
				}
			}
			return
		}
	}
}
