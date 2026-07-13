package relay

import (
	"math"
	"math/rand"
	"time"
)

// ErrorClass 表示上游错误的分类，用于驱动差异化的重试与 failover 策略。
type ErrorClass int

const (
	ErrorClassUnknown ErrorClass = iota
	ErrorClassRateLimited
	ErrorClassAuthFailed
	ErrorClassServerError
	ErrorClassTimeout
	ErrorClassClientError
	ErrorClassNetworkError
)

// maxRateLimitRetries 是 429 限流场景下对同一通道密钥的最大重试次数。
const maxRateLimitRetries = 2

// ClassifyHTTPError 根据 HTTP 状态码和错误对象分类上游错误。
//   - 429 → RateLimited
//   - 401, 403 → AuthFailed
//   - 500, 502, 503 → ServerError
//   - 408, 504 → Timeout
//   - 其他 4xx → ClientError
//   - 网络错误（err != nil 且无 statusCode）→ NetworkError
func ClassifyHTTPError(statusCode int, err error) ErrorClass {
	if statusCode == 0 {
		if err != nil {
			return ErrorClassNetworkError
		}
		return ErrorClassUnknown
	}
	switch statusCode {
	case 429:
		return ErrorClassRateLimited
	case 401, 403:
		return ErrorClassAuthFailed
	case 500, 502, 503:
		return ErrorClassServerError
	case 408, 504:
		return ErrorClassTimeout
	default:
		if statusCode >= 500 && statusCode < 600 {
			return ErrorClassServerError
		}
		if statusCode >= 400 && statusCode < 500 {
			return ErrorClassClientError
		}
		if err != nil {
			return ErrorClassNetworkError
		}
		return ErrorClassUnknown
	}
}

// ShouldRetry 返回是否应在同一通道密钥上重试。
// 仅 RateLimited 返回 true。
func ShouldRetry(class ErrorClass) bool {
	return class == ErrorClassRateLimited
}

// ShouldFailover 返回是否应切换到下一个通道。
// AuthFailed/ServerError/Timeout/NetworkError 返回 true，
// RateLimited/ClientError 返回 false。
func ShouldFailover(class ErrorClass) bool {
	switch class {
	case ErrorClassAuthFailed, ErrorClassServerError, ErrorClassTimeout, ErrorClassNetworkError:
		return true
	default:
		return false
	}
}

// ComputeBackoff 计算指数退避延迟：baseDelayMs × 1.5^attempt + rand(0, 500ms)。
func ComputeBackoff(attempt int, baseDelayMs int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	if baseDelayMs < 0 {
		baseDelayMs = 0
	}
	multiplier := math.Pow(1.5, float64(attempt))
	base := float64(baseDelayMs) * multiplier * float64(time.Millisecond)
	jitter := rand.Float64() * float64(500*time.Millisecond)
	return time.Duration(base + jitter)
}
