package splitter

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

// captureMessages returns a Handler[U] that records every dispatched
// message, and a getter returning a defensive copy.
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

// splitIntoThree is a happy-path SplitFn[string, int] that returns a
// fixed 3-item slice.
func splitIntoThree(_ context.Context, _ messaging.Message[string]) ([]int, error) {
	return []int{10, 20, 30}, nil
}

// splitToEmpty returns an empty slice, triggering the DropHandler path.
func splitToEmpty(_ context.Context, _ messaging.Message[string]) ([]int, error) {
	return []int{}, nil
}

func TestNewSplitter(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil component", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[string]()
		dst := messaging.NewPipelineChannel[int]()

		c := NewSplitter("test", src, dst, splitIntoThree)
		if c == nil {
			t.Fatal("expected non-nil component")
		}
	})

	t.Run("carries the given name", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[string]()
		dst := messaging.NewPipelineChannel[int]()

		c := NewSplitter("orders-splitter", src, dst, splitIntoThree)
		if c.Name() != "orders-splitter" {
			t.Fatalf("expected name orders-splitter, got %q", c.Name())
		}
	})
}

func TestSplitter_HappyPath(t *testing.T) {
	t.Parallel()

	t.Run("emits one message per slice item with sequence headers", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[string]()
		dst := messaging.NewPipelineChannel[int]()

		captureHandler, getCaptured := captureMessages()

		_, err := dst.Subscribe(captureHandler)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		s := NewSplitter("test", src, dst, splitIntoThree)

		err = s.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[string]{
			Payload: "order-abc",
			Headers: messaging.Headers{
				MessageID:     "msg-1",
				CorrelationID: "corr-1",
				Type:          "test",
			},
		})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		got := getCaptured()
		if len(got) != 3 {
			t.Fatalf("expected 3 emitted children, got %d", len(got))
		}

		for idx, child := range got {
			if child.Headers.SequenceNumber != idx {
				t.Fatalf("child[%d].SequenceNumber: got %d want %d", idx, child.Headers.SequenceNumber, idx)
			}

			if child.Headers.SequenceSize != 3 {
				t.Fatalf("child[%d].SequenceSize: got %d want 3", idx, child.Headers.SequenceSize)
			}

			if child.Headers.CorrelationID != "corr-1" {
				t.Fatalf("child[%d].CorrelationID: got %q want corr-1", idx, child.Headers.CorrelationID)
			}

			if child.Headers.CausationID != "msg-1" {
				t.Fatalf("child[%d].CausationID: got %q want msg-1", idx, child.Headers.CausationID)
			}

			if child.Headers.Type != "test" {
				t.Fatalf("child[%d].Type: got %q want test", idx, child.Headers.Type)
			}
		}

		wantIDs := []string{"msg-1-0", "msg-1-1", "msg-1-2"}
		for idx, want := range wantIDs {
			if got[idx].Headers.MessageID != want {
				t.Fatalf("child[%d].MessageID: got %q want %q", idx, got[idx].Headers.MessageID, want)
			}
		}

		wantPayloads := []int{10, 20, 30}
		for idx, want := range wantPayloads {
			if got[idx].Payload != want {
				t.Fatalf("child[%d].Payload: got %d want %d", idx, got[idx].Payload, want)
			}
		}
	})

	t.Run("single-item slice still populates sequence headers", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[string]()
		dst := messaging.NewPipelineChannel[int]()

		captureHandler, getCaptured := captureMessages()

		_, err := dst.Subscribe(captureHandler)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		singleton := func(_ context.Context, _ messaging.Message[string]) ([]int, error) {
			return []int{42}, nil
		}

		s := NewSplitter("test", src, dst, singleton)

		err = s.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[string]{
			Payload: "one",
			Headers: messaging.Headers{MessageID: "msg-1"},
		})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		got := getCaptured()
		if len(got) != 1 {
			t.Fatalf("expected 1 emitted child, got %d", len(got))
		}

		if got[0].Headers.SequenceNumber != 0 {
			t.Fatalf("SequenceNumber: got %d want 0", got[0].Headers.SequenceNumber)
		}

		if got[0].Headers.SequenceSize != 1 {
			t.Fatalf("SequenceSize: got %d want 1", got[0].Headers.SequenceSize)
		}
	})

	t.Run("source without MessageID still produces unique child IDs", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[string]()
		dst := messaging.NewPipelineChannel[int]()

		captureHandler, getCaptured := captureMessages()

		_, err := dst.Subscribe(captureHandler)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		s := NewSplitter("test", src, dst, splitIntoThree)

		err = s.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[string]{Payload: "x"})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		got := getCaptured()
		if len(got) != 3 {
			t.Fatalf("expected 3, got %d", len(got))
		}

		wantIDs := []string{"0", "1", "2"}
		for idx, want := range wantIDs {
			if got[idx].Headers.MessageID != want {
				t.Fatalf("child[%d].MessageID: got %q want %q", idx, got[idx].Headers.MessageID, want)
			}
		}
	})
}

