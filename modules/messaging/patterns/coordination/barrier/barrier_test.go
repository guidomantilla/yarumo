package barrier

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

// captureDrops returns a thread-safe DropHandler that records every
// dropped message and a getter for the count.
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

func TestNewBarrier(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil component", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		c := NewBarrier("test", src, dst, 3, WithGroupTimeout(time.Second))
		if c == nil {
			t.Fatal("expected non-nil component")
		}
	})

	t.Run("carries the given name", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		c := NewBarrier("orders-barrier", src, dst, 3, WithGroupTimeout(time.Second))
		if c.Name() != "orders-barrier" {
			t.Fatalf("expected name orders-barrier, got %q", c.Name())
		}
	})
}

func TestBarrier_HappyPath(t *testing.T) {
	t.Parallel()

	t.Run("releases all messages when quorum reached", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		dstHandler, getMsgs := captureMessages()

		_, err := dst.Subscribe(dstHandler)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		b := NewBarrier("test", src, dst, 3, WithGroupTimeout(time.Second))

		err = b.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() {
			_ = b.Stop(context.Background())
		})

		for _, p := range []int{10, 20, 30} {
			err = src.Send(context.Background(), messaging.Message[int]{
				Payload: p,
				Headers: messaging.Headers{CorrelationID: "saga-1"},
			})
			if err != nil {
				t.Fatalf("send %d: %v", p, err)
			}
		}

		msgs := getMsgs()
		if len(msgs) != 3 {
			t.Fatalf("expected 3 messages released, got %d", len(msgs))
		}

		want := []int{10, 20, 30}
		for i, w := range want {
			if msgs[i].Payload != w {
				t.Fatalf("msgs[%d].Payload: got %d want %d", i, msgs[i].Payload, w)
			}
		}
	})

	t.Run("nothing released before quorum", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		dstHandler, getMsgs := captureMessages()

		_, err := dst.Subscribe(dstHandler)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		b := NewBarrier("test", src, dst, 3, WithGroupTimeout(time.Second))

		err = b.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() {
			_ = b.Stop(context.Background())
		})

		err = src.Send(context.Background(), messaging.Message[int]{
			Payload: 1,
			Headers: messaging.Headers{CorrelationID: "saga-1"},
		})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{
			Payload: 2,
			Headers: messaging.Headers{CorrelationID: "saga-1"},
		})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		msgs := getMsgs()
		if len(msgs) != 0 {
			t.Fatalf("expected 0 messages before quorum, got %d", len(msgs))
		}
	})
}

func TestBarrier_MultipleCorrelations(t *testing.T) {
	t.Parallel()

	t.Run("each correlation releases independently", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		dstHandler, getMsgs := captureMessages()

		_, err := dst.Subscribe(dstHandler)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		b := NewBarrier("test", src, dst, 2, WithGroupTimeout(time.Second))

		err = b.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() {
			_ = b.Stop(context.Background())
		})

		seq := []struct {
			payload     int
			correlation string
		}{
			{1, "a"},
			{2, "b"},
			{3, "a"}, // releases group a → [1, 3]
			{4, "c"},
			{5, "b"}, // releases group b → [2, 5]
			{6, "c"}, // releases group c → [4, 6]
		}

		for _, s := range seq {
			err = src.Send(context.Background(), messaging.Message[int]{
				Payload: s.payload,
				Headers: messaging.Headers{CorrelationID: s.correlation},
			})
			if err != nil {
				t.Fatalf("send %d: %v", s.payload, err)
			}
		}

		msgs := getMsgs()
		if len(msgs) != 6 {
			t.Fatalf("expected 6 messages, got %d", len(msgs))
		}

		// Verify all payloads from each correlation are present, in the
		// order they were appended within that correlation.
		byCorr := map[string][]int{}
		for _, m := range msgs {
			byCorr[m.Headers.CorrelationID] = append(byCorr[m.Headers.CorrelationID], m.Payload)
		}

		expect := map[string][]int{
			"a": {1, 3},
			"b": {2, 5},
			"c": {4, 6},
		}

		for c, want := range expect {
			got := byCorr[c]
			if len(got) != len(want) {
				t.Fatalf("correlation %q: got %v want %v", c, got, want)
			}

			for i, w := range want {
				if got[i] != w {
					t.Fatalf("correlation %q index %d: got %d want %d", c, i, got[i], w)
				}
			}
		}
	})
}

