package delayer

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

// captureDrops returns a thread-safe DropHandler that records every
// drop, and a getter returning the recorded errors.
func captureDrops() (DropHandler, func() []error) {
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

// timingSink returns a Handler[int] that records the wall-clock time of
// every message it receives along with the payload.
func timingSink() (messaging.Handler[int], func() []timedMsg) {
	var mu sync.Mutex

	captured := []timedMsg{}

	handler := func(_ context.Context, msg messaging.Message[int]) error {
		mu.Lock()
		defer mu.Unlock()

		captured = append(captured, timedMsg{payload: msg.Payload, at: time.Now()})

		return nil
	}

	get := func() []timedMsg {
		mu.Lock()
		defer mu.Unlock()

		out := make([]timedMsg, len(captured))
		copy(out, captured)

		return out
	}

	return handler, get
}

type timedMsg struct {
	payload int
	at      time.Time
}

func TestNewDelayer(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil delayer", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		d := NewDelayer("test", src, dst)
		if d == nil {
			t.Fatal("expected non-nil delayer")
		}
	})

	t.Run("carries the given name", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		d := NewDelayer("orders-delayer", src, dst)
		if d.Name() != "orders-delayer" {
			t.Fatalf("expected name orders-delayer, got %q", d.Name())
		}
	})
}

func TestDelayer_FixedDelay(t *testing.T) {
	t.Parallel()

	t.Run("waits the configured delay before forwarding", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		sink, get := timingSink()

		_, err := dst.Subscribe(sink)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		delay := 50 * time.Millisecond

		d := NewDelayer("test", src, dst, WithFixedDelay[int](delay))

		err = d.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() { _ = d.Stop(context.Background()) })

		sentAt := time.Now()

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		deadline := time.After(2 * time.Second)
		for len(get()) != 1 {
			select {
			case <-deadline:
				t.Fatal("message never delivered")
			case <-time.After(5 * time.Millisecond):
			}
		}

		got := get()[0]

		elapsed := got.at.Sub(sentAt)
		if elapsed < delay {
			t.Fatalf("message delivered too early: elapsed=%s want>=%s", elapsed, delay)
		}
	})
}

func TestDelayer_DelayFn(t *testing.T) {
	t.Parallel()

	t.Run("uses caller-computed delay per message", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		sink, get := timingSink()

		_, err := dst.Subscribe(sink)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		// Payload encodes the delay in milliseconds.
		delayFn := func(_ context.Context, msg messaging.Message[int]) time.Duration {
			return time.Duration(msg.Payload) * time.Millisecond
		}

		d := NewDelayer("test", src, dst, WithDelayFn[int](delayFn))

		err = d.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() { _ = d.Stop(context.Background()) })

		// Send in reverse delay order: 80ms, 30ms, 50ms.
		err = src.Send(context.Background(), messaging.Message[int]{Payload: 80})
		if err != nil {
			t.Fatalf("send 80: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 30})
		if err != nil {
			t.Fatalf("send 30: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 50})
		if err != nil {
			t.Fatalf("send 50: %v", err)
		}

		deadline := time.After(2 * time.Second)
		for len(get()) != 3 {
			select {
			case <-deadline:
				t.Fatalf("only %d messages delivered", len(get()))
			case <-time.After(5 * time.Millisecond):
			}
		}

		out := get()
		// Order must be delay-ascending: 30, 50, 80.
		if out[0].payload != 30 {
			t.Fatalf("first delivered payload: want 30, got %d", out[0].payload)
		}

		if out[1].payload != 50 {
			t.Fatalf("second delivered payload: want 50, got %d", out[1].payload)
		}

		if out[2].payload != 80 {
			t.Fatalf("third delivered payload: want 80, got %d", out[2].payload)
		}
	})
}

func TestDelayer_ExpirationTime(t *testing.T) {
	t.Parallel()

	t.Run("uses Headers.ExpirationTime when no other strategy is set", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		sink, get := timingSink()

		_, err := dst.Subscribe(sink)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		d := NewDelayer("test", src, dst)

		err = d.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() { _ = d.Stop(context.Background()) })

		deliverAt := time.Now().Add(40 * time.Millisecond)

		err = src.Send(context.Background(), messaging.Message[int]{
			Payload: 1,
			Headers: messaging.Headers{ExpirationTime: deliverAt},
		})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		deadline := time.After(2 * time.Second)
		for len(get()) != 1 {
			select {
			case <-deadline:
				t.Fatal("never delivered")
			case <-time.After(5 * time.Millisecond):
			}
		}

		got := get()[0]
		if got.at.Before(deliverAt) {
			t.Fatalf("delivered before ExpirationTime: at=%s want>=%s", got.at, deliverAt)
		}
	})

	t.Run("ExpirationTime in the past forwards immediately", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		var delivered atomic.Int32

		_, err := dst.Subscribe(func(_ context.Context, _ messaging.Message[int]) error {
			delivered.Add(1)

			return nil
		})
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		d := NewDelayer("test", src, dst)

		err = d.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() { _ = d.Stop(context.Background()) })

		err = src.Send(context.Background(), messaging.Message[int]{
			Payload: 1,
			Headers: messaging.Headers{ExpirationTime: time.Now().Add(-1 * time.Second)},
		})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		// Pipeline source is synchronous and the immediate forward path
		// is in the same goroutine, so delivery should be observed by
		// the time Send returns.
		if got := delivered.Load(); got != 1 {
			t.Fatalf("expected 1 delivered, got %d", got)
		}
	})
}

