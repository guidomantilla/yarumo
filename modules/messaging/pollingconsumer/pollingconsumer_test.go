package pollingconsumer

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/messaging"
)

// captureErrors returns a thread-safe ErrorHandler that appends every
// reported error, and a getter that returns a defensive copy.
func captureErrors() (messaging.ErrorHandler, func() []error) {
	var mu sync.Mutex

	captured := []error{}

	handler := func(_ context.Context, _ any, err error) {
		mu.Lock()
		defer mu.Unlock()

		captured = append(captured, err)
	}

	get := func() []error {
		mu.Lock()
		defer mu.Unlock()

		out := make([]error, len(captured))
		copy(out, captured)

		return out
	}

	return handler, get
}

// waitFor polls cond until it returns true or 2s elapses. Returns true
// if cond was satisfied within the window, false otherwise. The fixed
// timeout matches the rest of this file's "wait long enough that flaky
// CI hosts still observe completion".
func waitFor(cond func() bool) bool {
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}

		time.Sleep(5 * time.Millisecond)
	}

	return cond()
}

func TestNewPollingConsumer(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil consumer", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPollableChannel[int]()

		handler := func(_ context.Context, _ messaging.Message[int]) error { return nil }

		c := NewPollingConsumer("test", src, handler)
		if c == nil {
			t.Fatal("expected non-nil consumer")
		}
	})

	t.Run("carries the given name", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPollableChannel[int]()

		handler := func(_ context.Context, _ messaging.Message[int]) error { return nil }

		c := NewPollingConsumer("orders-poller", src, handler)
		if c.Name() != "orders-poller" {
			t.Fatalf("expected name orders-poller, got %q", c.Name())
		}
	})
}

func TestPollingConsumer_HappyPath(t *testing.T) {
	t.Parallel()

	t.Run("delivers every message to the handler", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPollableChannel[int]()

		var received atomic.Int32

		handler := func(_ context.Context, _ messaging.Message[int]) error {
			received.Add(1)

			return nil
		}

		c := NewPollingConsumer("test", src, handler)

		err := c.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() { _ = c.Stop(context.Background()) })

		for i := range 5 {
			err = src.Send(context.Background(), messaging.Message[int]{Payload: i})
			if err != nil {
				t.Fatalf("send %d: %v", i, err)
			}
		}

		if !waitFor(func() bool { return received.Load() == 5 }) {
			t.Fatalf("expected 5 messages handled, got %d", received.Load())
		}
	})

	t.Run("preserves payload end-to-end", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPollableChannel[int]()

		seen := make(chan messaging.Message[int], 1)

		handler := func(_ context.Context, msg messaging.Message[int]) error {
			seen <- msg

			return nil
		}

		c := NewPollingConsumer("test", src, handler)

		err := c.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() { _ = c.Stop(context.Background()) })

		original := messaging.Message[int]{
			Payload: 42,
			Headers: messaging.Headers{MessageID: "msg-1", Type: "test"},
		}

		err = src.Send(context.Background(), original)
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		select {
		case got := <-seen:
			if got.Payload != 42 {
				t.Fatalf("payload: want 42 got %d", got.Payload)
			}

			if got.Headers.MessageID != "msg-1" {
				t.Fatalf("MessageID: want msg-1 got %q", got.Headers.MessageID)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("never received")
		}
	})
}

func TestPollingConsumer_MaxConcurrency(t *testing.T) {
	t.Parallel()

	t.Run("multiple workers process in parallel", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPollableChannel[int](messaging.WithBufferSize(16))

		var inFlight atomic.Int32

		var maxInFlight atomic.Int32

		gate := make(chan struct{})

		handler := func(_ context.Context, _ messaging.Message[int]) error {
			cur := inFlight.Add(1)
			defer inFlight.Add(-1)

			for {
				prev := maxInFlight.Load()
				if cur <= prev {
					break
				}

				if maxInFlight.CompareAndSwap(prev, cur) {
					break
				}
			}

			// Block until gate is closed so multiple handlers are alive
			// at the same time.
			<-gate

			return nil
		}

		c := NewPollingConsumer("test", src, handler, WithMaxConcurrency(4))

		err := c.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() {
			close(gate)
			_ = c.Stop(context.Background())
		})

		for i := range 4 {
			err = src.Send(context.Background(), messaging.Message[int]{Payload: i})
			if err != nil {
				t.Fatalf("send %d: %v", i, err)
			}
		}

		if !waitFor(func() bool { return maxInFlight.Load() >= 2 }) {
			t.Fatalf("expected at least 2 concurrent handlers, observed max %d", maxInFlight.Load())
		}
	})
}

