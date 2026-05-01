package handlers

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
)

type handlerTestHTTPServer struct {
	URL  string
	addr string
}

func (s *handlerTestHTTPServer) Client() *http.Client {
	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return &http.Client{}
	}
	cloned := transport.Clone()
	cloned.Proxy = nil
	return &http.Client{Transport: cloned}
}

func (s *handlerTestHTTPServer) Close() {
	if s == nil {
		return
	}
	unregisterHandlerTestHTTPHandler(s.addr)
}

var (
	handlerTestHTTPHookOnce sync.Once
	handlerTestHTTPMu       sync.RWMutex
	handlerTestHTTPHandlers = make(map[string]http.Handler)
	handlerTestHTTPSeq      uint64
)

func installHandlerTestHTTPHook() {
	handlerTestHTTPHookOnce.Do(func() {
		transport, ok := http.DefaultTransport.(*http.Transport)
		if !ok {
			return
		}
		transport.Proxy = nil
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			handlerTestHTTPMu.RLock()
			handler := handlerTestHTTPHandlers[addr]
			handlerTestHTTPMu.RUnlock()
			if handler == nil {
				var d net.Dialer
				return d.DialContext(ctx, network, addr)
			}

			clientConn, serverConn := net.Pipe()
			go serveHandlerTestHTTPConn(serverConn, handler)
			return clientConn, nil
		}
	})
}

func registerHandlerTestHTTPHandler(addr string, handler http.Handler) {
	handlerTestHTTPMu.Lock()
	defer handlerTestHTTPMu.Unlock()
	handlerTestHTTPHandlers[addr] = handler
}

func unregisterHandlerTestHTTPHandler(addr string) {
	handlerTestHTTPMu.Lock()
	defer handlerTestHTTPMu.Unlock()
	delete(handlerTestHTTPHandlers, addr)
}

func newIPv4TestServer(t *testing.T, handler http.Handler) *handlerTestHTTPServer {
	t.Helper()
	if handler == nil {
		handler = http.NotFoundHandler()
	}

	installHandlerTestHTTPHook()

	id := atomic.AddUint64(&handlerTestHTTPSeq, 1)
	host := fmt.Sprintf("handler-test-%d.invalid", id)
	addr := host + ":80"
	registerHandlerTestHTTPHandler(addr, handler)
	t.Cleanup(func() {
		unregisterHandlerTestHTTPHandler(addr)
	})

	return &handlerTestHTTPServer{URL: "http://" + host, addr: addr}
}

func serveHandlerTestHTTPConn(conn net.Conn, handler http.Handler) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	req, err := http.ReadRequest(reader)
	if err != nil {
		return
	}
	req.RemoteAddr = conn.RemoteAddr().String()

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if req.Method == http.MethodHead {
		recorder.Body.Reset()
	}
	resp := recorder.Result()
	resp.Close = true
	if err := resp.Write(conn); err != nil {
		_ = resp.Body.Close()
		_ = req.Body.Close()
		return
	}
	_ = resp.Body.Close()
	_, _ = io.Copy(io.Discard, req.Body)
	_ = req.Body.Close()
}