func TestDelayer_ImmediateForward(t *testing.T) {
	t.Parallel()

	t.Run("delay <= 0 forwards immediately without scheduling", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		var delivered atomic.Int32

		_, err := dst.Subscribe(func(_ context.Context, _ messaging.Message[int]) error {
			delivered.Add(1)

			return nil
		})
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		delayFn := func(_ context.Context, _ messaging.Message[int]) time.Duration {
			return 0
		}

		d := NewDelayer("test", src, dst, WithDelayFn[int](delayFn))

		err = d.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() { _ = d.Stop(context.Background()) })

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got := delivered.Load(); got != 1 {
			t.Fatalf("expected immediate delivery, got %d", got)
		}
	})
}

func TestDelayer_MaxPending(t *testing.T) {
	t.Parallel()

	t.Run("drops messages when bound is reached", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		_, err := dst.Subscribe(func(_ context.Context, _ messaging.Message[int]) error {
			return nil
		})
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		dropHandler, getDrops := captureDrops()

		// Long delay + max 2 pending → first two scheduled, rest dropped.
		d := NewDelayer("test", src, dst,
			WithFixedDelay[int](5*time.Second),
			WithMaxPending[int](2),
			WithDropHandler[int](dropHandler),
		)

		err = d.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() { _ = d.Stop(context.Background()) })

		for i := range 5 {
			err = src.Send(context.Background(), messaging.Message[int]{Payload: i})
			if err != nil {
				t.Fatalf("send %d: %v", i, err)
			}
		}

		drops := getDrops()
		if len(drops) != 3 {
			t.Fatalf("expected 3 drops (5 sent, 2 capacity), got %d", len(drops))
		}

		for i, err := range drops {
			if !errors.Is(err, ErrMaxPendingExceeded) {
				t.Fatalf("drop %d: expected ErrMaxPendingExceeded, got %v", i, err)
			}

			if !errors.Is(err, ErrDelayerFailed) {
				t.Fatalf("drop %d: expected ErrDelayerFailed, got %v", i, err)
			}
		}
	})
}

func TestDelayer_ForwardFailure(t *testing.T) {
	t.Parallel()

	t.Run("reports ErrForwardFailed when destination errs (immediate path)", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()

		boom := errors.New("dst down")
		dst := &failingChannel[int]{err: boom}

		errHandler, getErrs := captureErrors()

		// delay <= 0 → immediate forward, which fails.
		delayFn := func(_ context.Context, _ messaging.Message[int]) time.Duration {
			return 0
		}

		d := NewDelayer("test", src, dst,
			WithDelayFn[int](delayFn),
			WithErrorHandler[int](errHandler),
		)

		err := d.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() { _ = d.Stop(context.Background()) })

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		errs := getErrs()
		if len(errs) != 1 {
			t.Fatalf("expected 1 captured error, got %d", len(errs))
		}

		if !errors.Is(errs[0], ErrForwardFailed) {
			t.Fatalf("expected ErrForwardFailed, got %v", errs[0])
		}

		if !errors.Is(errs[0], boom) {
			t.Fatalf("expected wrapped boom, got %v", errs[0])
		}
	})

	t.Run("forward failure does not propagate to source caller", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()

		boom := errors.New("dst down")
		dst := &failingChannel[int]{err: boom}

		delayFn := func(_ context.Context, _ messaging.Message[int]) time.Duration {
			return 0
		}

		d := NewDelayer("test", src, dst,
			WithDelayFn[int](delayFn),
			WithErrorHandler[int](messaging.SilentErrorHandler),
		)

		err := d.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() { _ = d.Stop(context.Background()) })

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("expected nil from src.Send (delayer swallows), got %v", err)
		}
	})
}