func TestPollingConsumer_HandlerError(t *testing.T) {
	t.Parallel()

	t.Run("reports ErrHandlerFailed via ErrorHandler", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPollableChannel[int]()

		boom := errors.New("handler boom")
		handler := func(_ context.Context, _ messaging.Message[int]) error {
			return boom
		}

		errHandler, getErrs := captureErrors()

		c := NewPollingConsumer("test", src, handler, WithErrorHandler(errHandler))

		err := c.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() { _ = c.Stop(context.Background()) })

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if !waitFor(func() bool { return len(getErrs()) == 1 }) {
			t.Fatalf("expected 1 error reported, got %d", len(getErrs()))
		}

		errs := getErrs()
		if !errors.Is(errs[0], ErrHandlerFailed) {
			t.Fatalf("expected ErrHandlerFailed, got %v", errs[0])
		}

		if !errors.Is(errs[0], boom) {
			t.Fatalf("expected wrapped boom, got %v", errs[0])
		}

		if !errors.Is(errs[0], ErrPollingConsumerFailed) {
			t.Fatalf("expected ErrPollingConsumerFailed, got %v", errs[0])
		}
	})
}

func TestPollingConsumer_HandlerPanic(t *testing.T) {
	t.Parallel()

	t.Run("reports ErrHandlerPanic via ErrorHandler", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPollableChannel[int]()

		handler := func(_ context.Context, _ messaging.Message[int]) error {
			panic("oh no")
		}

		errHandler, getErrs := captureErrors()

		c := NewPollingConsumer("test", src, handler, WithErrorHandler(errHandler))

		err := c.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() { _ = c.Stop(context.Background()) })

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if !waitFor(func() bool { return len(getErrs()) == 1 }) {
			t.Fatalf("expected 1 error reported, got %d", len(getErrs()))
		}

		errs := getErrs()
		if !errors.Is(errs[0], ErrHandlerPanic) {
			t.Fatalf("expected ErrHandlerPanic, got %v", errs[0])
		}
	})

	t.Run("worker survives a handler panic and keeps polling", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPollableChannel[int]()

		var ok atomic.Int32

		handler := func(_ context.Context, msg messaging.Message[int]) error {
			if msg.Payload == 0 {
				panic("first one")
			}

			ok.Add(1)

			return nil
		}

		c := NewPollingConsumer("test", src, handler, WithErrorHandler(messaging.SilentErrorHandler))

		err := c.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() { _ = c.Stop(context.Background()) })

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 0})
		if err != nil {
			t.Fatalf("send 0: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("send 1: %v", err)
		}

		if !waitFor(func() bool { return ok.Load() == 1 }) {
			t.Fatalf("expected 1 successful handle after panic, got %d", ok.Load())
		}
	})
}

func TestPollingConsumer_SourceClose(t *testing.T) {
	t.Parallel()

	t.Run("workers exit cleanly when source closes", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPollableChannel[int]()

		var received atomic.Int32

		handler := func(_ context.Context, _ messaging.Message[int]) error {
			received.Add(1)

			return nil
		}

		c := NewPollingConsumer("test", src, handler, WithMaxConcurrency(2))

		err := c.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		// Send a couple of messages then close the source.
		for i := range 3 {
			err = src.Send(context.Background(), messaging.Message[int]{Payload: i})
			if err != nil {
				t.Fatalf("send %d: %v", i, err)
			}
		}

		if !waitFor(func() bool { return received.Load() == 3 }) {
			t.Fatalf("expected 3 received before close, got %d", received.Load())
		}

		err = src.Close()
		if err != nil {
			t.Fatalf("close: %v", err)
		}

		// Done must close once workers observe ErrChannelClosed.
		select {
		case <-c.Done():
		case <-time.After(2 * time.Second):
			t.Fatal("workers did not exit after source close")
		}
	})
}

