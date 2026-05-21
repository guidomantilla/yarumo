package lifecycle

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewComponent(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil component", func(t *testing.T) {
		t.Parallel()

		c := NewComponent("worker-1")
		if c == nil {
			t.Fatal("expected non-nil component")
		}
	})

	t.Run("carries the given name", func(t *testing.T) {
		t.Parallel()

		c := NewComponent("worker-2")
		if c.Name() != "worker-2" {
			t.Fatalf("expected name %q, got %q", "worker-2", c.Name())
		}
	})

	t.Run("done channel is open at construction", func(t *testing.T) {
		t.Parallel()

		c := NewComponent("worker-3")
		select {
		case <-c.Done():
			t.Fatal("expected Done channel to be open before Stop")
		default:
		}
	})
}

func TestComponent_Start(t *testing.T) {
	t.Parallel()

	t.Run("returns nil with live ctx", func(t *testing.T) {
		t.Parallel()

		c := NewComponent("worker-start-1")
		err := c.Start(context.Background())
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("returns nil even with cancelled ctx", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		c := NewComponent("worker-start-2")
		err := c.Start(ctx)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})
}

func TestComponent_Stop(t *testing.T) {
	t.Parallel()

	t.Run("closes the done channel on first call", func(t *testing.T) {
		t.Parallel()

		c := NewComponent("worker-stop-1")
		err := c.Stop(context.Background())
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		select {
		case <-c.Done():
		default:
			t.Fatal("expected Done channel closed after Stop")
		}
	})

	t.Run("is idempotent across multiple calls", func(t *testing.T) {
		t.Parallel()

		c := NewComponent("worker-stop-2")
		err := c.Stop(context.Background())
		if err != nil {
			t.Fatalf("first Stop returned %v", err)
		}

		err = c.Stop(context.Background())
		if err != nil {
			t.Fatalf("second Stop returned %v", err)
		}
	})

	t.Run("returns ErrShutdownTimeout when ctx already expired", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), -time.Second)
		defer cancel()

		c := NewComponent("worker-stop-3")
		err := c.Stop(ctx)
		if err == nil {
			t.Fatal("expected non-nil error")
		}

		if !errors.Is(err, ErrShutdownTimeout) {
			t.Fatalf("expected error to wrap ErrShutdownTimeout, got %v", err)
		}

		if !errors.Is(err, ErrShutdownFailed) {
			t.Fatalf("expected error to wrap ErrShutdownFailed, got %v", err)
		}
	})

	t.Run("done is closed even when ctx expired", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), -time.Second)
		defer cancel()

		c := NewComponent("worker-stop-4")
		_ = c.Stop(ctx)

		select {
		case <-c.Done():
		default:
			t.Fatal("expected Done channel closed after Stop with expired ctx")
		}
	})
}

func TestComponent_Done(t *testing.T) {
	t.Parallel()

	t.Run("unblocks readers after Stop", func(t *testing.T) {
		t.Parallel()

		c := NewComponent("worker-done-1")

		ready := make(chan struct{})
		done := make(chan struct{})

		go func() {
			close(ready)
			<-c.Done()
			close(done)
		}()

		<-ready

		err := c.Stop(context.Background())
		if err != nil {
			t.Fatalf("Stop returned %v", err)
		}

		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("expected reader to unblock after Stop")
		}
	})
}