func TestDelayer_Lifecycle(t *testing.T) {
	t.Parallel()

	t.Run("Start is idempotent", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		var n atomic.Int32

		_, err := dst.Subscribe(func(_ context.Context, _ messaging.Message[int]) error {
			n.Add(1)

			return nil
		})
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		// Use 0 delay so the immediate path triggers and we can count.
		delayFn := func(_ context.Context, _ messaging.Message[int]) time.Duration { return 0 }

		d := NewDelayer("test", src, dst, WithDelayFn[int](delayFn))

		err = d.Start(context.Background())
		if err != nil {
			t.Fatalf("first start: %v", err)
		}

		err = d.Start(context.Background())
		if err != nil {
			t.Fatalf("second start: %v", err)
		}

		t.Cleanup(func() { _ = d.Stop(context.Background()) })

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got := n.Load(); got != 1 {
			t.Fatalf("dst should receive once (no double subscription), got %d", got)
		}
	})

	t.Run("Stop is idempotent", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		d := NewDelayer("test", src, dst)

		err := d.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = d.Stop(context.Background())
		if err != nil {
			t.Fatalf("first stop: %v", err)
		}

		err = d.Stop(context.Background())
		if err != nil {
			t.Fatalf("second stop: %v", err)
		}
	})

	t.Run("Done closes after Stop", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		d := NewDelayer("test", src, dst)

		err := d.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		select {
		case <-d.Done():
			t.Fatal("Done closed before Stop")
		default:
		}

		err = d.Stop(context.Background())
		if err != nil {
			t.Fatalf("stop: %v", err)
		}

		select {
		case <-d.Done():
		default:
			t.Fatal("Done not closed after Stop")
		}
	})

	t.Run("Stop with expired ctx returns ErrShutdownTimeout", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		d := NewDelayer("test", src, dst)

		err := d.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err = d.Stop(ctx)
		if !errors.Is(err, lifecycle.ErrShutdownTimeout) {
			t.Fatalf("expected ErrShutdownTimeout, got %v", err)
		}
	})

	t.Run("Stop drops pending messages (best-effort scheduled semantics)", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		var delivered atomic.Int32

		_, err := dst.Subscribe(func(_ context.Context, _ messaging.Message[int]) error {
			delivered.Add(1)

			return nil
		})
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		d := NewDelayer("test", src, dst, WithFixedDelay[int](5*time.Second))

		err = d.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		// Stop before the 5s delay elapses → pending msg is dropped.
		err = d.Stop(context.Background())
		if err != nil {
			t.Fatalf("stop: %v", err)
		}

		if got := delivered.Load(); got != 0 {
			t.Fatalf("expected 0 delivered (dropped on stop), got %d", got)
		}
	})
}

func TestDelayer_Options(t *testing.T) {
	t.Parallel()

	t.Run("WithErrorHandler(nil) preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions[int](WithErrorHandler[int](nil))
		if opts.errorHandler == nil {
			t.Fatal("expected default error handler preserved on nil arg")
		}
	})

	t.Run("WithDelayFn(nil) preserves previous", func(t *testing.T) {
		t.Parallel()

		first := func(_ context.Context, _ messaging.Message[int]) time.Duration { return time.Second }
		opts := NewOptions(WithDelayFn[int](first), WithDelayFn[int](nil))
		if opts.delayFn == nil {
			t.Fatal("expected previously installed DelayFn preserved on nil arg")
		}
	})

	t.Run("WithFixedDelay rejects non-positive", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions[int](WithFixedDelay[int](-5 * time.Second))
		if opts.fixedDelay != 0 {
			t.Fatalf("expected 0 fixedDelay (non-positive ignored), got %s", opts.fixedDelay)
		}
	})

	t.Run("WithMaxPending rejects non-positive", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions[int](WithMaxPending[int](0))
		if opts.maxPending != defaultMaxPending {
			t.Fatalf("expected defaultMaxPending preserved, got %d", opts.maxPending)
		}
	})

	t.Run("defaults install messaging.DefaultErrorHandler", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions[int]()
		if opts.errorHandler == nil {
			t.Fatal("default error handler should be installed")
		}

		if opts.maxPending != defaultMaxPending {
			t.Fatalf("default maxPending: got %d, want %d", opts.maxPending, defaultMaxPending)
		}
	})
}

// failingChannel is a Channel[T] test double whose Send always returns
// the configured err. Subscribe is a no-op returning a no-op Cancel.
type failingChannel[T any] struct {
	err error
}

func (c *failingChannel[T]) Send(_ context.Context, _ messaging.Message[T]) error {
	return c.err
}

func (c *failingChannel[T]) Subscribe(_ messaging.Handler[T]) (messaging.Cancel, error) {
	return func() {}, nil
}
