package cmd

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestHealthcheckSkipsConfigLoadWhenURLProvided(t *testing.T) {
	oldCfgFile := cfgFile
	oldHealthcheckURL := healthcheckURL
	oldClient := healthcheckHTTPClient
	t.Cleanup(func() {
		cfgFile = oldCfgFile
		healthcheckURL = oldHealthcheckURL
		healthcheckHTTPClient = oldClient
	})

	cfgFile = t.TempDir() + string('/') + "missing-config.json"
	healthcheckURL = "http://example.invalid/healthz"
	healthcheckHTTPClient = &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/healthz" {
			t.Fatalf("unexpected path: %s", req.URL.Path)
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("ok")), Header: make(http.Header), Request: req}, nil
	})}

	if err := healthcheckCmd.RunE(healthcheckCmd, []string{}); err != nil {
		t.Fatalf("healthcheckCmd.RunE() error = %v", err)
	}
}

func TestWrapHealthcheckRequestErrorClassifiesLocalServiceProviderFailure(t *testing.T) {
	err := wrapHealthcheckRequestError(
		"http://127.0.0.1:8080/healthz",
		errors.New("dial tcp 127.0.0.1:8080: socket: The requested service provider could not be loaded or initialized."),
	)

	if err == nil {
		t.Fatal("expected wrapped error, got nil")
	}

	message := err.Error()
	if !strings.Contains(message, "host networking blocker") {
		t.Fatalf("expected host-networking blocker classification, got %q", message)
	}
	if !strings.Contains(message, "127.0.0.1:8080/healthz") {
		t.Fatalf("expected target URL to be included, got %q", message)
	}
}

func TestWrapHealthcheckRequestErrorLeavesRemoteFailuresUntouched(t *testing.T) {
	original := errors.New("dial tcp 10.0.0.2:8080: socket: The requested service provider could not be loaded or initialized.")
	wrapped := wrapHealthcheckRequestError("http://10.0.0.2:8080/healthz", original)

	if !errors.Is(wrapped, original) {
		t.Fatalf("expected wrapped error to preserve original error")
	}
	if wrapped.Error() != original.Error() {
		t.Fatalf("expected remote error to stay unchanged, got %q", wrapped.Error())
	}
}
