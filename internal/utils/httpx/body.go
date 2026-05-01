package httpx

import (
	"fmt"
	"io"
)

const defaultTooLargeMessage = "response body too large"

func ReadLimitedBody(r io.Reader, limit int64, tooLargeMessage string) ([]byte, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("invalid read limit")
	}
	if tooLargeMessage == "" {
		tooLargeMessage = defaultTooLargeMessage
	}
	payload, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(payload)) > limit {
		return nil, fmt.Errorf("%s", tooLargeMessage)
	}
	return payload, nil
}
