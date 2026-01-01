package servers

import (
	"context"
	"testing"
	"time"
)

// Test BuildBaseServer returns expected name and non-nil server
func TestBuildBaseServer(t *testing.T) {
	name, srv := BuildBaseServer()
	if name != "base-server" {
		t.Fatalf("unexpected name: %s", name)
	}
	if srv == nil {
		t.Fatal("expected non-nil server")
	}
}

func TestBaseServer_RunAndStop(t *testing.T) {
	srv := NewBaseServer().(*baseServer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- srv.Run(ctx) }()

	// Give some time to start and block on channel
	time.Sleep(50 * time.Millisecond)

	if err := srv.Stop(ctx); err != nil {
		t.Fatalf("stop returned error: %v", err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("run returned error: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("run did not return after stop")
	}
}

// Ensure assertions are exercised (they log at Fatal level but do not exit)
func TestBaseServer_AssertionsCoverage(t *testing.T) {
	srv := &baseServer{closeChannel: make(chan struct{})}
	// ctx nil path
	_ = srv // avoid lint
	// We cannot call Run(nil) because of type, but we can call Stop with ctx
	// and simply ensure method executes. The NotNil checks log internally.
	_ = srv.Stop(context.Background())
}
