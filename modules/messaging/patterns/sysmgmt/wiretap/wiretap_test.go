package wiretap

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

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

// counter returns a Handler[int] that increments the returned int32
// once per dispatched message.
func counter() (messaging.Handler[int], func() int32) {
	var n int32

	handler := func(_ context.Context, _ messaging.Message[int]) error {
		atomic.AddInt32(&n, 1)

		return nil
	}

	get := func() int32 {
		return atomic.LoadInt32(&n)
	}

	return handler, get
}

func TestNewWiretap(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil component", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		tap := messaging.NewPipelineChannel[int]()

		c := NewWiretap("test", src, dst, tap)
		if c == nil {
			t.Fatal("expected non-nil component")
		}
	})

	t.Run("carries the given name", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		tap := messaging.NewPipelineChannel[int]()

		c := NewWiretap("orders-tap", src, dst, tap)
		if c.Name() != "orders-tap" {
			t.Fatalf("expected name orders-tap, got %q", c.Name())
		}
	})
}

func TestWiretap_HappyPath(t *testing.T) {
	t.Parallel()

	t.Run("forwards every message to both dst and tap", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		tap := messaging.NewPipelineChannel[int]()

		dstHandler, getDst := counter()
		tapHandler, getTap := counter()

		_, err := dst.Subscribe(dstHandler)
		if err != nil {
			t.Fatalf("subscribe dst: %v", err)
		}

		_, err = tap.Subscribe(tapHandler)
		if err != nil {
			t.Fatalf("subscribe tap: %v", err)
		}

		w := NewWiretap("test", src, dst, tap)

		err = w.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		for _, p := range []int{1, 2, 3, 4, 5} {
			err = src.Send(context.Background(), messaging.Message[int]{Payload: p})
			if err != nil {
				t.Fatalf("send %d: %v", p, err)
			}
		}

		if got, want := getDst(), int32(5); got != want {
			t.Fatalf("dst forwarded count: got %d want %d", got, want)
		}

		if got, want := getTap(), int32(5); got != want {
			t.Fatalf("tap forwarded count: got %d want %d", got, want)
		}
	})

	t.Run("preserves payload and headers on both branches", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		tap := messaging.NewPipelineChannel[int]()

		var (
			dstSeen messaging.Message[int]
			tapSeen messaging.Message[int]
		)

		_, err := dst.Subscribe(func(_ context.Context, msg messaging.Message[int]) error {
			dstSeen = msg

			return nil
		})
		if err != nil {
			t.Fatalf("subscribe dst: %v", err)
		}

		_, err = tap.Subscribe(func(_ context.Context, msg messaging.Message[int]) error {
			tapSeen = msg

			return nil
		})
		if err != nil {
			t.Fatalf("subscribe tap: %v", err)
		}

		w := NewWiretap("test", src, dst, tap)

		err = w.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		original := messaging.Message[int]{
			Payload: 42,
			Headers: messaging.Headers{
				MessageID:     "msg-1",
				CorrelationID: "corr-1",
			},
		}

		err = src.Send(context.Background(), original)
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if dstSeen.Payload != 42 || tapSeen.Payload != 42 {
			t.Fatalf("payload not preserved: dst=%d tap=%d", dstSeen.Payload, tapSeen.Payload)
		}

		if dstSeen.Headers.CorrelationID != "corr-1" || tapSeen.Headers.CorrelationID != "corr-1" {
			t.Fatalf("CorrelationID not preserved: dst=%q tap=%q", dstSeen.Headers.CorrelationID, tapSeen.Headers.CorrelationID)
		}
	})
}

func TestWiretap_TapFailure(t *testing.T) {
	t.Parallel()

	t.Run("tap failure does not block primary flow", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		dstHandler, getDst := counter()

		_, err := dst.Subscribe(dstHandler)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		boom := errors.New("tap down")
		tap := &failingChannel[int]{err: boom}

		w := NewWiretap("test", src, dst, tap, WithErrorHandler(messaging.SilentErrorHandler))

		err = w.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("expected nil from src.Send (wiretap swallows tap failure), got %v", err)
		}

		if got, want := getDst(), int32(1); got != want {
			t.Fatalf("dst should receive even when tap fails: got %d want %d", got, want)
		}
	})

	t.Run("tap failure is observable via ErrorHandler", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		boom := errors.New("tap down")
		tap := &failingChannel[int]{err: boom}

		errHandler, getErrs := captureErrors()

		w := NewWiretap("test", src, dst, tap, WithErrorHandler(errHandler))

		err := w.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		errs := getErrs()
		if len(errs) != 1 {
			t.Fatalf("expected 1 captured error, got %d", len(errs))
		}

		if !errors.Is(errs[0], ErrTapSendFailed) {
			t.Fatalf("expected ErrTapSendFailed, got %v", errs[0])
		}

		if !errors.Is(errs[0], ErrWiretapFailed) {
			t.Fatalf("expected ErrWiretapFailed, got %v", errs[0])
		}

		if !errors.Is(errs[0], boom) {
			t.Fatalf("expected wrapped tap error, got %v", errs[0])
		}
	})
}

