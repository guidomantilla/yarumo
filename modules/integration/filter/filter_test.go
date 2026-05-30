package filter

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/core/common/messaging"
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

// captureDrops returns a thread-safe DropHandler that counts drops, and
// a getter returning the count.
func captureDrops() (DropHandler, func() int32) {
	var n int32

	handler := func(_ context.Context, _ any) {
		atomic.AddInt32(&n, 1)
	}

	get := func() int32 {
		return atomic.LoadInt32(&n)
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

// alwaysTrue returns a PredicateFn[int] that always passes.
func alwaysTrue(_ context.Context, _ messaging.Message[int]) (bool, error) {
	return true, nil
}

// alwaysFalse returns a PredicateFn[int] that always drops.
func alwaysFalse(_ context.Context, _ messaging.Message[int]) (bool, error) {
	return false, nil
}

func TestNewFilter(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil component", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		c := NewFilter("test", src, dst, alwaysTrue)
		if c == nil {
			t.Fatal("expected non-nil component")
		}
	})

	t.Run("carries the given name", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		c := NewFilter("orders-filter", src, dst, alwaysTrue)
		if c.Name() != "orders-filter" {
			t.Fatalf("expected name orders-filter, got %q", c.Name())
		}
	})
}

func TestFilter_HappyPath(t *testing.T) {
	t.Parallel()

	t.Run("forwards messages where predicate returns true", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		h, get := counter()

		_, err := dst.Subscribe(h)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		// pass-through: only even numbers.
		predicate := func(_ context.Context, msg messaging.Message[int]) (bool, error) {
			return msg.Payload%2 == 0, nil
		}

		f := NewFilter("even-only", src, dst, predicate)

		err = f.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		for _, p := range []int{1, 2, 3, 4, 5, 6} {
			err = src.Send(context.Background(), messaging.Message[int]{Payload: p})
			if err != nil {
				t.Fatalf("send %d: %v", p, err)
			}
		}

		if got, want := get(), int32(3); got != want {
			t.Fatalf("forwarded count: got %d want %d", got, want)
		}
	})
}

func TestFilter_Drop(t *testing.T) {
	t.Parallel()

	t.Run("intentional drop is silent by default", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		dstHandler, getDst := counter()

		_, err := dst.Subscribe(dstHandler)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		errHandler, getErrs := captureErrors()

		f := NewFilter("drop-all", src, dst, alwaysFalse, WithErrorHandler(errHandler))

		err = f.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got := getDst(); got != 0 {
			t.Fatalf("dst should not receive on drop, got %d deliveries", got)
		}

		if errs := getErrs(); len(errs) != 0 {
			t.Fatalf("drops should NOT fire error handler, got %d errors: %v", len(errs), errs)
		}
	})

	t.Run("WithDropHandler observes intentional drops", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		dropHandler, getDrops := captureDrops()

		// keep odd numbers, drop even.
		predicate := func(_ context.Context, msg messaging.Message[int]) (bool, error) {
			return msg.Payload%2 == 1, nil
		}

		f := NewFilter("odd-only", src, dst, predicate, WithDropHandler(dropHandler))

		err := f.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		for _, p := range []int{1, 2, 3, 4} {
			err = src.Send(context.Background(), messaging.Message[int]{Payload: p})
			if err != nil {
				t.Fatalf("send %d: %v", p, err)
			}
		}

		if got, want := getDrops(), int32(2); got != want {
			t.Fatalf("drop count: got %d want %d", got, want)
		}
	})

	t.Run("DropHandler does NOT fire on predicate error", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		dropHandler, getDrops := captureDrops()
		errHandler, getErrs := captureErrors()

		boom := errors.New("predicate broken")

		predicate := func(_ context.Context, _ messaging.Message[int]) (bool, error) {
			return false, boom
		}

		f := NewFilter("erroring", src, dst, predicate,
			WithDropHandler(dropHandler),
			WithErrorHandler(errHandler))

		err := f.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got := getDrops(); got != 0 {
			t.Fatalf("DropHandler must not fire on predicate error, got %d drops", got)
		}

		if errs := getErrs(); len(errs) != 1 {
			t.Fatalf("expected 1 error, got %d", len(errs))
		}
	})
}

