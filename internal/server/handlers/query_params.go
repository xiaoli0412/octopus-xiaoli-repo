package handlers

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func parseOptionalBoolQuery(c *gin.Context, key string, defaultValue bool) (bool, error) {
	if c == nil {
		return defaultValue, nil
	}
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return defaultValue, nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("invalid %s", key)
	}
	return value, nil
}

func parseOptionalIntQuery(c *gin.Context, key string, defaultValue int) (int, error) {
	if c == nil {
		return defaultValue, nil
	}
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return defaultValue, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid %s", key)
	}
	return value, nil
}

func parseOptionalIntRangeQuery(c *gin.Context, startKey, endKey string) (*int, *int, error) {
	if c == nil {
		return nil, nil, nil
	}

	startRaw := strings.TrimSpace(c.Query(startKey))
	endRaw := strings.TrimSpace(c.Query(endKey))
	if (startRaw == "") != (endRaw == "") {
		return nil, nil, fmt.Errorf("%s and %s must be provided together", startKey, endKey)
	}
	if startRaw == "" {
		return nil, nil, nil
	}

	startValue, err := strconv.Atoi(startRaw)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid %s", startKey)
	}
	endValue, err := strconv.Atoi(endRaw)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid %s", endKey)
	}
	return &startValue, &endValue, nil
}

func parseOptionalPositiveIntQuery(c *gin.Context, key string) (int, bool, error) {
	if c == nil {
		return 0, false, nil
	}
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return 0, false, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, false, fmt.Errorf("invalid %s", key)
	}
	return value, true, nil
}

func parsePositivePathIDValue(c *gin.Context, name string) (int, bool) {
	id, err := strconv.Atoi(strings.TrimSpace(c.Param(name)))
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func validateAbsoluteHTTPURL(raw, fieldName string) error {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return fmt.Errorf("%s is invalid: %w", fieldName, err)
	}
	scheme := strings.ToLower(strings.TrimSpace(parsed.Scheme))
	if !parsed.IsAbs() || (scheme != "http" && scheme != "https") || strings.TrimSpace(parsed.Host) == "" {
		return fmt.Errorf("%s must be absolute http or https URL", fieldName)
	}
	return nil
}