func TestPollingConsumer_Lifecycle(t *testing.T) {
	t.Parallel()

	t.Run("Start is idempotent", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPollableChannel[int]()

		var n atomic.Int32

		handler := func(_ context.Context, _ messaging.Message[int]) error {
			n.Add(1)

			return nil
		}

		c := NewPollingConsumer("test", src, handler)

		err := c.Start(context.Background())
		if err != nil {
			t.Fatalf("first start: %v", err)
		}

		err = c.Start(context.Background())
		if err != nil {
			t.Fatalf("second start: %v", err)
		}

		t.Cleanup(func() { _ = c.Stop(context.Background()) })

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if !waitFor(func() bool { return n.Load() == 1 }) {
			t.Fatalf("expected 1 handle (single worker pool), got %d", n.Load())
		}
	})

	t.Run("Stop is idempotent", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPollableChannel[int]()

		handler := func(_ context.Context, _ messaging.Message[int]) error { return nil }

		c := NewPollingConsumer("test", src, handler)

		err := c.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = c.Stop(context.Background())
		if err != nil {
			t.Fatalf("first stop: %v", err)
		}

		err = c.Stop(context.Background())
		if err != nil {
			t.Fatalf("second stop: %v", err)
		}
	})

	t.Run("Done closes after Stop", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPollableChannel[int]()

		handler := func(_ context.Context, _ messaging.Message[int]) error { return nil }

		c := NewPollingConsumer("test", src, handler)

		err := c.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		select {
		case <-c.Done():
			t.Fatal("Done closed before Stop")
		default:
		}

		err = c.Stop(context.Background())
		if err != nil {
			t.Fatalf("stop: %v", err)
		}

		select {
		case <-c.Done():
		case <-time.After(2 * time.Second):
			t.Fatal("Done not closed after Stop")
		}
	})

	t.Run("Stop with expired ctx returns ErrShutdownTimeout when handler is blocked", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPollableChannel[int]()

		block := make(chan struct{})

		handler := func(_ context.Context, _ messaging.Message[int]) error {
			<-block

			return nil
		}

		c := NewPollingConsumer("test", src, handler)

		err := c.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		// Send one message and let the worker pick it up.
		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		// Give the worker a brief moment to dequeue.
		time.Sleep(50 * time.Millisecond)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err = c.Stop(ctx)
		if !errors.Is(err, lifecycle.ErrShutdownTimeout) {
			t.Fatalf("expected ErrShutdownTimeout, got %v", err)
		}

		// Release the blocked handler so the worker exits cleanly.
		close(block)
	})

	t.Run("in-flight handler completes before Done closes", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPollableChannel[int]()

		release := make(chan struct{})

		var finished atomic.Bool

		handler := func(_ context.Context, _ messaging.Message[int]) error {
			<-release

			finished.Store(true)

			return nil
		}

		c := NewPollingConsumer("test", src, handler)

		err := c.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		// Let the worker pick up the message.
		time.Sleep(50 * time.Millisecond)

		// Stop initiates cancellation; release the handler so it
		// completes before Stop's wait deadline.
		stopDone := make(chan error, 1)
		go func() {
			stopDone <- c.Stop(context.Background())
		}()

		close(release)

		err = <-stopDone
		if err != nil {
			t.Fatalf("stop: %v", err)
		}

		if !finished.Load() {
			t.Fatal("handler did not complete before Done closed")
		}
	})
}

func TestPollingConsumer_PollInterval(t *testing.T) {
	t.Parallel()

	t.Run("worker honors poll interval between calls", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPollableChannel[int](messaging.WithBufferSize(8))

		var received atomic.Int32

		handler := func(_ context.Context, _ messaging.Message[int]) error {
			received.Add(1)

			return nil
		}

		interval := 30 * time.Millisecond

		c := NewPollingConsumer("test", src, handler, WithPollInterval(interval))

		// Preload two messages so the worker can pull them back-to-back
		// (with the interval pause in between).
		for i := range 2 {
			err := src.Send(context.Background(), messaging.Message[int]{Payload: i})
			if err != nil {
				t.Fatalf("send %d: %v", i, err)
			}
		}

		started := time.Now()

		err := c.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() { _ = c.Stop(context.Background()) })

		if !waitFor(func() bool { return received.Load() == 2 }) {
			t.Fatalf("expected 2 received, got %d", received.Load())
		}

		elapsed := time.Since(started)
		if elapsed < interval {
			t.Fatalf("expected >= one interval between two messages, elapsed=%s interval=%s", elapsed, interval)
		}
	})
}

func TestPollingConsumer_Options(t *testing.T) {
	t.Parallel()

	t.Run("WithErrorHandler(nil) preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithErrorHandler(nil))
		if opts.errorHandler == nil {
			t.Fatal("expected default error handler preserved on nil arg")
		}
	})

	t.Run("WithMaxConcurrency rejects non-positive", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMaxConcurrency(0))
		if opts.maxConcurrency != defaultMaxConcurrency {
			t.Fatalf("expected default maxConcurrency preserved, got %d", opts.maxConcurrency)
		}
	})

	t.Run("WithPollInterval rejects non-positive", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithPollInterval(-5 * time.Second))
		if opts.pollInterval != 0 {
			t.Fatalf("expected 0 pollInterval preserved, got %s", opts.pollInterval)
		}
	})

	t.Run("defaults install messaging.DefaultErrorHandler", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()
		if opts.errorHandler == nil {
			t.Fatal("default error handler should be installed")
		}

		if opts.maxConcurrency != defaultMaxConcurrency {
			t.Fatalf("default maxConcurrency: got %d want %d", opts.maxConcurrency, defaultMaxConcurrency)
		}
	})
}
