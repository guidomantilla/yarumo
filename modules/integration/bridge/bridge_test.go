package bridge

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

func TestNewBridge(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil component", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		c := NewBridge("test", src, dst)
		if c == nil {
			t.Fatal("expected non-nil component")
		}
	})

	t.Run("carries the given name", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		c := NewBridge("orders-bridge", src, dst)
		if c.Name() != "orders-bridge" {
			t.Fatalf("expected name orders-bridge, got %q", c.Name())
		}
	})
}

func TestBridge_HappyPath(t *testing.T) {
	t.Parallel()

	t.Run("forwards every message to destination", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		h, get := counter()

		_, err := dst.Subscribe(h)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		b := NewBridge("test", src, dst)

		err = b.Start(context.Background())
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

	t.Run("preserves payload and headers end-to-end", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		var seen messaging.Message[int]

		h := func(_ context.Context, msg messaging.Message[int]) error {
			seen = msg

			return nil
		}

		_, err := dst.Subscribe(h)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		b := NewBridge("test", src, dst)

		err = b.Start(context.Background())
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

		if seen.Payload != 42 {
			t.Fatalf("payload not preserved: got %d", seen.Payload)
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

func TestBridge_ForwardFailure(t *testing.T) {
	t.Parallel()

	t.Run("reports ErrForwardFailed when destination errs", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()

		boom := errors.New("dst down")
		dst := &failingChannel[int]{err: boom}

		errHandler, getErrs := captureErrors()

		b := NewBridge("test", src, dst, WithErrorHandler(errHandler))

		err := b.Start(context.Background())
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

		if !errors.Is(errs[0], ErrBridgeFailed) {
			t.Fatalf("expected ErrBridgeFailed, got %v", errs[0])
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

		b := NewBridge("test", src, dst, WithErrorHandler(messaging.SilentErrorHandler))

		err := b.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("expected nil from src.Send (bridge swallows), got %v", err)
		}
	})
}

func TestBridge_Lifecycle(t *testing.T) {
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

		b := NewBridge("test", src, dst)

		err = b.Start(context.Background())
		if err != nil {
			t.Fatalf("first start: %v", err)
		}

		err = b.Start(context.Background())
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
		dst := messaging.NewPipelineChannel[int]()

		b := NewBridge("test", src, dst)

		err := b.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = b.Stop(context.Background())
		if err != nil {
			t.Fatalf("first stop: %v", err)
		}

		err = b.Stop(context.Background())
		if err != nil {
			t.Fatalf("second stop: %v", err)
		}
	})

	t.Run("Done closes after Stop", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		b := NewBridge("test", src, dst)

		err := b.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		select {
		case <-b.Done():
			t.Fatal("Done closed before Stop")
		default:
		}

		err = b.Stop(context.Background())
		if err != nil {
			t.Fatalf("stop: %v", err)
		}

		select {
		case <-b.Done():
		default:
			t.Fatal("Done not closed after Stop")
		}
	})

	t.Run("Stop with expired ctx returns ErrShutdownTimeout", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		b := NewBridge("test", src, dst)

		err := b.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err = b.Stop(ctx)
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

		b := NewBridge("test", src, dst)

		err = b.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("send pre-stop: %v", err)
		}

		err = b.Stop(context.Background())
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

func TestBridge_SyncToAsync(t *testing.T) {
	t.Parallel()

	t.Run("forwards sync source to async destination", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewBroadcastChannel[int]()
		dst := messaging.NewTopicChannel[int]("downstream")

		dstComponent, ok := dst.(lifecycle.Component)
		if !ok {
			t.Fatal("TopicChannel should implement lifecycle.Component")
		}

		err := dstComponent.Start(context.Background())
		if err != nil {
			t.Fatalf("dst start: %v", err)
		}

		t.Cleanup(func() {
			_ = dstComponent.Stop(context.Background())
		})

		received := make(chan int, 3)

		_, err = dst.Subscribe(func(_ context.Context, msg messaging.Message[int]) error {
			received <- msg.Payload

			return nil
		})
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		b := NewBridge("sync-to-async", src, dst)

		err = b.Start(context.Background())
		if err != nil {
			t.Fatalf("bridge start: %v", err)
		}

		t.Cleanup(func() {
			_ = b.Stop(context.Background())
		})

		for _, p := range []int{10, 20, 30} {
			err = src.Send(context.Background(), messaging.Message[int]{Payload: p})
			if err != nil {
				t.Fatalf("send %d: %v", p, err)
			}
		}

		got := make(map[int]bool)
		for range 3 {
			select {
			case v := <-received:
				got[v] = true
			case <-context.Background().Done():
				t.Fatal("did not receive in time")
			}
		}

		for _, want := range []int{10, 20, 30} {
			if !got[want] {
				t.Fatalf("missing payload %d, got %v", want, got)
			}
		}
	})
}

func TestBridge_Options(t *testing.T) {
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
