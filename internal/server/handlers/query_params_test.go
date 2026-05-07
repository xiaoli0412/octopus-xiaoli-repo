package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

func TestParseOptionalBoundedIntQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		query        string
		key          string
		defaultValue int
		minValue     int
		maxValue     int
		want         int
		wantOK       bool
		wantErr      bool
	}{
		{name: "default", key: "page", defaultValue: 3, minValue: 1, maxValue: 0, want: 3, wantOK: false},
		{name: "valid", query: "page=5", key: "page", defaultValue: 1, minValue: 1, maxValue: 0, want: 5, wantOK: true},
		{name: "blank explicit", query: "page=", key: "page", defaultValue: 1, minValue: 1, maxValue: 0, wantErr: true},
		{name: "whitespace explicit", query: "page=%20", key: "page", defaultValue: 1, minValue: 1, maxValue: 0, wantErr: true},
		{name: "below min", query: "page=0", key: "page", defaultValue: 1, minValue: 1, maxValue: 0, wantErr: true},
		{name: "above max", query: "limit=10001", key: "limit", defaultValue: 20, minValue: 1, maxValue: 10000, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var c *gin.Context
			if tc.query != "" {
				recorder := httptest.NewRecorder()
				c, _ = gin.CreateTestContext(recorder)
				c.Request = httptest.NewRequest("GET", "/", nil)
				c.Request.URL.RawQuery = tc.query
			}
			got, ok, err := parseOptionalBoundedIntQuery(c, tc.key, tc.defaultValue, tc.minValue, tc.maxValue)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want || ok != tc.wantOK {
				t.Fatalf("got %d/%v, want %d/%v", got, ok, tc.want, tc.wantOK)
			}
		})
	}
}

func TestParseOptionalBoolQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		query        string
		key          string
		defaultValue bool
		want         bool
		wantErr      bool
	}{
		{name: "default", key: "include_logs", defaultValue: true, want: true},
		{name: "valid true", query: "include_logs=true", key: "include_logs", defaultValue: false, want: true},
		{name: "valid false", query: "include_logs=false", key: "include_logs", defaultValue: true, want: false},
		{name: "blank explicit", query: "include_logs=", key: "include_logs", defaultValue: true, wantErr: true},
		{name: "whitespace explicit", query: "include_logs=%20", key: "include_logs", defaultValue: true, wantErr: true},
		{name: "invalid", query: "include_logs=maybe", key: "include_logs", defaultValue: true, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var c *gin.Context
			if tc.query != "" {
				recorder := httptest.NewRecorder()
				c, _ = gin.CreateTestContext(recorder)
				c.Request = httptest.NewRequest("GET", "/", nil)
				c.Request.URL.RawQuery = tc.query
			}
			got, err := parseOptionalBoolQuery(c, tc.key, tc.defaultValue)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestParseOptionalTrimmedStringQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		query  string
		key    string
		want   string
		wantOK bool
	}{
		{name: "default", key: "state"},
		{name: "valid", query: "state=abc123", key: "state", want: "abc123", wantOK: true},
		{name: "trimmed valid", query: "state=%20abc123%20", key: "state", want: "abc123", wantOK: true},
		{name: "blank explicit", query: "state=", key: "state"},
		{name: "whitespace explicit", query: "state=%20", key: "state"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var c *gin.Context
			if tc.query != "" {
				recorder := httptest.NewRecorder()
				c, _ = gin.CreateTestContext(recorder)
				c.Request = httptest.NewRequest("GET", "/", nil)
				c.Request.URL.RawQuery = tc.query
			}
			got, ok := parseOptionalTrimmedStringQuery(c, tc.key)
			if got != tc.want || ok != tc.wantOK {
				t.Fatalf("got %q/%v, want %q/%v", got, ok, tc.want, tc.wantOK)
			}
		})
	}
}

func TestParseRequiredTrimmedStringQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name    string
		query   string
		key     string
		want    string
		wantErr bool
	}{
		{name: "valid", query: "code=auth-code", key: "code", want: "auth-code"},
		{name: "trimmed valid", query: "code=%20auth-code%20", key: "code", want: "auth-code"},
		{name: "missing", key: "code", wantErr: true},
		{name: "blank explicit", query: "code=", key: "code", wantErr: true},
		{name: "whitespace explicit", query: "code=%20", key: "code", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var c *gin.Context
			if tc.query != "" {
				recorder := httptest.NewRecorder()
				c, _ = gin.CreateTestContext(recorder)
				c.Request = httptest.NewRequest("GET", "/", nil)
				c.Request.URL.RawQuery = tc.query
			}
			got, err := parseRequiredTrimmedStringQuery(c, tc.key)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestParseOptionalNonEmptyTrimmedStringQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name    string
		query   string
		key     string
		want    string
		wantOK  bool
		wantErr bool
	}{
		{name: "default", key: "status"},
		{name: "valid", query: "status=succeeded", key: "status", want: "succeeded", wantOK: true},
		{name: "trimmed valid", query: "status=%20succeeded%20", key: "status", want: "succeeded", wantOK: true},
		{name: "blank explicit", query: "status=", key: "status", wantErr: true},
		{name: "whitespace explicit", query: "status=%20", key: "status", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var c *gin.Context
			if tc.query != "" {
				recorder := httptest.NewRecorder()
				c, _ = gin.CreateTestContext(recorder)
				c.Request = httptest.NewRequest("GET", "/", nil)
				c.Request.URL.RawQuery = tc.query
			}
			got, ok, err := parseOptionalNonEmptyTrimmedStringQuery(c, tc.key)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want || ok != tc.wantOK {
				t.Fatalf("got %q/%v, want %q/%v", got, ok, tc.want, tc.wantOK)
			}
		})
	}
}

