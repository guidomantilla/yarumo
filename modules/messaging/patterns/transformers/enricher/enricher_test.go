package enricher

import (
	"context"
	"errors"
	"maps"
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

func TestNewEnricher(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil component", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		enrich := func(_ context.Context, msg messaging.Message[int]) (messaging.Message[int], error) {
			return msg, nil
		}

		c := NewEnricher("test", src, dst, enrich)
		if c == nil {
			t.Fatal("expected non-nil component")
		}
	})

	t.Run("carries the given name", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		enrich := func(_ context.Context, msg messaging.Message[int]) (messaging.Message[int], error) {
			return msg, nil
		}

		c := NewEnricher("audit-enricher", src, dst, enrich)
		if c.Name() != "audit-enricher" {
			t.Fatalf("expected name audit-enricher, got %q", c.Name())
		}
	})
}

func TestEnricher_HappyPath(t *testing.T) {
	t.Parallel()

	t.Run("EnrichFn adds a Custom header observed downstream", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		var seen messaging.Message[int]

		_, err := dst.Subscribe(func(_ context.Context, msg messaging.Message[int]) error {
			seen = msg

			return nil
		})
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		enrich := func(_ context.Context, msg messaging.Message[int]) (messaging.Message[int], error) {
			custom := map[string]any{"audit-tag": "gateway-A"}
			maps.Copy(custom, msg.Headers.Custom)

			out := msg
			out.Headers.Custom = custom

			return out, nil
		}

		e := NewEnricher("test", src, dst, enrich)

		err = e.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{
			Payload: 1,
			Headers: messaging.Headers{MessageID: "m-1"},
		})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if v, ok := seen.Headers.Custom["audit-tag"]; !ok || v != "gateway-A" {
			t.Fatalf("audit-tag not enriched, got ok=%v v=%v", ok, v)
		}

		if seen.Headers.MessageID != "m-1" {
			t.Fatalf("MessageID should pass through, got %q", seen.Headers.MessageID)
		}

		if seen.Payload != 1 {
			t.Fatalf("Payload should pass through, got %d", seen.Payload)
		}
	})

	t.Run("EnrichFn modifies Payload observed downstream", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		var seen messaging.Message[int]

		_, err := dst.Subscribe(func(_ context.Context, msg messaging.Message[int]) error {
			seen = msg

			return nil
		})
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		// double the payload.
		enrich := func(_ context.Context, msg messaging.Message[int]) (messaging.Message[int], error) {
			out := msg
			out.Payload = msg.Payload * 2

			return out, nil
		}

		e := NewEnricher("test", src, dst, enrich)

		err = e.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 21})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if seen.Payload != 42 {
			t.Fatalf("Payload should be doubled, got %d", seen.Payload)
		}
	})

	t.Run("EnrichFn sets a header field observed downstream", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		var seen messaging.Message[int]

		_, err := dst.Subscribe(func(_ context.Context, msg messaging.Message[int]) error {
			seen = msg

			return nil
		})
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		enrich := func(_ context.Context, msg messaging.Message[int]) (messaging.Message[int], error) {
			out := msg
			out.Headers.Source = "svc-gateway"

			return out, nil
		}

		e := NewEnricher("test", src, dst, enrich)

		err = e.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if seen.Headers.Source != "svc-gateway" {
			t.Fatalf("Source not enriched, got %q", seen.Headers.Source)
		}
	})
}

