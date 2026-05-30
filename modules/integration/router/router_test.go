package router

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/core/common/messaging"
)

const keyMissing = "unknown"

// captureErrors returns a thread-safe ErrorHandler that appends every
// reported error to the returned slice (via mutex), and a getter that
// returns a defensive copy.
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

func TestNewRouter(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil component", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		routes := map[string]messaging.Channel[int]{"k": dst}

		decide := func(_ context.Context, _ messaging.Message[int]) (string, error) {
			return "k", nil
		}

		c := NewRouter("test", src, decide, routes)
		if c == nil {
			t.Fatal("expected non-nil component")
		}
	})

	t.Run("carries the given name", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		routes := map[string]messaging.Channel[int]{"k": dst}

		decide := func(_ context.Context, _ messaging.Message[int]) (string, error) {
			return "k", nil
		}

		c := NewRouter("orders-router", src, decide, routes)
		if c.Name() != "orders-router" {
			t.Fatalf("expected name orders-router, got %q", c.Name())
		}
	})
}

func TestRouter_HappyPath(t *testing.T) {
	t.Parallel()

	t.Run("forwards to destination matching key", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		aCh := messaging.NewPipelineChannel[int]()
		bCh := messaging.NewPipelineChannel[int]()

		hA, getA := counter()
		hB, getB := counter()

		_, err := aCh.Subscribe(hA)
		if err != nil {
			t.Fatalf("subscribe a: %v", err)
		}

		_, err = bCh.Subscribe(hB)
		if err != nil {
			t.Fatalf("subscribe b: %v", err)
		}

		decide := func(_ context.Context, msg messaging.Message[int]) (string, error) {
			if msg.Payload%2 == 0 {
				return "even", nil
			}

			return "odd", nil
		}

		r := NewRouter("split", src, decide, map[string]messaging.Channel[int]{
			"even": aCh,
			"odd":  bCh,
		})

		err = r.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		ctx := context.Background()

		for _, p := range []int{2, 4, 1, 6, 3, 5} {
			err = src.Send(ctx, messaging.Message[int]{Payload: p})
			if err != nil {
				t.Fatalf("send %d: %v", p, err)
			}
		}

		if got, want := getA(), int32(3); got != want {
			t.Fatalf("even count: got %d want %d", got, want)
		}

		if got, want := getB(), int32(3); got != want {
			t.Fatalf("odd count: got %d want %d", got, want)
		}
	})
}

func TestRouter_NoRoute(t *testing.T) {
	t.Parallel()

	t.Run("reports ErrNoRoute when key absent and no default", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		routes := map[string]messaging.Channel[int]{"known": dst}

		decide := func(_ context.Context, _ messaging.Message[int]) (string, error) {
			return keyMissing, nil
		}

		errHandler, getErrs := captureErrors()

		r := NewRouter("test", src, decide, routes, WithErrorHandler[int](errHandler))

		err := r.Start(context.Background())
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

		if !errors.Is(errs[0], ErrNoRoute) {
			t.Fatalf("expected ErrNoRoute, got %v", errs[0])
		}

		if !errors.Is(errs[0], ErrRouteFailed) {
			t.Fatalf("expected ErrRouteFailed, got %v", errs[0])
		}
	})

	t.Run("forwards to default channel when key absent and default set", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		defaultCh := messaging.NewPipelineChannel[int]()

		h, get := counter()

		_, err := defaultCh.Subscribe(h)
		if err != nil {
			t.Fatalf("subscribe default: %v", err)
		}

		decide := func(_ context.Context, _ messaging.Message[int]) (string, error) {
			return keyMissing, nil
		}

		errHandler, getErrs := captureErrors()

		r := NewRouter("test", src, decide,
			map[string]messaging.Channel[int]{"known": dst},
			WithDefaultChannel[int](defaultCh),
			WithErrorHandler[int](errHandler))

		err = r.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 42})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got, want := get(), int32(1); got != want {
			t.Fatalf("default channel deliveries: got %d want %d", got, want)
		}

		if errs := getErrs(); len(errs) != 0 {
			t.Fatalf("expected 0 captured errors, got %d: %v", len(errs), errs)
		}
	})
}

func TestRouter_RouteFnError(t *testing.T) {
	t.Parallel()

	t.Run("wraps RouteFn error as ErrRouteFnFailed", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		boom := errors.New("decision boom")

		decide := func(_ context.Context, _ messaging.Message[int]) (string, error) {
			return "", boom
		}

		errHandler, getErrs := captureErrors()

		r := NewRouter("test", src, decide,
			map[string]messaging.Channel[int]{"k": dst},
			WithErrorHandler[int](errHandler))

		err := r.Start(context.Background())
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

		if !errors.Is(errs[0], ErrRouteFnFailed) {
			t.Fatalf("expected ErrRouteFnFailed, got %v", errs[0])
		}

		if !errors.Is(errs[0], boom) {
			t.Fatalf("expected wrapped origin error, got %v", errs[0])
		}
	})

	t.Run("wraps RouteFn panic as ErrRoutePanic", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		decide := func(_ context.Context, _ messaging.Message[int]) (string, error) {
			panic("kaboom")
		}

		errHandler, getErrs := captureErrors()

		r := NewRouter("test", src, decide,
			map[string]messaging.Channel[int]{"k": dst},
			WithErrorHandler[int](errHandler))

		err := r.Start(context.Background())
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

		if !errors.Is(errs[0], ErrRoutePanic) {
			t.Fatalf("expected ErrRoutePanic, got %v", errs[0])
		}
	})
}