func TestParseOptionalNonEmptyTrimmedPostForm(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name    string
		form    string
		key     string
		want    string
		wantOK  bool
		wantErr bool
	}{
		{name: "default", key: "preview_token"},
		{name: "valid", form: "preview_token=token-value", key: "preview_token", want: "token-value", wantOK: true},
		{name: "trimmed valid", form: "preview_token=%20token-value%20", key: "preview_token", want: "token-value", wantOK: true},
		{name: "blank explicit", form: "preview_token=", key: "preview_token", wantErr: true},
		{name: "whitespace explicit", form: "preview_token=%20", key: "preview_token", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var c *gin.Context
			if tc.form != "" {
				recorder := httptest.NewRecorder()
				c, _ = gin.CreateTestContext(recorder)
				c.Request = httptest.NewRequest("POST", "/", strings.NewReader(tc.form))
				c.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}
			got, ok, err := parseOptionalNonEmptyTrimmedPostForm(c, tc.key)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want || ok != tc.wantOK {
				t.Fatalf("got %q/%v, want %q/%v", got, ok, tc.want, tc.wantOK)
			}
		})
	}
}

func TestParseOptionalNonEmptyTrimmedHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name    string
		header  string
		key     string
		want    string
		wantOK  bool
		wantErr bool
	}{
		{name: "default", key: "X-Octopus-Import-Preview-Token"},
		{name: "valid", header: "token-value", key: "X-Octopus-Import-Preview-Token", want: "token-value", wantOK: true},
		{name: "trimmed valid", header: "  token-value  ", key: "X-Octopus-Import-Preview-Token", want: "token-value", wantOK: true},
		{name: "blank explicit", header: "", key: "X-Octopus-Import-Preview-Token", wantErr: true},
		{name: "whitespace explicit", header: "   ", key: "X-Octopus-Import-Preview-Token", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest("POST", "/", nil)
			if tc.name != "default" {
				c.Request.Header.Add(tc.key, tc.header)
			}
			got, ok, err := parseOptionalNonEmptyTrimmedHeader(c, tc.key)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want || ok != tc.wantOK {
				t.Fatalf("got %q/%v, want %q/%v", got, ok, tc.want, tc.wantOK)
			}
		})
	}
}

func TestParseOptionalDBImportModeQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		query        string
		defaultValue model.DBImportMode
		want         model.DBImportMode
		wantOK       bool
		wantErr      bool
	}{
		{name: "default", defaultValue: model.DBImportModeIncremental, want: model.DBImportModeIncremental, wantOK: false},
		{name: "valid", query: "mode=replace", defaultValue: model.DBImportModeIncremental, want: model.DBImportModeReplace, wantOK: true},
		{name: "trimmed valid", query: "mode=  merge  ", defaultValue: model.DBImportModeIncremental, want: model.DBImportModeMerge, wantOK: true},
		{name: "blank explicit", query: "mode=", defaultValue: model.DBImportModeIncremental, wantErr: true},
		{name: "whitespace explicit", query: "mode=%20", defaultValue: model.DBImportModeIncremental, wantErr: true},
		{name: "invalid", query: "mode=surprise", defaultValue: model.DBImportModeIncremental, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var c *gin.Context
			if tc.query != "" {
				recorder := httptest.NewRecorder()
				c, _ = gin.CreateTestContext(recorder)
				c.Request = httptest.NewRequest("GET", "/", nil)
				c.Request.URL.RawQuery = tc.query
			}
			got, ok, err := parseOptionalDBImportModeQuery(c, "mode", tc.defaultValue)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want || ok != tc.wantOK {
				t.Fatalf("got %q/%v, want %q/%v", got, ok, tc.want, tc.wantOK)
			}
		})
	}
}

func TestParseOptionalDBExportFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		query        string
		defaultValue dbExportFormat
		want         dbExportFormat
		wantErr      bool
	}{
		{name: "default", defaultValue: dbExportFormatStandard, want: dbExportFormatStandard},
		{name: "valid standard", query: "format=standard", defaultValue: dbExportFormatLegacy, want: dbExportFormatStandard},
		{name: "valid legacy", query: "format=legacy", defaultValue: dbExportFormatStandard, want: dbExportFormatLegacy},
		{name: "trimmed valid", query: "format=%20legacy%20", defaultValue: dbExportFormatStandard, want: dbExportFormatLegacy},
		{name: "blank explicit", query: "format=", defaultValue: dbExportFormatStandard, wantErr: true},
		{name: "whitespace explicit", query: "format=%20", defaultValue: dbExportFormatStandard, wantErr: true},
		{name: "invalid", query: "format=surprise", defaultValue: dbExportFormatStandard, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var c *gin.Context
			if tc.query != "" {
				recorder := httptest.NewRecorder()
				c, _ = gin.CreateTestContext(recorder)
				c.Request = httptest.NewRequest("GET", "/", nil)
				c.Request.URL.RawQuery = tc.query
			}
			got, err := parseOptionalDBExportFormat(c, "format", tc.defaultValue)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestParseOptionalRelayLogExportFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		query        string
		defaultValue relayLogExportFormat
		want         relayLogExportFormat
		wantErr      bool
	}{
		{name: "default", defaultValue: relayLogExportFormatJSONL, want: relayLogExportFormatJSONL},
		{name: "valid json", query: "format=json", defaultValue: relayLogExportFormatJSONL, want: relayLogExportFormatJSON},
		{name: "valid jsonl", query: "format=jsonl", defaultValue: relayLogExportFormatJSON, want: relayLogExportFormatJSONL},
		{name: "trimmed valid", query: "format=%20json%20", defaultValue: relayLogExportFormatJSONL, want: relayLogExportFormatJSON},
		{name: "blank explicit", query: "format=", defaultValue: relayLogExportFormatJSONL, wantErr: true},
		{name: "whitespace explicit", query: "format=%20", defaultValue: relayLogExportFormatJSONL, wantErr: true},
		{name: "invalid", query: "format=surprise", defaultValue: relayLogExportFormatJSONL, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var c *gin.Context
			if tc.query != "" {
				recorder := httptest.NewRecorder()
				c, _ = gin.CreateTestContext(recorder)
				c.Request = httptest.NewRequest("GET", "/", nil)
				c.Request.URL.RawQuery = tc.query
			}
			got, err := parseOptionalRelayLogExportFormat(c, "format", tc.defaultValue)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestParseOptionalRFC3339TimeQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	valid := "2026-05-02T20:10:30Z"
	validParsed, err := time.Parse(time.RFC3339, valid)
	if err != nil {
		t.Fatalf("time.Parse(valid) error = %v", err)
	}

	tests := []struct {
		name    string
		query   string
		want    *time.Time
		wantErr bool
	}{
		{name: "default"},
		{name: "valid", query: "created_from=" + valid, want: &validParsed},
		{name: "trimmed valid", query: "created_from=%20" + valid + "%20", want: &validParsed},
		{name: "blank explicit", query: "created_from=", wantErr: true},
		{name: "whitespace explicit", query: "created_from=%20", wantErr: true},
		{name: "invalid", query: "created_from=not-a-time", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var c *gin.Context
			if tc.query != "" {
				recorder := httptest.NewRecorder()
				c, _ = gin.CreateTestContext(recorder)
				c.Request = httptest.NewRequest("GET", "/", nil)
				c.Request.URL.RawQuery = tc.query
			}
			got, err := parseOptionalRFC3339TimeQuery(c, "created_from")
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.want == nil {
				if got != nil {
					t.Fatalf("got %v, want nil", got)
				}
				return
			}
			if got == nil || !got.Equal(*tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestParseOptionalIntRangeQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name      string
		query     string
		wantStart *int
		wantEnd   *int
		wantErr   bool
	}{
		{name: "default"},
		{name: "valid", query: "start_time=1700000000&end_time=1700000100", wantStart: intPtr(1700000000), wantEnd: intPtr(1700000100)},
		{name: "trimmed valid", query: "start_time=%201700000000%20&end_time=%201700000100%20", wantStart: intPtr(1700000000), wantEnd: intPtr(1700000100)},
		{name: "start only", query: "start_time=1700000000", wantErr: true},
		{name: "end only", query: "end_time=1700000100", wantErr: true},
		{name: "blank explicit", query: "start_time=&end_time=", wantErr: true},
		{name: "whitespace explicit", query: "start_time=%20&end_time=%20", wantErr: true},
		{name: "invalid start", query: "start_time=nan&end_time=1700000100", wantErr: true},
		{name: "invalid end", query: "start_time=1700000000&end_time=nan", wantErr: true},
		{name: "reversed range", query: "start_time=1700000100&end_time=1700000000", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var c *gin.Context
			if tc.query != "" {
				recorder := httptest.NewRecorder()
				c, _ = gin.CreateTestContext(recorder)
				c.Request = httptest.NewRequest("GET", "/", nil)
				c.Request.URL.RawQuery = tc.query
			}
			gotStart, gotEnd, err := parseOptionalIntRangeQuery(c, "start_time", "end_time")
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !equalIntPtr(gotStart, tc.wantStart) || !equalIntPtr(gotEnd, tc.wantEnd) {
				t.Fatalf("got %v/%v, want %v/%v", gotStart, gotEnd, tc.wantStart, tc.wantEnd)
			}
		})
	}
}

func intPtr(v int) *int {
	return &v
}

func equalIntPtr(a, b *int) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}
