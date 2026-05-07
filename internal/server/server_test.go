package server

import (
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestNewHTTPServerUsesSecurityTimeouts(t *testing.T) {
	srv := newHTTPServer("127.0.0.1:1088", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	if srv.ReadHeaderTimeout != httpReadHeaderTimeout {
		t.Fatalf("ReadHeaderTimeout = %v, want %v", srv.ReadHeaderTimeout, httpReadHeaderTimeout)
	}
	if srv.ReadTimeout != httpReadTimeout {
		t.Fatalf("ReadTimeout = %v, want %v", srv.ReadTimeout, httpReadTimeout)
	}
	if srv.IdleTimeout != httpIdleTimeout {
		t.Fatalf("IdleTimeout = %v, want %v", srv.IdleTimeout, httpIdleTimeout)
	}
	if srv.Handler == nil {
		t.Fatalf("Handler = nil, want non-nil")
	}
}

func TestNewEngineDoesNotTrustForwardedClientIPHeadersByDefault(t *testing.T) {
	engine, err := newEngine()
	if err != nil {
		t.Fatalf("newEngine() error = %v", err)
	}
	engine.GET("/client-ip", func(c *gin.Context) {
		c.String(http.StatusOK, c.ClientIP())
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	request.RemoteAddr = "203.0.113.10:4567"
	request.Header.Set("X-Forwarded-For", "198.51.100.77")
	request.URL.Path = "/client-ip"

	engine.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("client-ip status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if got := recorder.Body.String(); got != "203.0.113.10" {
		t.Fatalf("ClientIP() = %q, want remote address when proxies are untrusted", got)
	}
}

func TestCloseHTTPServerGracefullyWaitsForInflightRequest(t *testing.T) {
	requestStarted := make(chan struct{})
	allowResponse := make(chan struct{})
	requestDone := make(chan error, 1)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(requestStarted)
		<-allowResponse
		w.WriteHeader(http.StatusNoContent)
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}
	defer listener.Close()

	srv := newHTTPServer(listener.Addr().String(), handler)
	serveDone := make(chan error, 1)
	go func() {
		serveDone <- srv.Serve(listener)
	}()

	client := &http.Client{}
	go func() {
		resp, err := client.Get("http://" + listener.Addr().String())
		if err != nil {
			requestDone <- err
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNoContent {
			requestDone <- errors.New(resp.Status)
			return
		}
		requestDone <- nil
	}()

	select {
	case <-requestStarted:
	case <-time.After(time.Second):
		t.Fatal("request did not reach handler before shutdown")
	}

	closeDone := make(chan error, 1)
	go func() {
		closeDone <- closeHTTPServer(&srv, 250*time.Millisecond)
	}()

	select {
	case err := <-closeDone:
		t.Fatalf("closeHTTPServer() returned early with %v; want graceful wait for inflight request", err)
	case <-time.After(80 * time.Millisecond):
	}

	close(allowResponse)

	select {
	case err := <-closeDone:
		if err != nil {
			t.Fatalf("closeHTTPServer() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("closeHTTPServer() did not finish after inflight request completed")
	}

	select {
	case err := <-requestDone:
		if err != nil {
			t.Fatalf("client request error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("client request did not finish")
	}

	select {
	case err := <-serveDone:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			t.Fatalf("Serve() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Serve() did not exit after shutdown")
	}
}
