package managed

import (
	"context"
	"testing"
	"time"
)

func TestBuildBaseWorker(t *testing.T) {
	t.Run("build succeeds and stop completes", func(t *testing.T) {
		errCh := make(chan error, 1)

		component, stopFn, err := BuildBaseWorker(context.Background(), "test-base", nil, errCh)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		if component.name != "test-base" {
			t.Fatalf("expected name test-base, got %s", component.name)
		}

		if component.internal == nil {
			t.Fatal("expected non-nil internal")
		}

		if stopFn == nil {
			t.Fatal("expected non-nil stopFn")
		}

		stopFn(context.Background(), 5*time.Second)

		select {
		case err := <-errCh:
			t.Fatalf("unexpected error: %v", err)
		default:
		}
	})

	t.Run("stop with short timeout logs error", func(t *testing.T) {
		errCh := make(chan error, 1)

		_, stopFn, err := BuildBaseWorker(context.Background(), "test-base-timeout", nil, errCh)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		stopFn(ctx, time.Nanosecond)
	})
}