func TestFilter_PredicateError(t *testing.T) {
	t.Parallel()

	t.Run("wraps predicate error as ErrPredicateFailed", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		boom := errors.New("predicate boom")

		predicate := func(_ context.Context, _ messaging.Message[int]) (bool, error) {
			return false, boom
		}

		errHandler, getErrs := captureErrors()

		f := NewFilter("test", src, dst, predicate, WithErrorHandler(errHandler))

		err := f.Start(context.Background())
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

		if !errors.Is(errs[0], ErrPredicateFailed) {
			t.Fatalf("expected ErrPredicateFailed, got %v", errs[0])
		}

		if !errors.Is(errs[0], ErrFilterFailed) {
			t.Fatalf("expected ErrFilterFailed, got %v", errs[0])
		}

		if !errors.Is(errs[0], boom) {
			t.Fatalf("expected wrapped origin error, got %v", errs[0])
		}
	})

	t.Run("wraps predicate panic as ErrPredicatePanic", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		predicate := func(_ context.Context, _ messaging.Message[int]) (bool, error) {
			panic("kaboom")
		}

		errHandler, getErrs := captureErrors()

		f := NewFilter("test", src, dst, predicate, WithErrorHandler(errHandler))

		err := f.Start(context.Background())
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

		if !errors.Is(errs[0], ErrPredicatePanic) {
			t.Fatalf("expected ErrPredicatePanic, got %v", errs[0])
		}
	})
}

func TestFilter_ForwardFailure(t *testing.T) {
	t.Parallel()

	t.Run("reports ErrForwardFailed when destination errs", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()

		boom := errors.New("dst down")
		dst := &failingChannel[int]{err: boom}

		errHandler, getErrs := captureErrors()

		f := NewFilter("test", src, dst, alwaysTrue, WithErrorHandler(errHandler))

		err := f.Start(context.Background())
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
}

func TestFilter_Lifecycle(t *testing.T) {
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

		f := NewFilter("test", src, dst, alwaysTrue)

		err = f.Start(context.Background())
		if err != nil {
			t.Fatalf("first start: %v", err)
		}

		err = f.Start(context.Background())
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

		f := NewFilter("test", src, dst, alwaysTrue)

		err := f.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = f.Stop(context.Background())
		if err != nil {
			t.Fatalf("first stop: %v", err)
		}

		err = f.Stop(context.Background())
		if err != nil {
			t.Fatalf("second stop: %v", err)
		}
	})

	t.Run("Done closes after Stop", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		f := NewFilter("test", src, dst, alwaysTrue)

		err := f.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		select {
		case <-f.Done():
			t.Fatal("Done closed before Stop")
		default:
		}

		err = f.Stop(context.Background())
		if err != nil {
			t.Fatalf("stop: %v", err)
		}

		select {
		case <-f.Done():
		default:
			t.Fatal("Done not closed after Stop")
		}
	})

	t.Run("Stop with expired ctx returns ErrShutdownTimeout", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		f := NewFilter("test", src, dst, alwaysTrue)

		err := f.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err = f.Stop(ctx)
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

		f := NewFilter("test", src, dst, alwaysTrue)

		err = f.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("send pre-stop: %v", err)
		}

		err = f.Stop(context.Background())
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

func TestFilter_Options(t *testing.T) {
	t.Parallel()

	t.Run("WithErrorHandler(nil) is a no-op", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithErrorHandler(nil))
		if opts.errorHandler == nil {
			t.Fatal("expected default error handler preserved on nil arg")
		}
	})

	t.Run("WithDropHandler(nil) is a no-op", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithDropHandler(nil))
		if opts.dropHandler != nil {
			t.Fatal("expected default drop handler (nil) preserved on nil arg")
		}
	})

	t.Run("defaults install messaging.DefaultErrorHandler and nil DropHandler", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()
		if opts.errorHandler == nil {
			t.Fatal("default error handler should be installed")
		}

		if opts.dropHandler != nil {
			t.Fatal("default drop handler should be nil (silent drop)")
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