func TestSplitter_EmptySlice(t *testing.T) {
	t.Parallel()

	t.Run("empty slice fires DropHandler when wired", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[string]()
		dst := messaging.NewPipelineChannel[int]()

		captureHandler, getCaptured := captureMessages()

		_, err := dst.Subscribe(captureHandler)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		dropHandler, getDrops := captureDrops()

		s := NewSplitter("test", src, dst, splitToEmpty, WithDropHandler[int](dropHandler))

		err = s.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[string]{Payload: "ignored"})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got := getDrops(); got != 1 {
			t.Fatalf("drop count: got %d want 1", got)
		}

		if got := getCaptured(); len(got) != 0 {
			t.Fatalf("dst should not receive on empty-slice drop, got %d", len(got))
		}
	})

	t.Run("empty slice silent by default", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[string]()
		dst := messaging.NewPipelineChannel[int]()

		captureHandler, getCaptured := captureMessages()

		_, err := dst.Subscribe(captureHandler)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		errHandler, getErrs := captureErrors()

		s := NewSplitter("test", src, dst, splitToEmpty, WithErrorHandler[int](errHandler))

		err = s.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[string]{Payload: "ignored"})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got := getCaptured(); len(got) != 0 {
			t.Fatalf("dst should not receive on empty-slice drop, got %d", len(got))
		}

		if errs := getErrs(); len(errs) != 0 {
			t.Fatalf("empty-slice drop must NOT fire error handler, got %d errors: %v", len(errs), errs)
		}
	})
}

func TestSplitter_SplitError(t *testing.T) {
	t.Parallel()

	t.Run("wraps split error as ErrSplitFailed", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[string]()
		dst := messaging.NewPipelineChannel[int]()

		boom := errors.New("split boom")

		split := func(_ context.Context, _ messaging.Message[string]) ([]int, error) {
			return nil, boom
		}

		errHandler, getErrs := captureErrors()

		s := NewSplitter("test", src, dst, split, WithErrorHandler[int](errHandler))

		err := s.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[string]{Payload: "x"})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		errs := getErrs()
		if len(errs) != 1 {
			t.Fatalf("expected 1 captured error, got %d", len(errs))
		}

		if !errors.Is(errs[0], ErrSplitFailed) {
			t.Fatalf("expected ErrSplitFailed, got %v", errs[0])
		}

		if !errors.Is(errs[0], ErrSplitterFailed) {
			t.Fatalf("expected ErrSplitterFailed, got %v", errs[0])
		}

		if !errors.Is(errs[0], boom) {
			t.Fatalf("expected wrapped origin error, got %v", errs[0])
		}
	})

	t.Run("wraps split panic as ErrSplitterPanic", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[string]()
		dst := messaging.NewPipelineChannel[int]()

		split := func(_ context.Context, _ messaging.Message[string]) ([]int, error) {
			panic("kaboom")
		}

		errHandler, getErrs := captureErrors()

		s := NewSplitter("test", src, dst, split, WithErrorHandler[int](errHandler))

		err := s.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[string]{Payload: "x"})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		errs := getErrs()
		if len(errs) != 1 {
			t.Fatalf("expected 1 captured error, got %d", len(errs))
		}

		if !errors.Is(errs[0], ErrSplitterPanic) {
			t.Fatalf("expected ErrSplitterPanic, got %v", errs[0])
		}
	})
}

