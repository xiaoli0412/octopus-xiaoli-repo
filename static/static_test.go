package static

import "testing"

func TestHasFrontend(t *testing.T) {
	if !HasFrontend() {
		t.Fatal("HasFrontend() = false, want true (static/out should contain frontend assets)")
	}
}
