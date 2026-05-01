package relay

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newIPv4TestServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen(tcp4) error = %v", err)
	}

	server := &httptest.Server{
		Listener: listener,
		Config:   &http.Server{Handler: handler},
	}
	server.Start()
	return server
}
