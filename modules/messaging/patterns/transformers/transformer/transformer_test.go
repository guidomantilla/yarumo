package transformer

import (
	"context"
	"errors"
	"strconv"
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

// counter returns a Handler[string] that increments the returned int32
// once per dispatched message.
func counter() (messaging.Handler[string], func() int32) {
	var n int32

	handler := func(_ context.Context, _ messaging.Message[string]) error {
		atomic.AddInt32(&n, 1)

		return nil
	}

	get := func() int32 {
		return atomic.LoadInt32(&n)
	}

	return handler, get
}

// intToString is a happy-path TransformFn[int, string] that converts the
// payload to its decimal representation, preserving the headers.
func intToString(_ context.Context, msg messaging.Message[int]) (messaging.Message[string], error) {
	return messaging.Message[string]{
		Payload: strconv.Itoa(msg.Payload),
		Headers: msg.Headers,
	}, nil
}

func TestNewTransformer(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil component", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[string]()

		c := NewTransformer("test", src, dst, intToString)
		if c == nil {
			t.Fatal("expected non-nil component")
		}
	})

	t.Run("carries the given name", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[string]()

		c := NewTransformer("orders-translator", src, dst, intToString)
		if c.Name() != "orders-translator" {
			t.Fatalf("expected name orders-translator, got %q", c.Name())
		}
	})
}

func TestTransformer_HappyPath(t *testing.T) {
	t.Parallel()

	t.Run("transforms and forwards every message", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[string]()

		h, get := counter()

		_, err := dst.Subscribe(h)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		x := NewTransformer("test", src, dst, intToString)

		err = x.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		for _, p := range []int{1, 2, 3, 4, 5} {
			err = src.Send(context.Background(), messaging.Message[int]{Payload: p})
			if err != nil {
				t.Fatalf("send %d: %v", p, err)
			}
		}

		if got, want := get(), int32(5); got != want {
			t.Fatalf("forwarded count: got %d want %d", got, want)
		}
	})

	t.Run("preserves headers across the type change", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[string]()

		var seen messaging.Message[string]

		h := func(_ context.Context, msg messaging.Message[string]) error {
			seen = msg

			return nil
		}

		_, err := dst.Subscribe(h)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		x := NewTransformer("test", src, dst, intToString)

		err = x.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		original := messaging.Message[int]{
			Payload: 42,
			Headers: messaging.Headers{
				MessageID:     "msg-1",
				CorrelationID: "corr-1",
				Type:          "test",
			},
		}

		err = src.Send(context.Background(), original)
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if seen.Payload != "42" {
			t.Fatalf("payload not transformed: got %q", seen.Payload)
		}

		if seen.Headers.MessageID != "msg-1" {
			t.Fatalf("MessageID not preserved: got %q", seen.Headers.MessageID)
		}

		if seen.Headers.CorrelationID != "corr-1" {
			t.Fatalf("CorrelationID not preserved: got %q", seen.Headers.CorrelationID)
		}

		if seen.Headers.Type != "test" {
			t.Fatalf("Type not preserved: got %q", seen.Headers.Type)
		}
	})
}

func TestTransformer_TransformError(t *testing.T) {
	t.Parallel()

	t.Run("wraps transform error as ErrTransformFailed", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[string]()

		dstHandler, getDst := counter()

		_, err := dst.Subscribe(dstHandler)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		boom := errors.New("transform boom")

		transform := func(_ context.Context, _ messaging.Message[int]) (messaging.Message[string], error) {
			return messaging.Message[string]{}, boom
		}

		errHandler, getErrs := captureErrors()

		x := NewTransformer("test", src, dst, transform, WithErrorHandler(errHandler))

		err = x.Start(context.Background())
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

		if !errors.Is(errs[0], ErrTransformFailed) {
			t.Fatalf("expected ErrTransformFailed, got %v", errs[0])
		}

		if !errors.Is(errs[0], ErrTransformerFailed) {
			t.Fatalf("expected ErrTransformerFailed, got %v", errs[0])
		}

		if !errors.Is(errs[0], boom) {
			t.Fatalf("expected wrapped origin error, got %v", errs[0])
		}

		if got := getDst(); got != 0 {
			t.Fatalf("dst should not receive on transform error, got %d deliveries", got)
		}
	})

	t.Run("wraps transform panic as ErrTransformerPanic", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[string]()

		dstHandler, getDst := counter()

		_, err := dst.Subscribe(dstHandler)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		transform := func(_ context.Context, _ messaging.Message[int]) (messaging.Message[string], error) {
			panic("kaboom")
		}

		errHandler, getErrs := captureErrors()

		x := NewTransformer("test", src, dst, transform, WithErrorHandler(errHandler))

		err = x.Start(context.Background())
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

		if !errors.Is(errs[0], ErrTransformerPanic) {
			t.Fatalf("expected ErrTransformerPanic, got %v", errs[0])
		}

		if got := getDst(); got != 0 {
			t.Fatalf("dst should not receive on transform panic, got %d deliveries", got)
		}
	})
}