func TestRouter_ForwardFailure(t *testing.T) {
	t.Parallel()

	t.Run("reports ErrForwardFailed when destination errs", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()

		boom := errors.New("dest down")
		dst := &failingChannel[int]{err: boom}

		decide := func(_ context.Context, _ messaging.Message[int]) (string, error) {
			return "k", nil
		}

		errHandler, getErrs := captureErrors()

		r := NewRouter("test", src, decide,
			map[string]messaging.Channel[int]{"k": dst},
			WithErrorHandler[int](errHandler))

		err := r.Start(context.Background())
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
			t.Fatalf("expected wrapped destination error, got %v", errs[0])
		}
	})

	t.Run("reports ErrForwardFailed when default channel errs", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		known := messaging.NewPipelineChannel[int]()

		boom := errors.New("default down")
		defaultCh := &failingChannel[int]{err: boom}

		decide := func(_ context.Context, _ messaging.Message[int]) (string, error) {
			return keyMissing, nil
		}

		errHandler, getErrs := captureErrors()

		r := NewRouter("test", src, decide,
			map[string]messaging.Channel[int]{"known": known},
			WithDefaultChannel[int](defaultCh),
			WithErrorHandler[int](errHandler))

		err := r.Start(context.Background())
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
	})
}

func TestRouter_Lifecycle(t *testing.T) {
	t.Parallel()

	t.Run("Start is idempotent", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		h, get := counter()

		_, err := dst.Subscribe(h)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		decide := func(_ context.Context, _ messaging.Message[int]) (string, error) {
			return "k", nil
		}

		r := NewRouter("test", src, decide, map[string]messaging.Channel[int]{"k": dst})

		err = r.Start(context.Background())
		if err != nil {
			t.Fatalf("first start: %v", err)
		}

		err = r.Start(context.Background())
		if err != nil {
			t.Fatalf("second start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got, want := get(), int32(1); got != want {
			t.Fatalf("dest should receive once, got %d", got)
		}
	})

	t.Run("Stop is idempotent", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		decide := func(_ context.Context, _ messaging.Message[int]) (string, error) {
			return "k", nil
		}

		r := NewRouter("test", src, decide, map[string]messaging.Channel[int]{"k": dst})

		err := r.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = r.Stop(context.Background())
		if err != nil {
			t.Fatalf("first stop: %v", err)
		}

		err = r.Stop(context.Background())
		if err != nil {
			t.Fatalf("second stop: %v", err)
		}
	})

	t.Run("Done closes after Stop", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		decide := func(_ context.Context, _ messaging.Message[int]) (string, error) {
			return "k", nil
		}

		r := NewRouter("test", src, decide, map[string]messaging.Channel[int]{"k": dst})

		err := r.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		select {
		case <-r.Done():
			t.Fatal("Done closed before Stop")
		default:
		}

		err = r.Stop(context.Background())
		if err != nil {
			t.Fatalf("stop: %v", err)
		}

		select {
		case <-r.Done():
		default:
			t.Fatal("Done not closed after Stop")
		}
	})

	t.Run("Stop with expired ctx returns ErrShutdownTimeout", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		decide := func(_ context.Context, _ messaging.Message[int]) (string, error) {
			return "k", nil
		}

		r := NewRouter("test", src, decide, map[string]messaging.Channel[int]{"k": dst})

		err := r.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err = r.Stop(ctx)
		if !errors.Is(err, lifecycle.ErrShutdownTimeout) {
			t.Fatalf("expected ErrShutdownTimeout, got %v", err)
		}
	})

	t.Run("Subscription stops receiving after Stop", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		h, get := counter()

		_, err := dst.Subscribe(h)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		decide := func(_ context.Context, _ messaging.Message[int]) (string, error) {
			return "k", nil
		}

		r := NewRouter("test", src, decide, map[string]messaging.Channel[int]{"k": dst})

		err = r.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("send pre-stop: %v", err)
		}

		err = r.Stop(context.Background())
		if err != nil {
			t.Fatalf("stop: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 2})
		if err != nil {
			t.Fatalf("send post-stop: %v", err)
		}

		if got, want := get(), int32(1); got != want {
			t.Fatalf("post-stop dest should have received only the pre-stop message, got %d", got)
		}
	})
}

func TestRouter_Options(t *testing.T) {
	t.Parallel()

	t.Run("WithDefaultChannel(nil) is a no-op", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithDefaultChannel[int](nil))
		if opts.defaultChannel != nil {
			t.Fatal("expected default channel unchanged on nil arg")
		}
	})

	t.Run("WithErrorHandler(nil) is a no-op", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithErrorHandler[int](nil))
		if opts.errorHandler == nil {
			t.Fatal("expected default error handler preserved on nil arg")
		}
	})

	t.Run("defaults install messaging.DefaultErrorHandler", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions[int]()
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
