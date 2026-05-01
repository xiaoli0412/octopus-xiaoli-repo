package client

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

func TestGetHTTPClientCustomProxyAllowsSocksScheme(t *testing.T) {
	if _, err := GetHTTPClientCustomProxy("socks://127.0.0.1:1080"); err != nil {
		t.Fatalf("GetHTTPClientCustomProxy() error = %v, want nil", err)
	}
}

type stubContextDialer struct {
	called bool
	err    error
}

func (d *stubContextDialer) Dial(network, addr string) (net.Conn, error) {
	return nil, errors.New("Dial should not be used when ContextDialer is available")
}

func (d *stubContextDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	d.called = true
	return nil, d.err
}

type blockingDialer struct {
	released chan struct{}
}

func (d blockingDialer) Dial(network, addr string) (net.Conn, error) {
	<-d.released
	return nil, errors.New("unblocked")
}

func TestDialProxyContextUsesContextDialerWhenAvailable(t *testing.T) {
	wantErr := errors.New("context dialer called")
	dialer := &stubContextDialer{err: wantErr}

	_, err := dialProxyContext(context.Background(), dialer, "tcp", "example.com:443")
	if !errors.Is(err, wantErr) {
		t.Fatalf("dialProxyContext() error = %v, want %v", err, wantErr)
	}
	if !dialer.called {
		t.Fatal("dialProxyContext() did not use ContextDialer")
	}
}

func TestDialProxyContextCancelsFallbackDial(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	dialer := blockingDialer{released: make(chan struct{})}
	defer close(dialer.released)

	done := make(chan error, 1)
	go func() {
		_, err := dialProxyContext(ctx, dialer, "tcp", "example.com:443")
		done <- err
	}()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("dialProxyContext() error = %v, want %v", err, context.Canceled)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("dialProxyContext() did not return after context cancellation")
	}
}
