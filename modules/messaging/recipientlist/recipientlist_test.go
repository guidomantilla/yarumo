package recipientlist

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/messaging"
)

const keyMissing = "unknown"

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

func TestNewRecipientList(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil component", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		routes := map[string]messaging.Channel[int]{"k": dst}

		selector := func(_ context.Context, _ messaging.Message[int]) ([]string, error) {
			return []string{"k"}, nil
		}

		c := NewRecipientList("test", src, selector, routes)
		if c == nil {
			t.Fatal("expected non-nil component")
		}
	})

	t.Run("carries the given name", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		routes := map[string]messaging.Channel[int]{"k": dst}

		selector := func(_ context.Context, _ messaging.Message[int]) ([]string, error) {
			return []string{"k"}, nil
		}

		c := NewRecipientList("orders-recipients", src, selector, routes)
		if c.Name() != "orders-recipients" {
			t.Fatalf("expected name orders-recipients, got %q", c.Name())
		}
	})
}

func TestRecipientList_HappyPath(t *testing.T) {
	t.Parallel()

	t.Run("fans out to every recipient matching selector keys", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		aCh := messaging.NewPipelineChannel[int]()
		bCh := messaging.NewPipelineChannel[int]()
		cCh := messaging.NewPipelineChannel[int]()

		hA, getA := counter()
		hB, getB := counter()
		hC, getC := counter()

		_, err := aCh.Subscribe(hA)
		if err != nil {
			t.Fatalf("subscribe a: %v", err)
		}

		_, err = bCh.Subscribe(hB)
		if err != nil {
			t.Fatalf("subscribe b: %v", err)
		}

		_, err = cCh.Subscribe(hC)
		if err != nil {
			t.Fatalf("subscribe c: %v", err)
		}

		// selector: every payload is sent to A and B; C only when even.
		selector := func(_ context.Context, msg messaging.Message[int]) ([]string, error) {
			keys := []string{"a", "b"}
			if msg.Payload%2 == 0 {
				keys = append(keys, "c")
			}

			return keys, nil
		}

		r := NewRecipientList("split", src, selector, map[string]messaging.Channel[int]{
			"a": aCh,
			"b": bCh,
			"c": cCh,
		})

		err = r.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		// 4 messages → 4 to A + 4 to B + 2 to C (even).
		for _, p := range []int{1, 2, 3, 4} {
			err = src.Send(context.Background(), messaging.Message[int]{Payload: p})
			if err != nil {
				t.Fatalf("send %d: %v", p, err)
			}
		}

		if got, want := getA(), int32(4); got != want {
			t.Fatalf("a count: got %d want %d", got, want)
		}

		if got, want := getB(), int32(4); got != want {
			t.Fatalf("b count: got %d want %d", got, want)
		}

		if got, want := getC(), int32(2); got != want {
			t.Fatalf("c count: got %d want %d", got, want)
		}
	})
}

func TestRecipientList_PartialFailure(t *testing.T) {
	t.Parallel()

	t.Run("missing key reports ErrNoRoute but valid recipient still delivered", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		valid := messaging.NewPipelineChannel[int]()

		hValid, getValid := counter()

		_, err := valid.Subscribe(hValid)
		if err != nil {
			t.Fatalf("subscribe valid: %v", err)
		}

		selector := func(_ context.Context, _ messaging.Message[int]) ([]string, error) {
			return []string{"valid", keyMissing}, nil
		}

		errHandler, getErrs := captureErrors()

		r := NewRecipientList("test", src, selector,
			map[string]messaging.Channel[int]{"valid": valid},
			WithErrorHandler(errHandler))

		err = r.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 42})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got, want := getValid(), int32(1); got != want {
			t.Fatalf("valid recipient: got %d want %d", got, want)
		}

		errs := getErrs()
		if len(errs) != 1 {
			t.Fatalf("expected 1 captured error (per missing key), got %d", len(errs))
		}

		if !errors.Is(errs[0], ErrNoRoute) {
			t.Fatalf("expected ErrNoRoute, got %v", errs[0])
		}

		if !errors.Is(errs[0], ErrRecipientListFailed) {
			t.Fatalf("expected ErrRecipientListFailed, got %v", errs[0])
		}
	})

	t.Run("forward failure for one recipient does not abort others", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		good := messaging.NewPipelineChannel[int]()

		hGood, getGood := counter()

		_, err := good.Subscribe(hGood)
		if err != nil {
			t.Fatalf("subscribe good: %v", err)
		}

		boom := errors.New("bad down")
		bad := &failingChannel[int]{err: boom}

		selector := func(_ context.Context, _ messaging.Message[int]) ([]string, error) {
			return []string{"good", "bad"}, nil
		}

		errHandler, getErrs := captureErrors()

		r := NewRecipientList("test", src, selector,
			map[string]messaging.Channel[int]{"good": good, "bad": bad},
			WithErrorHandler(errHandler))

		err = r.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got, want := getGood(), int32(1); got != want {
			t.Fatalf("good recipient: got %d want %d", got, want)
		}

		errs := getErrs()
		if len(errs) != 1 {
			t.Fatalf("expected 1 captured error for bad recipient, got %d", len(errs))
		}

		if !errors.Is(errs[0], ErrForwardFailed) {
			t.Fatalf("expected ErrForwardFailed, got %v", errs[0])
		}

		if !errors.Is(errs[0], boom) {
			t.Fatalf("expected wrapped destination error, got %v", errs[0])
		}
	})
}