func TestWiretap_DestinationFailure(t *testing.T) {
	t.Parallel()

	t.Run("dst failure reports ErrForwardFailed", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()

		boom := errors.New("dst down")
		dst := &failingChannel[int]{err: boom}

		tap := messaging.NewPipelineChannel[int]()

		errHandler, getErrs := captureErrors()

		w := NewWiretap("test", src, dst, tap, WithErrorHandler(errHandler))

		err := w.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

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
			t.Fatalf("expected wrapped dst error, got %v", errs[0])
		}
	})

	t.Run("dst failure still attempts the tap send", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()

		dstBoom := errors.New("dst down")
		dst := &failingChannel[int]{err: dstBoom}

		tap := messaging.NewPipelineChannel[int]()
		tapHandler, getTap := counter()

		_, err := tap.Subscribe(tapHandler)
		if err != nil {
			t.Fatalf("subscribe tap: %v", err)
		}

		w := NewWiretap("test", src, dst, tap, WithErrorHandler(messaging.SilentErrorHandler))

		err = w.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got, want := getTap(), int32(1); got != want {
			t.Fatalf("tap should receive even when dst fails: got %d want %d", got, want)
		}
	})

	t.Run("dst failure does not propagate to source caller", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()

		boom := errors.New("dst down")
		dst := &failingChannel[int]{err: boom}

		tap := messaging.NewPipelineChannel[int]()

		w := NewWiretap("test", src, dst, tap, WithErrorHandler(messaging.SilentErrorHandler))

		err := w.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("expected nil from src.Send (wiretap swallows), got %v", err)
		}
	})
}

func TestWiretap_Lifecycle(t *testing.T) {
	t.Parallel()

	t.Run("Start is idempotent", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		tap := messaging.NewPipelineChannel[int]()

		h, get := counter()

		_, err := dst.Subscribe(h)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		w := NewWiretap("test", src, dst, tap)

		err = w.Start(context.Background())
		if err != nil {
			t.Fatalf("first start: %v", err)
		}

		err = w.Start(context.Background())
		if err != nil {
			t.Fatalf("second start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got, want := get(), int32(1); got != want {
			t.Fatalf("dst should receive once (no double subscription), got %d", got)
		}
	})

	t.Run("Stop is idempotent", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		tap := messaging.NewPipelineChannel[int]()

		w := NewWiretap("test", src, dst, tap)

		err := w.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = w.Stop(context.Background())
		if err != nil {
			t.Fatalf("first stop: %v", err)
		}

		err = w.Stop(context.Background())
		if err != nil {
			t.Fatalf("second stop: %v", err)
		}
	})

	t.Run("Done closes after Stop", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		tap := messaging.NewPipelineChannel[int]()

		w := NewWiretap("test", src, dst, tap)

		err := w.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		select {
		case <-w.Done():
			t.Fatal("Done closed before Stop")
		default:
		}

		err = w.Stop(context.Background())
		if err != nil {
			t.Fatalf("stop: %v", err)
		}

		select {
		case <-w.Done():
		default:
			t.Fatal("Done not closed after Stop")
		}
	})

	t.Run("Stop with expired ctx returns ErrShutdownTimeout", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		tap := messaging.NewPipelineChannel[int]()

		w := NewWiretap("test", src, dst, tap)

		err := w.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err = w.Stop(ctx)
		if !errors.Is(err, lifecycle.ErrShutdownTimeout) {
			t.Fatalf("expected ErrShutdownTimeout, got %v", err)
		}
	})

	t.Run("Subscription stops receiving after Stop", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		tap := messaging.NewPipelineChannel[int]()

		h, get := counter()

		_, err := dst.Subscribe(h)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		w := NewWiretap("test", src, dst, tap)

		err = w.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("send pre-stop: %v", err)
		}

		err = w.Stop(context.Background())
		if err != nil {
			t.Fatalf("stop: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 2})
		if err != nil {
			t.Fatalf("send post-stop: %v", err)
		}

		if got, want := get(), int32(1); got != want {
			t.Fatalf("post-stop dst should have received only the pre-stop message, got %d", got)
		}
	})
}

func TestWiretap_Options(t *testing.T) {
	t.Parallel()

	t.Run("WithErrorHandler(nil) is a no-op", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithErrorHandler(nil))
		if opts.errorHandler == nil {
			t.Fatal("expected default error handler preserved on nil arg")
		}
	})

	t.Run("defaults install messaging.DefaultErrorHandler", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()
		if opts.errorHandler == nil {
			t.Fatal("default error handler should be installed")
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
