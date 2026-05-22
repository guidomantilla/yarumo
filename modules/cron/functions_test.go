package cron

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/guidomantilla/yarumo/common/lifecycle"
)

func TestBuildScheduler(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil scheduler and closeFn", func(t *testing.T) {
		t.Parallel()

		errChan := make(chan error, 1)

		scheduler, closeFn, err := BuildScheduler(context.Background(), "build-1", errChan)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		if scheduler == nil {
			t.Fatal("expected non-nil scheduler")
		}

		if closeFn == nil {
			t.Fatal("expected non-nil closeFn")
		}

		closeFn(context.Background(), time.Second)
	})

	t.Run("scheduler carries the given name", func(t *testing.T) {
		t.Parallel()

		errChan := make(chan error, 1)

		scheduler, closeFn, err := BuildScheduler(context.Background(), "build-named", errChan)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		defer closeFn(context.Background(), time.Second)

		if scheduler.Name() != "build-named" {
			t.Fatalf("expected name %q, got %q", "build-named", scheduler.Name())
		}
	})

	t.Run("returned closeFn drains the background goroutine before returning", func(t *testing.T) {
		t.Parallel()

		errChan := make(chan error, 1)

		scheduler, closeFn, err := BuildScheduler(context.Background(), "build-drain", errChan)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		closeFn(context.Background(), time.Second)

		// Done should already be closed (Stop ran inside closeFn) and the
		// background lifecycle.Start goroutine should already be gone.
		select {
		case <-scheduler.Done():
		default:
			t.Fatal("expected scheduler Done closed after closeFn returned")
		}
	})

	t.Run("closeFn is safe to call from defer with the same ctx", func(t *testing.T) {
		t.Parallel()

		errChan := make(chan error, 1)
		ctx := context.Background()

		_, closeFn, err := BuildScheduler(ctx, "build-defer", errChan)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		defer closeFn(ctx, time.Second)
	})

	t.Run("registered job fires on schedule", func(t *testing.T) {
		t.Parallel()

		errChan := make(chan error, 1)

		scheduler, closeFn, err := BuildScheduler(context.Background(), "build-ticks", errChan)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		defer closeFn(context.Background(), time.Second)

		var counter atomic.Int32

		_, err = scheduler.AddFunc("@every 100ms", func() { counter.Add(1) })
		if err != nil {
			t.Fatalf("AddFunc returned %v", err)
		}

		// Wait long enough for at least one tick at 100ms cadence.
		time.Sleep(300 * time.Millisecond)

		if counter.Load() == 0 {
			t.Fatal("expected at least one tick within 300ms")
		}
	})

	t.Run("matches the BuildSchedulerFn signature", func(t *testing.T) {
		t.Parallel()

		var fn BuildSchedulerFn = BuildScheduler

		errChan := make(chan error, 1)

		_, closeFn, err := fn(context.Background(), "build-fn", errChan)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		closeFn(context.Background(), time.Second)
	})

	t.Run("errChan accepts startup errors without blocking", func(t *testing.T) {
		t.Parallel()

		// Buffered channel of zero capacity: a non-blocking send by
		// lifecycle.Start should fall through the default arm. The
		// build itself must still succeed.
		errChan := make(chan error)

		_, closeFn, err := BuildScheduler(context.Background(), "build-errchan", lifecycle.ErrChan(errChan))
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		closeFn(context.Background(), time.Second)
	})
}