func TestRecipientList_EmptySelection(t *testing.T) {
	t.Parallel()

	t.Run("empty selector result fires DropHandler and no recipients are called", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		hDst, getDst := counter()

		_, err := dst.Subscribe(hDst)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		selector := func(_ context.Context, _ messaging.Message[int]) ([]string, error) {
			return []string{}, nil
		}

		dropHandler, getDrops := captureDrops()
		errHandler, getErrs := captureErrors()

		r := NewRecipientList("test", src, selector,
			map[string]messaging.Channel[int]{"k": dst},
			WithDropHandler(dropHandler),
			WithErrorHandler(errHandler))

		err = r.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got := getDst(); got != 0 {
			t.Fatalf("no recipient should fire on empty selection, got %d deliveries", got)
		}

		if got, want := getDrops(), int32(1); got != want {
			t.Fatalf("drop count: got %d want %d", got, want)
		}

		if errs := getErrs(); len(errs) != 0 {
			t.Fatalf("empty selection should NOT fire error handler, got %d errors: %v", len(errs), errs)
		}
	})

	t.Run("nil selector result is treated as empty (drop)", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		selector := func(_ context.Context, _ messaging.Message[int]) ([]string, error) {
			return nil, nil
		}

		dropHandler, getDrops := captureDrops()

		r := NewRecipientList("test", src, selector,
			map[string]messaging.Channel[int]{"k": dst},
			WithDropHandler(dropHandler))

		err := r.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got, want := getDrops(), int32(1); got != want {
			t.Fatalf("drop count: got %d want %d", got, want)
		}
	})
}

func TestRecipientList_SelectorError(t *testing.T) {
	t.Parallel()

	t.Run("wraps selector error as ErrSelectorFnFailed", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		boom := errors.New("selector boom")

		selector := func(_ context.Context, _ messaging.Message[int]) ([]string, error) {
			return nil, boom
		}

		errHandler, getErrs := captureErrors()

		r := NewRecipientList("test", src, selector,
			map[string]messaging.Channel[int]{"k": dst},
			WithErrorHandler(errHandler))

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

		if !errors.Is(errs[0], ErrSelectorFnFailed) {
			t.Fatalf("expected ErrSelectorFnFailed, got %v", errs[0])
		}

		if !errors.Is(errs[0], boom) {
			t.Fatalf("expected wrapped origin error, got %v", errs[0])
		}
	})

	t.Run("wraps selector panic as ErrSelectorPanic", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		selector := func(_ context.Context, _ messaging.Message[int]) ([]string, error) {
			panic("kaboom")
		}

		errHandler, getErrs := captureErrors()

		r := NewRecipientList("test", src, selector,
			map[string]messaging.Channel[int]{"k": dst},
			WithErrorHandler(errHandler))

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

		if !errors.Is(errs[0], ErrSelectorPanic) {
			t.Fatalf("expected ErrSelectorPanic, got %v", errs[0])
		}
	})
}

func TestRecipientList_Lifecycle(t *testing.T) {
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

		selector := func(_ context.Context, _ messaging.Message[int]) ([]string, error) {
			return []string{"k"}, nil
		}

		r := NewRecipientList("test", src, selector, map[string]messaging.Channel[int]{"k": dst})

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

		selector := func(_ context.Context, _ messaging.Message[int]) ([]string, error) {
			return []string{"k"}, nil
		}

		r := NewRecipientList("test", src, selector, map[string]messaging.Channel[int]{"k": dst})

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

		selector := func(_ context.Context, _ messaging.Message[int]) ([]string, error) {
			return []string{"k"}, nil
		}

		r := NewRecipientList("test", src, selector, map[string]messaging.Channel[int]{"k": dst})

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

		selector := func(_ context.Context, _ messaging.Message[int]) ([]string, error) {
			return []string{"k"}, nil
		}

		r := NewRecipientList("test", src, selector, map[string]messaging.Channel[int]{"k": dst})

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
}

func TestRecipientList_Options(t *testing.T) {
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