func TestEnricher_EnrichError(t *testing.T) {
	t.Parallel()

	t.Run("wraps EnrichFn error as ErrEnrichFnFailed and skips forward", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		h, getDst := counter()

		_, err := dst.Subscribe(h)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		boom := errors.New("enrich boom")

		enrich := func(_ context.Context, _ messaging.Message[int]) (messaging.Message[int], error) {
			return messaging.Message[int]{}, boom
		}

		errHandler, getErrs := captureErrors()

		e := NewEnricher("test", src, dst, enrich, WithErrorHandler(errHandler))

		err = e.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got := getDst(); got != 0 {
			t.Fatalf("dst should NOT receive on enrich error, got %d deliveries", got)
		}

		errs := getErrs()
		if len(errs) != 1 {
			t.Fatalf("expected 1 captured error, got %d", len(errs))
		}

		if !errors.Is(errs[0], ErrEnrichFnFailed) {
			t.Fatalf("expected ErrEnrichFnFailed, got %v", errs[0])
		}

		if !errors.Is(errs[0], ErrEnricherFailed) {
			t.Fatalf("expected ErrEnricherFailed, got %v", errs[0])
		}

		if !errors.Is(errs[0], boom) {
			t.Fatalf("expected wrapped origin error, got %v", errs[0])
		}
	})

	t.Run("wraps EnrichFn panic as ErrEnrichPanic and skips forward", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		h, getDst := counter()

		_, err := dst.Subscribe(h)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		enrich := func(_ context.Context, _ messaging.Message[int]) (messaging.Message[int], error) {
			panic("kaboom")
		}

		errHandler, getErrs := captureErrors()

		e := NewEnricher("test", src, dst, enrich, WithErrorHandler(errHandler))

		err = e.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got := getDst(); got != 0 {
			t.Fatalf("dst should NOT receive on enrich panic, got %d deliveries", got)
		}

		errs := getErrs()
		if len(errs) != 1 {
			t.Fatalf("expected 1 captured error, got %d", len(errs))
		}

		if !errors.Is(errs[0], ErrEnrichPanic) {
			t.Fatalf("expected ErrEnrichPanic, got %v", errs[0])
		}
	})
}

func TestEnricher_ForwardFailure(t *testing.T) {
	t.Parallel()

	t.Run("reports ErrForwardFailed when destination errs", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()

		boom := errors.New("dst down")
		dst := &failingChannel[int]{err: boom}

		enrich := func(_ context.Context, msg messaging.Message[int]) (messaging.Message[int], error) {
			return msg, nil
		}

		errHandler, getErrs := captureErrors()

		e := NewEnricher("test", src, dst, enrich, WithErrorHandler(errHandler))

		err := e.Start(context.Background())
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

	t.Run("forward failure does not propagate to source caller", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()

		boom := errors.New("dst down")
		dst := &failingChannel[int]{err: boom}

		enrich := func(_ context.Context, msg messaging.Message[int]) (messaging.Message[int], error) {
			return msg, nil
		}

		e := NewEnricher("test", src, dst, enrich,
			WithErrorHandler(messaging.SilentErrorHandler))

		err := e.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("expected nil from src.Send (enricher swallows), got %v", err)
		}
	})
}

func TestEnricher_Lifecycle(t *testing.T) {
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

		enrich := func(_ context.Context, msg messaging.Message[int]) (messaging.Message[int], error) {
			return msg, nil
		}

		e := NewEnricher("test", src, dst, enrich)

		err = e.Start(context.Background())
		if err != nil {
			t.Fatalf("first start: %v", err)
		}

		err = e.Start(context.Background())
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

		enrich := func(_ context.Context, msg messaging.Message[int]) (messaging.Message[int], error) {
			return msg, nil
		}

		e := NewEnricher("test", src, dst, enrich)

		err := e.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = e.Stop(context.Background())
		if err != nil {
			t.Fatalf("first stop: %v", err)
		}

		err = e.Stop(context.Background())
		if err != nil {
			t.Fatalf("second stop: %v", err)
		}
	})

	t.Run("Done closes after Stop", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		enrich := func(_ context.Context, msg messaging.Message[int]) (messaging.Message[int], error) {
			return msg, nil
		}

		e := NewEnricher("test", src, dst, enrich)

		err := e.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		select {
		case <-e.Done():
			t.Fatal("Done closed before Stop")
		default:
		}

		err = e.Stop(context.Background())
		if err != nil {
			t.Fatalf("stop: %v", err)
		}

		select {
		case <-e.Done():
		default:
			t.Fatal("Done not closed after Stop")
		}
	})

	t.Run("Stop with expired ctx returns ErrShutdownTimeout", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		enrich := func(_ context.Context, msg messaging.Message[int]) (messaging.Message[int], error) {
			return msg, nil
		}

		e := NewEnricher("test", src, dst, enrich)

		err := e.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err = e.Stop(ctx)
		if !errors.Is(err, lifecycle.ErrShutdownTimeout) {
			t.Fatalf("expected ErrShutdownTimeout, got %v", err)
		}
	})
}

func TestEnricher_Options(t *testing.T) {
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