func TestBarrier_EmptyCorrelation(t *testing.T) {
	t.Parallel()

	t.Run("drops messages with empty CorrelationID", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		dropHandler, getDrops := captureDrops()
		dstHandler, getMsgs := captureMessages()

		_, err := dst.Subscribe(dstHandler)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		b := NewBarrier("test", src, dst, 2, WithGroupTimeout(time.Second), WithDropHandler(dropHandler))

		err = b.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() {
			_ = b.Stop(context.Background())
		})

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got, want := getDrops(), int32(1); got != want {
			t.Fatalf("drop count: got %d want %d", got, want)
		}

		if msgs := getMsgs(); len(msgs) != 0 {
			t.Fatalf("dst should not receive dropped messages, got %d", len(msgs))
		}
	})
}

func TestBarrier_GroupTimeout(t *testing.T) {
	t.Parallel()

	t.Run("drops all messages in a group that did not reach quorum", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		dropHandler, getDrops := captureDrops()
		dstHandler, getMsgs := captureMessages()

		_, err := dst.Subscribe(dstHandler)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		b := NewBarrier("test", src, dst, 3,
			WithGroupTimeout(50*time.Millisecond),
			WithSweepInterval(10*time.Millisecond),
			WithDropHandler(dropHandler))

		err = b.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() {
			_ = b.Stop(context.Background())
		})

		// Send 2 of the required 3 — group never completes.
		err = src.Send(context.Background(), messaging.Message[int]{
			Payload: 1,
			Headers: messaging.Headers{CorrelationID: "stuck"},
		})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{
			Payload: 2,
			Headers: messaging.Headers{CorrelationID: "stuck"},
		})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) {
			if getDrops() >= 2 {
				break
			}

			time.Sleep(20 * time.Millisecond)
		}

		if got, want := getDrops(), int32(2); got != want {
			t.Fatalf("expected 2 drops after timeout, got %d", got)
		}

		if msgs := getMsgs(); len(msgs) != 0 {
			t.Fatalf("dst should not receive timed-out messages, got %d", len(msgs))
		}
	})
}

func TestBarrier_MaxGroups(t *testing.T) {
	t.Parallel()

	t.Run("drops new correlations beyond cap", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		dropHandler, getDrops := captureDrops()

		b := NewBarrier("test", src, dst, 3,
			WithGroupTimeout(time.Second),
			WithMaxGroups(2),
			WithDropHandler(dropHandler))

		err := b.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() {
			_ = b.Stop(context.Background())
		})

		// Fill 2 correlations (each with 1 message — not enough for
		// quorum, so they stay in flight).
		err = src.Send(context.Background(), messaging.Message[int]{
			Payload: 1,
			Headers: messaging.Headers{CorrelationID: "a"},
		})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{
			Payload: 2,
			Headers: messaging.Headers{CorrelationID: "b"},
		})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		// A third correlation arrives — must be dropped.
		err = src.Send(context.Background(), messaging.Message[int]{
			Payload: 3,
			Headers: messaging.Headers{CorrelationID: "c"},
		})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got, want := getDrops(), int32(1); got != want {
			t.Fatalf("drop count: got %d want %d", got, want)
		}

		// Adding to an existing correlation is still allowed.
		err = src.Send(context.Background(), messaging.Message[int]{
			Payload: 11,
			Headers: messaging.Headers{CorrelationID: "a"},
		})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got, want := getDrops(), int32(1); got != want {
			t.Fatalf("drop count after existing-correlation add: got %d want %d", got, want)
		}
	})
}

