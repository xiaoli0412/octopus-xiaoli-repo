package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/xurl"
)

func parseOptionalBoolQuery(c *gin.Context, key string, defaultValue bool) (bool, error) {
	if c == nil {
		return defaultValue, nil
	}
	raw, exists := c.GetQuery(key)
	if !exists {
		return defaultValue, nil
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false, fmt.Errorf("invalid %s", key)
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("invalid %s", key)
	}
	return value, nil
}

func parseOptionalTrimmedStringQuery(c *gin.Context, key string) (string, bool) {
	if c == nil {
		return "", false
	}
	raw, exists := c.GetQuery(key)
	if !exists {
		return "", false
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false
	}
	return raw, true
}

func parseRequiredTrimmedStringQuery(c *gin.Context, key string) (string, error) {
	value, ok := parseOptionalTrimmedStringQuery(c, key)
	if !ok {
		return "", fmt.Errorf("missing %s", key)
	}
	return value, nil
}

func parseOptionalNonEmptyTrimmedStringQuery(c *gin.Context, key string) (string, bool, error) {
	if c == nil {
		return "", false, nil
	}
	raw, exists := c.GetQuery(key)
	if !exists {
		return "", false, nil
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false, fmt.Errorf("invalid %s", key)
	}
	return raw, true, nil
}

func parseOptionalNonEmptyTrimmedPostForm(c *gin.Context, key string) (string, bool, error) {
	if c == nil {
		return "", false, nil
	}
	raw, exists := c.GetPostForm(key)
	if !exists {
		return "", false, nil
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false, fmt.Errorf("invalid %s", key)
	}
	return raw, true, nil
}

func parseOptionalNonEmptyTrimmedHeader(c *gin.Context, key string) (string, bool, error) {
	if c == nil || c.Request == nil {
		return "", false, nil
	}
	values := c.Request.Header.Values(key)
	if len(values) == 0 {
		return "", false, nil
	}
	raw := strings.TrimSpace(values[0])
	if raw == "" {
		return "", false, fmt.Errorf("invalid %s", key)
	}
	return raw, true, nil
}

func parseOptionalIntRangeQuery(c *gin.Context, startKey, endKey string) (*int, *int, error) {
	if c == nil {
		return nil, nil, nil
	}

	startRaw, startExists := c.GetQuery(startKey)
	endRaw, endExists := c.GetQuery(endKey)
	if startExists != endExists {
		return nil, nil, fmt.Errorf("%s and %s must be provided together", startKey, endKey)
	}
	if !startExists {
		return nil, nil, nil
	}

	startRaw = strings.TrimSpace(startRaw)
	endRaw = strings.TrimSpace(endRaw)
	if startRaw == "" || endRaw == "" {
		return nil, nil, fmt.Errorf("invalid %s", startKey)
	}

	startValue, err := strconv.Atoi(startRaw)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid %s", startKey)
	}
	endValue, err := strconv.Atoi(endRaw)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid %s", endKey)
	}
	if startValue > endValue {
		return nil, nil, fmt.Errorf("%s must be less than or equal to %s", startKey, endKey)
	}
	return &startValue, &endValue, nil
}

func parseOptionalBoundedIntQuery(c *gin.Context, key string, defaultValue, minValue, maxValue int) (int, bool, error) {
	if c == nil {
		return defaultValue, false, nil
	}
	raw, exists := c.GetQuery(key)
	if !exists {
		return defaultValue, false, nil
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, false, fmt.Errorf("invalid %s", key)
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < minValue || (maxValue > 0 && value > maxValue) {
		return 0, false, fmt.Errorf("invalid %s", key)
	}
	return value, true, nil
}

func parseOptionalDBImportModeQuery(c *gin.Context, key string, defaultValue model.DBImportMode) (model.DBImportMode, bool, error) {
	if c == nil {
		return defaultValue, false, nil
	}
	raw, exists := c.GetQuery(key)
	if !exists {
		return defaultValue, false, nil
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false, fmt.Errorf("unsupported import mode")
	}
	mode := model.NormalizeDBImportMode(raw)
	if !model.IsValidDBImportMode(mode) {
		return "", false, fmt.Errorf("unsupported import mode")
	}
	return mode, true, nil
}

func normalizeDBExportFormat(input string) dbExportFormat {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case string(dbExportFormatStandard):
		return dbExportFormatStandard
	case string(dbExportFormatLegacy):
		return dbExportFormatLegacy
	default:
		return ""
	}
}

func parseOptionalDBExportFormat(c *gin.Context, key string, defaultValue dbExportFormat) (dbExportFormat, error) {
	if c == nil {
		return defaultValue, nil
	}
	raw, exists := c.GetQuery(key)
	if !exists {
		return defaultValue, nil
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("unsupported export format")
	}
	format := normalizeDBExportFormat(raw)
	if format == "" {
		return "", fmt.Errorf("unsupported export format")
	}
	return format, nil
}

func normalizeRelayLogExportFormat(input string) relayLogExportFormat {
	switch strings.TrimSpace(input) {
	case string(relayLogExportFormatJSON):
		return relayLogExportFormatJSON
	case string(relayLogExportFormatJSONL):
		return relayLogExportFormatJSONL
	default:
		return ""
	}
}

func parseOptionalRelayLogExportFormat(c *gin.Context, key string, defaultValue relayLogExportFormat) (relayLogExportFormat, error) {
	if c == nil {
		return defaultValue, nil
	}
	raw, exists := c.GetQuery(key)
	if !exists {
		return defaultValue, nil
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("unsupported format")
	}
	format := normalizeRelayLogExportFormat(raw)
	if format == "" {
		return "", fmt.Errorf("unsupported format")
	}
	return format, nil
}

func parseOptionalRFC3339TimeQuery(c *gin.Context, key string) (*time.Time, error) {
	if c == nil {
		return nil, nil
	}
	raw, exists := c.GetQuery(key)
	if !exists {
		return nil, nil
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("invalid %s", key)
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, fmt.Errorf("invalid %s", key)
	}
	return &parsed, nil
}

func parseOptionalPositiveIntQuery(c *gin.Context, key string) (int, bool, error) {
	value, ok, err := parseOptionalBoundedIntQuery(c, key, 0, 1, 0)
	if err != nil {
		return 0, false, err
	}
	if !ok {
		return 0, false, nil
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
	return xurl.ValidateAbsoluteHTTPURL(raw, fieldName)
}

func normalizeOptionalProxyURL(raw *string, fieldName string) (*string, error) {
	if raw == nil {
		return nil, nil
	}
	normalized := strings.TrimSpace(*raw)
	if normalized == "" {
		return nil, nil
	}
	if err := xurl.ValidateProxyURL(normalized, fieldName); err != nil {
		return nil, err
	}
	return &normalized, nil
}
