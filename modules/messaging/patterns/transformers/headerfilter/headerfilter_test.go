package headerfilter

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

const testMessageID = "m-1"

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

func TestNewHeaderFilter(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil component", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		c := NewHeaderFilter("test", src, dst)
		if c == nil {
			t.Fatal("expected non-nil component")
		}
	})

	t.Run("carries the given name", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		c := NewHeaderFilter("strip-reply", src, dst)
		if c.Name() != "strip-reply" {
			t.Fatalf("expected name strip-reply, got %q", c.Name())
		}
	})
}

func TestHeaderFilter_ClearKnownFields(t *testing.T) {
	t.Parallel()

	t.Run("WithClearHeader(ReplyTo) zeroes ReplyTo on forwarded msg", func(t *testing.T) {
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

		f := NewHeaderFilter("test", src, dst, WithClearHeader("ReplyTo"))

		err = f.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		original := messaging.Message[int]{
			Payload: 42,
			Headers: messaging.Headers{
				MessageID: testMessageID,
				ReplyTo:   "secret-queue",
				Type:      "test",
			},
		}

		err = src.Send(context.Background(), original)
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if seen.Headers.ReplyTo != "" {
			t.Fatalf("ReplyTo should be cleared, got %q", seen.Headers.ReplyTo)
		}

		if seen.Headers.MessageID != testMessageID {
			t.Fatalf("MessageID should be preserved, got %q", seen.Headers.MessageID)
		}

		if seen.Headers.Type != "test" {
			t.Fatalf("Type should be preserved, got %q", seen.Headers.Type)
		}

		if seen.Payload != 42 {
			t.Fatalf("payload should be preserved, got %d", seen.Payload)
		}
	})

	t.Run("multiple WithClearHeader calls clear multiple fields", func(t *testing.T) {
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

		f := NewHeaderFilter("test", src, dst,
			WithClearHeader("ReplyTo"),
			WithClearHeader("Source"),
			WithClearHeader("CorrelationID"),
		)

		err = f.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		original := messaging.Message[int]{
			Payload: 1,
			Headers: messaging.Headers{
				MessageID:     testMessageID,
				CorrelationID: "c-1",
				ReplyTo:       "queue",
				Source:        "svc-a",
				Type:          "keep",
			},
		}

		err = src.Send(context.Background(), original)
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if seen.Headers.ReplyTo != "" {
			t.Fatalf("ReplyTo not cleared, got %q", seen.Headers.ReplyTo)
		}

		if seen.Headers.Source != "" {
			t.Fatalf("Source not cleared, got %q", seen.Headers.Source)
		}

		if seen.Headers.CorrelationID != "" {
			t.Fatalf("CorrelationID not cleared, got %q", seen.Headers.CorrelationID)
		}

		if seen.Headers.MessageID != testMessageID {
			t.Fatalf("MessageID should be preserved, got %q", seen.Headers.MessageID)
		}

		if seen.Headers.Type != "keep" {
			t.Fatalf("Type should be preserved, got %q", seen.Headers.Type)
		}
	})

	t.Run("WithHeadersToClear variadic equivalent to many WithClearHeader", func(t *testing.T) {
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

		f := NewHeaderFilter("test", src, dst,
			WithHeadersToClear("Priority", "ExpirationTime", "SequenceNumber", "SequenceSize"),
		)

		err = f.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		original := messaging.Message[int]{
			Payload: 7,
			Headers: messaging.Headers{
				Priority:       9,
				ExpirationTime: time.Now().Add(time.Hour),
				SequenceNumber: 3,
				SequenceSize:   10,
				Type:           "keep",
			},
		}

		err = src.Send(context.Background(), original)
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if seen.Headers.Priority != 0 {
			t.Fatalf("Priority not cleared, got %d", seen.Headers.Priority)
		}

		if !seen.Headers.ExpirationTime.IsZero() {
			t.Fatalf("ExpirationTime not cleared, got %v", seen.Headers.ExpirationTime)
		}

		if seen.Headers.SequenceNumber != 0 {
			t.Fatalf("SequenceNumber not cleared, got %d", seen.Headers.SequenceNumber)
		}

		if seen.Headers.SequenceSize != 0 {
			t.Fatalf("SequenceSize not cleared, got %d", seen.Headers.SequenceSize)
		}

		if seen.Headers.Type != "keep" {
			t.Fatalf("Type should be preserved, got %q", seen.Headers.Type)
		}
	})
}

