package history

import (
	"context"
	"errors"
	"sync"
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

// captureMessages returns a Handler[int] that records every received
// message and a getter that returns a defensive copy.
func captureMessages() (messaging.Handler[int], func() []messaging.Message[int]) {
	var mu sync.Mutex

	captured := []messaging.Message[int]{}

	handler := func(_ context.Context, msg messaging.Message[int]) error {
		mu.Lock()
		defer mu.Unlock()

		captured = append(captured, msg)

		return nil
	}

	get := func() []messaging.Message[int] {
		mu.Lock()
		defer mu.Unlock()

		out := make([]messaging.Message[int], len(captured))
		copy(out, captured)

		return out
	}

	return handler, get
}

func TestNewHistory(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil component", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		c := NewHistory("test", src, dst)
		if c == nil {
			t.Fatal("expected non-nil component")
		}
	})

	t.Run("carries the given name", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		c := NewHistory("orders-history", src, dst)
		if c.Name() != "orders-history" {
			t.Fatalf("expected name orders-history, got %q", c.Name())
		}
	})
}

func TestHistory_HappyPath(t *testing.T) {
	t.Parallel()

	t.Run("single endpoint seeds Custom[History]", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		dstHandler, getMsgs := captureMessages()

		_, err := dst.Subscribe(dstHandler)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		h := NewHistory("stage-a", src, dst)

		err = h.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 42})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		msgs := getMsgs()
		if len(msgs) != 1 {
			t.Fatalf("expected 1 message, got %d", len(msgs))
		}

		trail, ok := msgs[0].Headers.Custom[DefaultHistoryKey].([]string)
		if !ok {
			t.Fatalf("expected Custom[%q] to be []string, got %T", DefaultHistoryKey, msgs[0].Headers.Custom[DefaultHistoryKey])
		}

		if len(trail) != 1 || trail[0] != "stage-a" {
			t.Fatalf("expected trail [stage-a], got %v", trail)
		}
	})

	t.Run("payload preserved end-to-end", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		dstHandler, getMsgs := captureMessages()

		_, err := dst.Subscribe(dstHandler)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		h := NewHistory("stamp", src, dst)

		err = h.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		original := messaging.Message[int]{
			Payload: 7,
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

		msgs := getMsgs()
		if len(msgs) != 1 {
			t.Fatalf("expected 1 message, got %d", len(msgs))
		}

		if msgs[0].Payload != 7 {
			t.Fatalf("payload not preserved: got %d", msgs[0].Payload)
		}

		if msgs[0].Headers.MessageID != "msg-1" {
			t.Fatalf("MessageID not preserved: got %q", msgs[0].Headers.MessageID)
		}

		if msgs[0].Headers.CorrelationID != "corr-1" {
			t.Fatalf("CorrelationID not preserved: got %q", msgs[0].Headers.CorrelationID)
		}

		if msgs[0].Headers.Type != "test" {
			t.Fatalf("Type not preserved: got %q", msgs[0].Headers.Type)
		}
	})
}

func TestHistory_Chain(t *testing.T) {
	t.Parallel()

	t.Run("chained endpoints accumulate trail in order", func(t *testing.T) {
		t.Parallel()

		ch1 := messaging.NewPipelineChannel[int]()
		ch2 := messaging.NewPipelineChannel[int]()
		ch3 := messaging.NewPipelineChannel[int]()
		ch4 := messaging.NewPipelineChannel[int]()

		dstHandler, getMsgs := captureMessages()

		_, err := ch4.Subscribe(dstHandler)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		ha := NewHistory("a", ch1, ch2)
		hb := NewHistory("b", ch2, ch3)
		hc := NewHistory("c", ch3, ch4)

		for _, c := range []History[int]{ha, hb, hc} {
			err = c.Start(context.Background())
			if err != nil {
				t.Fatalf("start: %v", err)
			}
		}

		err = ch1.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		msgs := getMsgs()
		if len(msgs) != 1 {
			t.Fatalf("expected 1 message, got %d", len(msgs))
		}

		trail, ok := msgs[0].Headers.Custom[DefaultHistoryKey].([]string)
		if !ok {
			t.Fatalf("expected []string trail, got %T", msgs[0].Headers.Custom[DefaultHistoryKey])
		}

		want := []string{"a", "b", "c"}
		if len(trail) != len(want) {
			t.Fatalf("trail length: got %d want %d (trail=%v)", len(trail), len(want), trail)
		}

		for i, w := range want {
			if trail[i] != w {
				t.Fatalf("trail[%d]: got %q want %q", i, trail[i], w)
			}
		}
	})
}