func TestBarrier_ForwardFailure(t *testing.T) {
	t.Parallel()

	t.Run("reports ErrForwardFailed per message when destination errs", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()

		boom := errors.New("dst down")
		dst := &failingChannel[int]{err: boom}

		errHandler, getErrs := captureErrors()

		b := NewBarrier("test", src, dst, 2,
			WithGroupTimeout(time.Second),
			WithErrorHandler(errHandler))

		err := b.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() {
			_ = b.Stop(context.Background())
		})

		for _, p := range []int{1, 2} {
			err = src.Send(context.Background(), messaging.Message[int]{
				Payload: p,
				Headers: messaging.Headers{CorrelationID: "corr"},
			})
			if err != nil {
				t.Fatalf("send %d: %v", p, err)
			}
		}

		errs := getErrs()
		if len(errs) != 2 {
			t.Fatalf("expected 2 captured errors, got %d", len(errs))
		}

		for _, e := range errs {
			if !errors.Is(e, ErrForwardFailed) {
				t.Fatalf("expected ErrForwardFailed, got %v", e)
			}

			if !errors.Is(e, ErrBarrierFailed) {
				t.Fatalf("expected ErrBarrierFailed, got %v", e)
			}

			if !errors.Is(e, boom) {
				t.Fatalf("expected wrapped destination error, got %v", e)
			}
		}
	})
}

func TestBarrier_Lifecycle(t *testing.T) {
	t.Parallel()

	t.Run("Stop drains in-flight groups via DropHandler", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		dropHandler, getDrops := captureDrops()

		b := NewBarrier("test", src, dst, 3,
			WithGroupTimeout(time.Hour),
			WithDropHandler(dropHandler))

		err := b.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		for _, p := range []int{1, 2} {
			err = src.Send(context.Background(), messaging.Message[int]{
				Payload: p,
				Headers: messaging.Headers{CorrelationID: "stuck"},
			})
			if err != nil {
				t.Fatalf("send: %v", err)
			}
		}

		err = b.Stop(context.Background())
		if err != nil {
			t.Fatalf("stop: %v", err)
		}

		if got, want := getDrops(), int32(2); got != want {
			t.Fatalf("expected 2 drops after Stop, got %d", got)
		}
	})

	t.Run("Start is idempotent", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		dstHandler, getMsgs := captureMessages()

		_, err := dst.Subscribe(dstHandler)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		b := NewBarrier("test", src, dst, 1, WithGroupTimeout(time.Second))

		err = b.Start(context.Background())
		if err != nil {
			t.Fatalf("first start: %v", err)
		}

		err = b.Start(context.Background())
		if err != nil {
			t.Fatalf("second start: %v", err)
		}

		t.Cleanup(func() {
			_ = b.Stop(context.Background())
		})

		err = src.Send(context.Background(), messaging.Message[int]{
			Payload: 1,
			Headers: messaging.Headers{CorrelationID: "single"},
		})
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

		b := NewBarrier("test", src, dst, 2, WithGroupTimeout(time.Second))

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

		b := NewBarrier("test", src, dst, 2, WithGroupTimeout(time.Second))

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

		b := NewBarrier("test", src, dst, 2, WithGroupTimeout(time.Second))

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
}

func TestBarrier_Options(t *testing.T) {
	t.Parallel()

	t.Run("defaults set MaxGroups and SweepInterval", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()
		if opts.maxGroups != DefaultMaxGroups {
			t.Fatalf("default maxGroups: got %d want %d", opts.maxGroups, DefaultMaxGroups)
		}

		if opts.sweepInterval != DefaultSweepInterval {
			t.Fatalf("default sweepInterval: got %v want %v", opts.sweepInterval, DefaultSweepInterval)
		}

		if opts.errorHandler == nil {
			t.Fatal("default error handler should be installed")
		}

		if opts.dropHandler != nil {
			t.Fatal("default drop handler should be nil")
		}
	})

	t.Run("non-positive options are ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(
			WithGroupTimeout(time.Second),
			WithGroupTimeout(0),
			WithMaxGroups(0),
			WithMaxGroups(-5),
			WithSweepInterval(-1),
		)

		if opts.groupTimeout != time.Second {
			t.Fatalf("groupTimeout should be preserved: got %v", opts.groupTimeout)
		}

		if opts.maxGroups != DefaultMaxGroups {
			t.Fatalf("maxGroups should remain default: got %d", opts.maxGroups)
		}

		if opts.sweepInterval != DefaultSweepInterval {
			t.Fatalf("sweepInterval should remain default: got %v", opts.sweepInterval)
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
