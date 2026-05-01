package openai

import (
	"fmt"
	"net/http"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/httpx"
)

const maxOpenAIResponseBodyBytes int64 = 32 << 20

func readOpenAIResponseBody(response *http.Response) ([]byte, error) {
	if response == nil {
		return nil, fmt.Errorf("response is nil")
	}
	body, err := httpx.ReadLimitedBody(response.Body, maxOpenAIResponseBodyBytes, "openai response body too large")
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("response body is empty")
	}
	return body, nil
}