func TestHeaderFilter_ClearCustomKey(t *testing.T) {
	t.Parallel()

	t.Run("unknown name deletes from Custom map", func(t *testing.T) {
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

		f := NewHeaderFilter("test", src, dst,
			WithClearHeader("x-debug-secret"),
		)

		err = f.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		original := messaging.Message[int]{
			Payload: 1,
			Headers: messaging.Headers{
				Custom: map[string]any{
					"x-debug-secret": "leakable",
					"x-trace-id":     "keep-me",
				},
			},
		}

		err = src.Send(context.Background(), original)
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if _, ok := seen.Headers.Custom["x-debug-secret"]; ok {
			t.Fatal("x-debug-secret should be removed from Custom")
		}

		if v, ok := seen.Headers.Custom["x-trace-id"]; !ok || v != "keep-me" {
			t.Fatalf("x-trace-id should be preserved, got ok=%v v=%v", ok, v)
		}
	})

	t.Run("source message Custom map is not mutated", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		_, err := dst.Subscribe(func(_ context.Context, _ messaging.Message[int]) error {
			return nil
		})
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		f := NewHeaderFilter("test", src, dst, WithClearHeader("secret"))

		err = f.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		original := messaging.Message[int]{
			Payload: 1,
			Headers: messaging.Headers{
				Custom: map[string]any{
					"secret": "value",
					"other":  "ok",
				},
			},
		}

		err = src.Send(context.Background(), original)
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if _, ok := original.Headers.Custom["secret"]; !ok {
			t.Fatal("source Custom['secret'] should not be mutated by headerfilter")
		}

		if v, ok := original.Headers.Custom["other"]; !ok || v != "ok" {
			t.Fatalf("source Custom['other'] should not be mutated, got ok=%v v=%v", ok, v)
		}
	})

	t.Run("absent Custom key is a no-op (no clone)", func(t *testing.T) {
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

		f := NewHeaderFilter("test", src, dst, WithClearHeader("absent-key"))

		err = f.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		original := messaging.Message[int]{
			Payload: 1,
			Headers: messaging.Headers{
				Custom: map[string]any{"present": "ok"},
			},
		}

		err = src.Send(context.Background(), original)
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if v, ok := seen.Headers.Custom["present"]; !ok || v != "ok" {
			t.Fatalf("present key should be preserved, got ok=%v v=%v", ok, v)
		}
	})
}

func TestHeaderFilter_Defaults(t *testing.T) {
	t.Parallel()

	t.Run("empty clear list forwards messages unchanged", func(t *testing.T) {
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

		f := NewHeaderFilter("test", src, dst)

		err = f.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		original := messaging.Message[int]{
			Payload: 99,
			Headers: messaging.Headers{
				MessageID: testMessageID,
				ReplyTo:   "rt",
				Type:      "t",
			},
		}

		err = src.Send(context.Background(), original)
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if seen.Headers.MessageID != testMessageID {
			t.Fatalf("MessageID not preserved, got %q", seen.Headers.MessageID)
		}

		if seen.Headers.ReplyTo != "rt" {
			t.Fatalf("ReplyTo not preserved (no clear configured), got %q", seen.Headers.ReplyTo)
		}

		if seen.Payload != 99 {
			t.Fatalf("Payload not preserved, got %d", seen.Payload)
		}
	})

	t.Run("WithClearHeader with empty name is a no-op", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithClearHeader(""))
		if len(opts.headersToClear) != 0 {
			t.Fatalf("empty name should not be added, got %v", opts.headersToClear)
		}
	})

	t.Run("WithClearHeader deduplicates", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(
			WithClearHeader("ReplyTo"),
			WithClearHeader("ReplyTo"),
			WithClearHeader("Source"),
		)

		if len(opts.headersToClear) != 2 {
			t.Fatalf("expected 2 unique names, got %v", opts.headersToClear)
		}
	})
}

func TestHeaderFilter_ForwardFailure(t *testing.T) {
	t.Parallel()

	t.Run("reports ErrForwardFailed when destination errs", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()

		boom := errors.New("dst down")
		dst := &failingChannel[int]{err: boom}

		errHandler, getErrs := captureErrors()

		f := NewHeaderFilter("test", src, dst,
			WithClearHeader("ReplyTo"),
			WithErrorHandler(errHandler))

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

		if !errors.Is(errs[0], ErrHeaderFilterFailed) {
			t.Fatalf("expected ErrHeaderFilterFailed, got %v", errs[0])
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

		f := NewHeaderFilter("test", src, dst,
			WithClearHeader("ReplyTo"),
			WithErrorHandler(messaging.SilentErrorHandler))

		err := f.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("expected nil from src.Send (filter swallows), got %v", err)
		}
	})
}

func TestHeaderFilter_Lifecycle(t *testing.T) {
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

		f := NewHeaderFilter("test", src, dst)

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

		f := NewHeaderFilter("test", src, dst)

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

		f := NewHeaderFilter("test", src, dst)

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

		f := NewHeaderFilter("test", src, dst)

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
}

func TestHeaderFilter_Options(t *testing.T) {
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