func TestHistory_PreservesExistingTrail(t *testing.T) {
	t.Parallel()

	t.Run("appends to caller-provided Custom[History]", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		dstHandler, getMsgs := captureMessages()

		_, err := dst.Subscribe(dstHandler)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		h := NewHistory("stage-z", src, dst)

		err = h.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		seed := messaging.Message[int]{
			Payload: 1,
			Headers: messaging.Headers{
				Custom: map[string]any{
					DefaultHistoryKey: []string{"upstream-1", "upstream-2"},
					"other-key":       "kept",
				},
			},
		}

		err = src.Send(context.Background(), seed)
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		msgs := getMsgs()
		if len(msgs) != 1 {
			t.Fatalf("expected 1 message, got %d", len(msgs))
		}

		trail, ok := msgs[0].Headers.Custom[DefaultHistoryKey].([]string)
		if !ok {
			t.Fatalf("expected []string trail, got %T", msgs[0].Headers.Custom[DefaultHistoryKey])
		}

		want := []string{"upstream-1", "upstream-2", "stage-z"}
		if len(trail) != len(want) {
			t.Fatalf("trail length: got %d want %d (trail=%v)", len(trail), len(want), trail)
		}

		for i, w := range want {
			if trail[i] != w {
				t.Fatalf("trail[%d]: got %q want %q", i, trail[i], w)
			}
		}

		if msgs[0].Headers.Custom["other-key"] != "kept" {
			t.Fatalf("unrelated Custom key not preserved: %v", msgs[0].Headers.Custom["other-key"])
		}
	})

	t.Run("source map is not mutated", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		_, err := dst.Subscribe(func(_ context.Context, _ messaging.Message[int]) error {
			return nil
		})
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		h := NewHistory("stamp", src, dst)

		err = h.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		sourceTrail := []string{"upstream"}
		sourceCustom := map[string]any{DefaultHistoryKey: sourceTrail}

		err = src.Send(context.Background(), messaging.Message[int]{
			Payload: 1,
			Headers: messaging.Headers{Custom: sourceCustom},
		})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		// Verify the source map was not mutated by the History endpoint.
		updated, ok := sourceCustom[DefaultHistoryKey].([]string)
		if !ok {
			t.Fatalf("source key shape changed: %T", sourceCustom[DefaultHistoryKey])
		}

		if len(updated) != 1 || updated[0] != "upstream" {
			t.Fatalf("source trail mutated: got %v want [upstream]", updated)
		}
	})

	t.Run("foreign value at history key is overwritten", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		dstHandler, getMsgs := captureMessages()

		_, err := dst.Subscribe(dstHandler)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		h := NewHistory("stamp", src, dst)

		err = h.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		// Foreign shape (string instead of []string) — must be replaced.
		err = src.Send(context.Background(), messaging.Message[int]{
			Payload: 1,
			Headers: messaging.Headers{
				Custom: map[string]any{DefaultHistoryKey: "not-a-slice"},
			},
		})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		msgs := getMsgs()
		if len(msgs) != 1 {
			t.Fatalf("expected 1 message, got %d", len(msgs))
		}

		trail, ok := msgs[0].Headers.Custom[DefaultHistoryKey].([]string)
		if !ok {
			t.Fatalf("expected []string trail after overwrite, got %T", msgs[0].Headers.Custom[DefaultHistoryKey])
		}

		if len(trail) != 1 || trail[0] != "stamp" {
			t.Fatalf("expected fresh trail [stamp], got %v", trail)
		}
	})
}

