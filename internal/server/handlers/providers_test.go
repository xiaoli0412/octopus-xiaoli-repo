package handlers

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestGetProvidersCachesRemotePayload(t *testing.T) {
	resetProvidersState()
	t.Cleanup(resetProvidersState)

	now := time.Date(2026, time.April, 21, 3, 0, 0, 0, time.UTC)
	providersNow = func() time.Time { return now }

	remoteCalls := 0
	providersRemoteFetch = func() ([]byte, error) {
		remoteCalls++
		return []byte(`[{"name":"remote","channel_type":0,"base_url":"https://remote.example/v1"}]`), nil
	}
	providersEmbeddedRead = func() ([]byte, error) {
		t.Fatal("embedded fallback should not be used when remote fetch succeeds")
		return nil, nil
	}

	first, source, err := getProvidersPayload()
	if err != nil {
		t.Fatalf("getProvidersPayload() first error = %v", err)
	}
	if source != "remote" {
		t.Fatalf("source = %q, want remote", source)
	}
	second, secondSource, err := getProvidersPayload()
	if err != nil {
		t.Fatalf("getProvidersPayload() second error = %v", err)
	}
	if secondSource != "remote" {
		t.Fatalf("secondSource = %q, want remote", secondSource)
	}

	if remoteCalls != 1 {
		t.Fatalf("remoteCalls = %d, want 1", remoteCalls)
	}
	if string(first) != `[{"name":"remote","channel_type":0,"base_url":"https://remote.example/v1"}]` {
		t.Fatalf("first payload = %s", string(first))
	}
	if string(second) != string(first) {
		t.Fatalf("second payload = %s, want %s", string(second), string(first))
	}
	if &first[0] == &second[0] {
		t.Fatalf("expected cached payload copies to avoid shared mutation")
	}
}