func TestTransformer_ForwardFailure(t *testing.T) {
	t.Parallel()

	t.Run("reports ErrForwardFailed when destination errs", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()

		boom := errors.New("dst down")
		dst := &failingChannel[string]{err: boom}

		errHandler, getErrs := captureErrors()

		x := NewTransformer("test", src, dst, intToString, WithErrorHandler(errHandler))

		err := x.Start(context.Background())
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

		if !errors.Is(errs[0], ErrTransformerFailed) {
			t.Fatalf("expected ErrTransformerFailed, got %v", errs[0])
		}

		if !errors.Is(errs[0], boom) {
			t.Fatalf("expected wrapped destination error, got %v", errs[0])
		}
	})

	t.Run("forward failure does not propagate to source caller", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()

		boom := errors.New("dst down")
		dst := &failingChannel[string]{err: boom}

		x := NewTransformer("test", src, dst, intToString, WithErrorHandler(messaging.SilentErrorHandler))

		err := x.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("expected nil from src.Send (transformer swallows), got %v", err)
		}
	})
}

func TestTransformer_Lifecycle(t *testing.T) {
	t.Parallel()

	t.Run("Start is idempotent", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[string]()

		h, get := counter()

		_, err := dst.Subscribe(h)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		x := NewTransformer("test", src, dst, intToString)

		err = x.Start(context.Background())
		if err != nil {
			t.Fatalf("first start: %v", err)
		}

		err = x.Start(context.Background())
		if err != nil {
			t.Fatalf("second start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got, want := get(), int32(1); got != want {
			t.Fatalf("dest should receive once (no double subscription), got %d", got)
		}
	})

	t.Run("Stop is idempotent", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[string]()

		x := NewTransformer("test", src, dst, intToString)

		err := x.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = x.Stop(context.Background())
		if err != nil {
			t.Fatalf("first stop: %v", err)
		}

		err = x.Stop(context.Background())
		if err != nil {
			t.Fatalf("second stop: %v", err)
		}
	})

	t.Run("Done closes after Stop", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[string]()

		x := NewTransformer("test", src, dst, intToString)

		err := x.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		select {
		case <-x.Done():
			t.Fatal("Done closed before Stop")
		default:
		}

		err = x.Stop(context.Background())
		if err != nil {
			t.Fatalf("stop: %v", err)
		}

		select {
		case <-x.Done():
		default:
			t.Fatal("Done not closed after Stop")
		}
	})

	t.Run("Stop with expired ctx returns ErrShutdownTimeout", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[string]()

		x := NewTransformer("test", src, dst, intToString)

		err := x.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err = x.Stop(ctx)
		if !errors.Is(err, lifecycle.ErrShutdownTimeout) {
			t.Fatalf("expected ErrShutdownTimeout, got %v", err)
		}
	})

	t.Run("Subscription stops receiving after Stop", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[string]()

		h, get := counter()

		_, err := dst.Subscribe(h)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		x := NewTransformer("test", src, dst, intToString)

		err = x.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("send pre-stop: %v", err)
		}

		err = x.Stop(context.Background())
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

func TestTransformer_Options(t *testing.T) {
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
