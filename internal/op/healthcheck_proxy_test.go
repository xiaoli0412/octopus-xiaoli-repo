package op

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

type stubHealthContextDialer struct {
	called bool
	err    error
}

func (d *stubHealthContextDialer) Dial(network, addr string) (net.Conn, error) {
	return nil, errors.New("Dial should not be used when ContextDialer is available")
}

func (d *stubHealthContextDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	d.called = true
	return nil, d.err
}

type blockingHealthDialer struct {
	released chan struct{}
}

func (d blockingHealthDialer) Dial(network, addr string) (net.Conn, error) {
	<-d.released
	return nil, errors.New("unblocked")
}

func TestHealthDialProxyContextUsesContextDialerWhenAvailable(t *testing.T) {
	wantErr := errors.New("context dialer called")
	dialer := &stubHealthContextDialer{err: wantErr}

	_, err := dialProxyContext(context.Background(), dialer, "tcp", "example.com:443")
	if !errors.Is(err, wantErr) {
		t.Fatalf("dialProxyContext() error = %v, want %v", err, wantErr)
	}
	if !dialer.called {
		t.Fatal("dialProxyContext() did not use ContextDialer")
	}
}

func TestHealthDialProxyContextCancelsFallbackDial(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	dialer := blockingHealthDialer{released: make(chan struct{})}
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