func TestGetProvidersFallsBackToEmbeddedPayload(t *testing.T) {
	resetProvidersState()
	t.Cleanup(resetProvidersState)

	providersRemoteFetch = func() ([]byte, error) {
		return nil, io.EOF
	}
	providersEmbeddedRead = func() ([]byte, error) {
		return []byte(`[{"name":"embedded","channel_type":0,"base_url":"https://embedded.example/v1"}]`), nil
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/providers", nil)

	GetProviders(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if got := recorder.Header().Get("Content-Type"); !strings.HasPrefix(got, providersResponseMimeType) {
		t.Fatalf("content-type = %q, want prefix %q", got, providersResponseMimeType)
	}
	if got := recorder.Header().Get(providersSourceHeader); got != "embedded" {
		t.Fatalf("%s = %q, want embedded", providersSourceHeader, got)
	}
	if got := recorder.Header().Get(providersCommitHeader); got != providersPinnedCommitSHA {
		t.Fatalf("%s = %q, want %q", providersCommitHeader, got, providersPinnedCommitSHA)
	}
	if recorder.Body.String() != `[{"name":"embedded","channel_type":0,"base_url":"https://embedded.example/v1"}]` {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestGetProvidersFallsBackWhenRemotePayloadInvalid(t *testing.T) {
	resetProvidersState()
	t.Cleanup(resetProvidersState)

	remoteCalls := 0
	providersRemoteFetch = func() ([]byte, error) {
		remoteCalls++
		return []byte(`<html>bad gateway</html>`), nil
	}
	providersEmbeddedRead = func() ([]byte, error) {
		return []byte(`[{"name":"embedded","channel_type":0,"base_url":"https://example.com/v1"}]`), nil
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/providers", nil)

	GetProviders(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if remoteCalls != 1 {
		t.Fatalf("remoteCalls = %d, want 1", remoteCalls)
	}
	if got := recorder.Header().Get(providersSourceHeader); got != "embedded" {
		t.Fatalf("%s = %q, want embedded", providersSourceHeader, got)
	}
	if recorder.Body.String() != `[{"name":"embedded","channel_type":0,"base_url":"https://example.com/v1"}]` {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestGetProvidersMarksRemoteSourceInResponseHeaders(t *testing.T) {
	resetProvidersState()
	t.Cleanup(resetProvidersState)

	providersRemoteFetch = func() ([]byte, error) {
		return []byte(`[{"name":"remote","channel_type":0,"base_url":"https://remote.example/v1"}]`), nil
	}
	providersEmbeddedRead = func() ([]byte, error) {
		t.Fatal("embedded fallback should not be used when remote fetch succeeds")
		return nil, nil
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/providers", nil)

	GetProviders(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if got := recorder.Header().Get(providersSourceHeader); got != "remote" {
		t.Fatalf("%s = %q, want remote", providersSourceHeader, got)
	}
	if got := recorder.Header().Get(providersCommitHeader); got != providersPinnedCommitSHA {
		t.Fatalf("%s = %q, want %q", providersCommitHeader, got, providersPinnedCommitSHA)
	}
}

func TestGetProvidersReturnsErrorWhenEmbeddedPayloadInvalid(t *testing.T) {
	resetProvidersState()
	t.Cleanup(resetProvidersState)

	providersRemoteFetch = func() ([]byte, error) {
		return nil, io.EOF
	}
	providersEmbeddedRead = func() ([]byte, error) {
		return []byte(`[{"name":"broken"}]`), nil
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/providers", nil)

	GetProviders(ctx)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusInternalServerError, recorder.Body.String())
	}
}

func TestFetchProvidersFromGitHubRejectsOversizedPayload(t *testing.T) {
	resetProvidersState()
	t.Cleanup(resetProvidersState)

	providersHTTPClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(bytes.NewReader(bytes.Repeat([]byte("a"), maxProvidersPayloadBytes+1))),
			}, nil
		}),
	}

	_, err := fetchProvidersFromGitHub()
	if err == nil {
		t.Fatal("expected oversized payload error")
	}
	if !strings.Contains(err.Error(), "payload too large") {
		t.Fatalf("error = %v, want payload too large", err)
	}
}

func TestFetchProvidersFromGitHubRejectsInvalidPayload(t *testing.T) {
	resetProvidersState()
	t.Cleanup(resetProvidersState)

	providersHTTPClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"name":"oops"}`)),
			}, nil
		}),
	}

	_, err := fetchProvidersFromGitHub()
	if err == nil {
		t.Fatal("expected invalid payload error")
	}
	if !strings.Contains(err.Error(), "invalid providers payload") {
		t.Fatalf("error = %v, want invalid providers payload", err)
	}
}

func TestFetchProvidersFromGitHubRejectsRedirectResponses(t *testing.T) {
	resetProvidersState()
	t.Cleanup(resetProvidersState)

	providersHTTPClient = &http.Client{
		Timeout: providersRemoteTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusFound,
				Header: http.Header{
					"Location": []string{"https://redirected.example/providers.json"},
				},
				Body:    io.NopCloser(strings.NewReader("")),
				Request: req,
			}, nil
		}),
	}

	_, err := fetchProvidersFromGitHub()
	if err == nil {
		t.Fatal("expected redirect response to be rejected")
	}
	if !strings.Contains(err.Error(), "unexpected status: 302") {
		t.Fatalf("error = %v, want unexpected status: 302", err)
	}
}

func TestValidateProvidersPayloadAllowsHTTPBaseURL(t *testing.T) {
	resetProvidersState()
	t.Cleanup(resetProvidersState)

	if err := validateProvidersPayload([]byte(`[{"name":"lan","channel_type":0,"base_url":"http://example.com/v1"}]`)); err != nil {
		t.Fatalf("expected http base_url to be accepted, got %v", err)
	}
}

func TestValidateProvidersPayloadRejectsUnsupportedBaseURLScheme(t *testing.T) {
	resetProvidersState()
	t.Cleanup(resetProvidersState)

	err := validateProvidersPayload([]byte(`[{"name":"unsupported","channel_type":0,"base_url":"ftp://example.com/v1"}]`))
	if err == nil {
		t.Fatal("expected unsupported base_url scheme to be rejected")
	}
	if !strings.Contains(err.Error(), "absolute http or https URL") {
		t.Fatalf("error = %v, want http/https URL validation", err)
	}
}

func TestProvidersGitHubURLPinnedToCommitSHA(t *testing.T) {
	if strings.Contains(providersGitHubURL, "/refs/heads/") {
		t.Fatalf("providersGitHubURL must not point at a mutable branch: %s", providersGitHubURL)
	}

	pattern := regexp.MustCompile(`^https://raw\.githubusercontent\.com/xiaoli0412/octopus-xiaoli-repo/[0-9a-f]{40}/internal/assets/providers\.json$`)
	if !pattern.MatchString(providersGitHubURL) {
		t.Fatalf("providersGitHubURL must be pinned to an immutable commit raw URL: %s", providersGitHubURL)
	}
}
