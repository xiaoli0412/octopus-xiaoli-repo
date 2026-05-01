package op

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

type testHTTPServer struct {
	URL  string
	addr string
}

func (s *testHTTPServer) Close() {
	if s == nil {
		return
	}
	unregisterOpTestHTTPHandler(s.addr)
}

var (
	opTestHTTPHookOnce sync.Once
	opTestHTTPMu       sync.RWMutex
	opTestHTTPHandlers = make(map[string]http.Handler)
	opTestHTTPSeq      uint64
)

func installOpTestHTTPHook() {
	opTestHTTPHookOnce.Do(func() {
		transport, ok := http.DefaultTransport.(*http.Transport)
		if !ok {
			return
		}
		transport.Proxy = nil
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			opTestHTTPMu.RLock()
			handler := opTestHTTPHandlers[addr]
			opTestHTTPMu.RUnlock()
			if handler == nil {
				var d net.Dialer
				return d.DialContext(ctx, network, addr)
			}

			clientConn, serverConn := net.Pipe()
			go serveOpTestHTTPConn(serverConn, handler)
			return clientConn, nil
		}
	})
}

func registerOpTestHTTPHandler(addr string, handler http.Handler) {
	opTestHTTPMu.Lock()
	defer opTestHTTPMu.Unlock()
	opTestHTTPHandlers[addr] = handler
}

func unregisterOpTestHTTPHandler(addr string) {
	opTestHTTPMu.Lock()
	defer opTestHTTPMu.Unlock()
	delete(opTestHTTPHandlers, addr)
}

func newIPv4TestServer(t *testing.T, handler http.Handler) *testHTTPServer {
	t.Helper()
	if handler == nil {
		handler = http.NotFoundHandler()
	}

	installOpTestHTTPHook()

	id := atomic.AddUint64(&opTestHTTPSeq, 1)
	host := fmt.Sprintf("op-test-%d.invalid", id)
	addr := host + ":80"
	registerOpTestHTTPHandler(addr, handler)
	t.Cleanup(func() {
		unregisterOpTestHTTPHandler(addr)
	})

	return &testHTTPServer{URL: "http://" + host, addr: addr}
}

func serveOpTestHTTPConn(conn net.Conn, handler http.Handler) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
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
		resp.Close = req.Close
		if err := resp.Write(conn); err != nil {
			_ = resp.Body.Close()
			_ = req.Body.Close()
			return
		}
		_ = resp.Body.Close()
		_, _ = io.Copy(io.Discard, req.Body)
		_ = req.Body.Close()
		if req.Close {
			return
		}
	}
}