func TestSplitter_ForwardFailure(t *testing.T) {
	t.Parallel()

	t.Run("reports ErrForwardFailed per failed child Send", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[string]()

		boom := errors.New("dst down")
		dst := &failingChannel[int]{err: boom}

		errHandler, getErrs := captureErrors()

		s := NewSplitter("test", src, dst, splitIntoThree, WithErrorHandler[int](errHandler))

		err := s.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[string]{Payload: "x"})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		errs := getErrs()
		if len(errs) != 3 {
			t.Fatalf("expected 3 captured errors (one per child), got %d", len(errs))
		}

		for idx, captured := range errs {
			if !errors.Is(captured, ErrForwardFailed) {
				t.Fatalf("err[%d] expected ErrForwardFailed, got %v", idx, captured)
			}

			if !errors.Is(captured, boom) {
				t.Fatalf("err[%d] expected wrapped destination error, got %v", idx, captured)
			}
		}
	})
}

func TestSplitter_Lifecycle(t *testing.T) {
	t.Parallel()

	t.Run("Start is idempotent", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[string]()
		dst := messaging.NewPipelineChannel[int]()

		captureHandler, getCaptured := captureMessages()

		_, err := dst.Subscribe(captureHandler)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		s := NewSplitter("test", src, dst, splitIntoThree)

		err = s.Start(context.Background())
		if err != nil {
			t.Fatalf("first start: %v", err)
		}

		err = s.Start(context.Background())
		if err != nil {
			t.Fatalf("second start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[string]{Payload: "x"})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got := getCaptured(); len(got) != 3 {
			t.Fatalf("dst should receive 3 (no double subscription), got %d", len(got))
		}
	})

	t.Run("Stop is idempotent", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[string]()
		dst := messaging.NewPipelineChannel[int]()

		s := NewSplitter("test", src, dst, splitIntoThree)

		err := s.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = s.Stop(context.Background())
		if err != nil {
			t.Fatalf("first stop: %v", err)
		}

		err = s.Stop(context.Background())
		if err != nil {
			t.Fatalf("second stop: %v", err)
		}
	})

	t.Run("Done closes after Stop", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[string]()
		dst := messaging.NewPipelineChannel[int]()

		s := NewSplitter("test", src, dst, splitIntoThree)

		err := s.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		select {
		case <-s.Done():
			t.Fatal("Done closed before Stop")
		default:
		}

		err = s.Stop(context.Background())
		if err != nil {
			t.Fatalf("stop: %v", err)
		}

		select {
		case <-s.Done():
		default:
			t.Fatal("Done not closed after Stop")
		}
	})

	t.Run("Stop with expired ctx returns ErrShutdownTimeout", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[string]()
		dst := messaging.NewPipelineChannel[int]()

		s := NewSplitter("test", src, dst, splitIntoThree)

		err := s.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err = s.Stop(ctx)
		if !errors.Is(err, lifecycle.ErrShutdownTimeout) {
			t.Fatalf("expected ErrShutdownTimeout, got %v", err)
		}
	})

	t.Run("Subscription stops receiving after Stop", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[string]()
		dst := messaging.NewPipelineChannel[int]()

		captureHandler, getCaptured := captureMessages()

		_, err := dst.Subscribe(captureHandler)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		s := NewSplitter("test", src, dst, splitIntoThree)

		err = s.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[string]{Payload: "pre"})
		if err != nil {
			t.Fatalf("send pre-stop: %v", err)
		}

		err = s.Stop(context.Background())
		if err != nil {
			t.Fatalf("stop: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[string]{Payload: "post"})
		if err != nil {
			t.Fatalf("send post-stop: %v", err)
		}

		if got := getCaptured(); len(got) != 3 {
			t.Fatalf("post-stop dst should have received only the 3 pre-stop children, got %d", len(got))
		}
	})
}

func TestSplitter_Options(t *testing.T) {
	t.Parallel()

	t.Run("WithErrorHandler(nil) is a no-op", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithErrorHandler[int](nil))
		if opts.errorHandler == nil {
			t.Fatal("expected default error handler preserved on nil arg")
		}
	})

	t.Run("WithDropHandler(nil) is a no-op", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithDropHandler[int](nil))
		if opts.dropHandler != nil {
			t.Fatal("expected default drop handler (nil) preserved on nil arg")
		}
	})

	t.Run("defaults install messaging.DefaultErrorHandler and nil DropHandler", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions[int]()
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
