package shutdown

import "testing"

func TestShutdownWithoutLoggerDoesNotPanic(t *testing.T) {
	originalLogger := ilog
	originalFuncs := funcs
	ilog = nil
	called := false
	funcs = []func() error{
		func() error {
			called = true
			return nil
		},
	}
	t.Cleanup(func() {
		ilog = originalLogger
		funcs = originalFuncs
	})

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Shutdown() panicked: %v", r)
		}
	}()

	Shutdown()
	if !called {
		t.Fatal("registered shutdown func was not executed")
	}
}