func TestHistory_WithHistoryKey(t *testing.T) {
	t.Parallel()

	t.Run("uses custom key", func(t *testing.T) {
		t.Parallel()

		const customKey = "Path"

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		dstHandler, getMsgs := captureMessages()

		_, err := dst.Subscribe(dstHandler)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		h := NewHistory("stamp", src, dst, WithHistoryKey(customKey))

		err = h.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		msgs := getMsgs()
		if len(msgs) != 1 {
			t.Fatalf("expected 1 message, got %d", len(msgs))
		}

		trail, ok := msgs[0].Headers.Custom[customKey].([]string)
		if !ok {
			t.Fatalf("expected Custom[%q] to be []string, got %T", customKey, msgs[0].Headers.Custom[customKey])
		}

		if len(trail) != 1 || trail[0] != "stamp" {
			t.Fatalf("expected trail [stamp] under %q, got %v", customKey, trail)
		}

		if _, present := msgs[0].Headers.Custom[DefaultHistoryKey]; present {
			t.Fatalf("default key should be absent when overridden, got %v", msgs[0].Headers.Custom)
		}
	})

	t.Run("WithHistoryKey(\"\") is a no-op", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithHistoryKey(""))
		if opts.historyKey != DefaultHistoryKey {
			t.Fatalf("expected default key preserved on empty arg, got %q", opts.historyKey)
		}
	})
}

func TestHistory_ForwardFailure(t *testing.T) {
	t.Parallel()

	t.Run("reports ErrForwardFailed when destination errs", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()

		boom := errors.New("dst down")
		dst := &failingChannel[int]{err: boom}

		errHandler, getErrs := captureErrors()

		h := NewHistory("stamp", src, dst, WithErrorHandler(errHandler))

		err := h.Start(context.Background())
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

		if !errors.Is(errs[0], ErrHistoryFailed) {
			t.Fatalf("expected ErrHistoryFailed, got %v", errs[0])
		}

		if !errors.Is(errs[0], boom) {
			t.Fatalf("expected wrapped destination error, got %v", errs[0])
		}
	})

	t.Run("forward failure does not propagate to source caller", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()

		dst := &failingChannel[int]{err: errors.New("dst down")}

		h := NewHistory("stamp", src, dst, WithErrorHandler(messaging.SilentErrorHandler))

		err := h.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("expected nil from src.Send (history swallows), got %v", err)
		}
	})
}

func TestHistory_Lifecycle(t *testing.T) {
	t.Parallel()

	t.Run("Start is idempotent", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		dstHandler, getMsgs := captureMessages()

		_, err := dst.Subscribe(dstHandler)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		h := NewHistory("stamp", src, dst)

		err = h.Start(context.Background())
		if err != nil {
			t.Fatalf("first start: %v", err)
		}

		err = h.Start(context.Background())
		if err != nil {
			t.Fatalf("second start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		msgs := getMsgs()
		if len(msgs) != 1 {
			t.Fatalf("dst should receive once (no double subscription), got %d", len(msgs))
		}
	})

	t.Run("Stop is idempotent", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		h := NewHistory("stamp", src, dst)

		err := h.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = h.Stop(context.Background())
		if err != nil {
			t.Fatalf("first stop: %v", err)
		}

		err = h.Stop(context.Background())
		if err != nil {
			t.Fatalf("second stop: %v", err)
		}
	})

	t.Run("Done closes after Stop", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		h := NewHistory("stamp", src, dst)

		err := h.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		select {
		case <-h.Done():
			t.Fatal("Done closed before Stop")
		default:
		}

		err = h.Stop(context.Background())
		if err != nil {
			t.Fatalf("stop: %v", err)
		}

		select {
		case <-h.Done():
		default:
			t.Fatal("Done not closed after Stop")
		}
	})

	t.Run("Stop with expired ctx returns ErrShutdownTimeout", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		h := NewHistory("stamp", src, dst)

		err := h.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err = h.Stop(ctx)
		if !errors.Is(err, lifecycle.ErrShutdownTimeout) {
			t.Fatalf("expected ErrShutdownTimeout, got %v", err)
		}
	})
}

func TestHistory_Options(t *testing.T) {
	t.Parallel()

	t.Run("WithErrorHandler(nil) is a no-op", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithErrorHandler(nil))
		if opts.errorHandler == nil {
			t.Fatal("expected default error handler preserved on nil arg")
		}
	})

	t.Run("defaults install DefaultHistoryKey + DefaultErrorHandler", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()
		if opts.historyKey != DefaultHistoryKey {
			t.Fatalf("default history key: got %q want %q", opts.historyKey, DefaultHistoryKey)
		}

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
