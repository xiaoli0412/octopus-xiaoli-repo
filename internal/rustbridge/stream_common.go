package rustbridge

import (
	"encoding/json"
	"strings"
)

// FeedStreamBuffer is a pure-function helper that feeds a chunk into the
// provided partial buffer and returns the updated partial buffer plus complete
// events. It is shared between the pure-Go and Rust builds so that benchmarks
// can compare both paths without duplicating the logic.
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
