package resequencer

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

// seqMsg constructs a Message[int] carrying the given correlation,
// sequence size and 0-based sequence number.
func seqMsg(payload int, correlation string, seqNumber, size int) messaging.Message[int] {
	return messaging.Message[int]{
		Payload: payload,
		Headers: messaging.Headers{
			CorrelationID:  correlation,
			SequenceNumber: seqNumber,
			SequenceSize:   size,
		},
	}
}

func TestNewResequencer(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil component", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		c := NewResequencer("test", src, dst, WithGroupTimeout(time.Second))
		if c == nil {
			t.Fatal("expected non-nil component")
		}
	})

	t.Run("carries the given name", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		c := NewResequencer("orders-resequencer", src, dst, WithGroupTimeout(time.Second))
		if c.Name() != "orders-resequencer" {
			t.Fatalf("expected name orders-resequencer, got %q", c.Name())
		}
	})
}

func TestResequencer_HappyPath(t *testing.T) {
	t.Parallel()

	t.Run("emits in order when received out of order", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		dstHandler, getMsgs := captureMessages()

		_, err := dst.Subscribe(dstHandler)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		r := NewResequencer("test", src, dst, WithGroupTimeout(time.Second))

		err = r.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() {
			_ = r.Stop(context.Background())
		})

		// Arrive out of order: 2, 0, 1 (sequence size 3).
		err = src.Send(context.Background(), seqMsg(20, "c", 2, 3))
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		err = src.Send(context.Background(), seqMsg(0, "c", 0, 3))
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		err = src.Send(context.Background(), seqMsg(10, "c", 1, 3))
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		msgs := getMsgs()
		if len(msgs) != 3 {
			t.Fatalf("expected 3 messages, got %d", len(msgs))
		}

		want := []int{0, 10, 20}
		for i, w := range want {
			if msgs[i].Payload != w {
				t.Fatalf("msgs[%d].Payload: got %d want %d", i, msgs[i].Payload, w)
			}
		}
	})

	t.Run("emits incrementally as next position arrives", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		dstHandler, getMsgs := captureMessages()

		_, err := dst.Subscribe(dstHandler)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		r := NewResequencer("test", src, dst, WithGroupTimeout(time.Second))

		err = r.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() {
			_ = r.Stop(context.Background())
		})

		// 1 arrives first — buffered, no emit yet.
		err = src.Send(context.Background(), seqMsg(10, "c", 1, 3))
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got := len(getMsgs()); got != 0 {
			t.Fatalf("expected 0 emits after seq=1, got %d", got)
		}

		// 0 arrives — emits 0 and 1 (cursor advances to 2).
		err = src.Send(context.Background(), seqMsg(0, "c", 0, 3))
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		msgs := getMsgs()
		if len(msgs) != 2 {
			t.Fatalf("expected 2 emits after seq=0, got %d", len(msgs))
		}

		if msgs[0].Payload != 0 || msgs[1].Payload != 10 {
			t.Fatalf("expected payloads [0, 10], got [%d, %d]", msgs[0].Payload, msgs[1].Payload)
		}

		// 2 arrives — emits 2, group complete.
		err = src.Send(context.Background(), seqMsg(20, "c", 2, 3))
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		msgs = getMsgs()
		if len(msgs) != 3 {
			t.Fatalf("expected 3 emits after seq=2, got %d", len(msgs))
		}

		if msgs[2].Payload != 20 {
			t.Fatalf("msgs[2].Payload: got %d want 20", msgs[2].Payload)
		}
	})
}

func TestResequencer_ConcurrentProducers(t *testing.T) {
	t.Parallel()

	t.Run("each correlation is emitted in seq order", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewTopicChannel[int]("src")
		dst := messaging.NewPipelineChannel[int]()

		topicComponent, ok := src.(lifecycle.Component)
		if !ok {
			t.Fatal("TopicChannel must implement lifecycle.Component")
		}

		err := topicComponent.Start(context.Background())
		if err != nil {
			t.Fatalf("topic start: %v", err)
		}

		t.Cleanup(func() {
			_ = topicComponent.Stop(context.Background())
		})

		dstHandler, getMsgs := captureMessages()

		_, err = dst.Subscribe(dstHandler)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		r := NewResequencer("test", src, dst, WithGroupTimeout(2*time.Second))

		err = r.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() {
			_ = r.Stop(context.Background())
		})

		const correlations = 5
		const size = 4

		var wg sync.WaitGroup

		for c := range correlations {
			wg.Add(1)

			go func(corr int) {
				defer wg.Done()

				correlation := correlationName(corr)

				// Send in reversed order to maximize reordering.
				for i := size - 1; i >= 0; i-- {
					_ = src.Send(context.Background(), seqMsg(corr*100+i, correlation, i, size))
				}
			}(c)
		}

		wg.Wait()

		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) {
			if len(getMsgs()) >= correlations*size {
				break
			}

			time.Sleep(20 * time.Millisecond)
		}

		msgs := getMsgs()
		if len(msgs) != correlations*size {
			t.Fatalf("expected %d msgs, got %d", correlations*size, len(msgs))
		}

		// Verify per-correlation order.
		seenByCorr := map[string][]int{}
		for _, m := range msgs {
			seenByCorr[m.Headers.CorrelationID] = append(seenByCorr[m.Headers.CorrelationID], m.Headers.SequenceNumber)
		}

		for corr, seq := range seenByCorr {
			if len(seq) != size {
				t.Fatalf("correlation %q: got %d msgs want %d", corr, len(seq), size)
			}

			for i, s := range seq {
				if s != i {
					t.Fatalf("correlation %q: out-of-order seq at index %d: got %d want %d", corr, i, s, i)
				}
			}
		}
	})
}

func correlationName(i int) string {
	return string(rune('a' + i))
}

func TestResequencer_Drops(t *testing.T) {
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

		r := NewResequencer("test", src, dst,
			WithGroupTimeout(time.Second),
			WithDropHandler(dropHandler))

		err = r.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() {
			_ = r.Stop(context.Background())
		})

		err = src.Send(context.Background(), messaging.Message[int]{
			Payload: 1,
			Headers: messaging.Headers{SequenceNumber: 0, SequenceSize: 1},
		})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got, want := getDrops(), int32(1); got != want {
			t.Fatalf("drop count: got %d want %d", got, want)
		}

		if got := len(getMsgs()); got != 0 {
			t.Fatalf("dst should not receive dropped messages, got %d", got)
		}
	})

	t.Run("drops messages with SequenceSize=0", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		dropHandler, getDrops := captureDrops()

		r := NewResequencer("test", src, dst,
			WithGroupTimeout(time.Second),
			WithDropHandler(dropHandler))

		err := r.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() {
			_ = r.Stop(context.Background())
		})

		err = src.Send(context.Background(), messaging.Message[int]{
			Payload: 1,
			Headers: messaging.Headers{CorrelationID: "c", SequenceNumber: 0, SequenceSize: 0},
		})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got, want := getDrops(), int32(1); got != want {
			t.Fatalf("drop count: got %d want %d", got, want)
		}
	})

	t.Run("drops messages with out-of-range SequenceNumber", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		dropHandler, getDrops := captureDrops()

		r := NewResequencer("test", src, dst,
			WithGroupTimeout(time.Second),
			WithDropHandler(dropHandler))

		err := r.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() {
			_ = r.Stop(context.Background())
		})

		err = src.Send(context.Background(), seqMsg(1, "c", 5, 3))
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		err = src.Send(context.Background(), seqMsg(1, "c", -1, 3))
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got, want := getDrops(), int32(2); got != want {
			t.Fatalf("drop count: got %d want %d", got, want)
		}
	})

	t.Run("drops messages with mismatched SequenceSize", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		dropHandler, getDrops := captureDrops()

		r := NewResequencer("test", src, dst,
			WithGroupTimeout(time.Second),
			WithDropHandler(dropHandler))

		err := r.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() {
			_ = r.Stop(context.Background())
		})

		// Establish a group with expected size 3.
		err = src.Send(context.Background(), seqMsg(1, "c", 0, 3))
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		// Subsequent message disagrees on size.
		err = src.Send(context.Background(), seqMsg(2, "c", 1, 5))
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got, want := getDrops(), int32(1); got != want {
			t.Fatalf("drop count: got %d want %d", got, want)
		}
	})

	t.Run("drops duplicate sequence numbers", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		dropHandler, getDrops := captureDrops()

		r := NewResequencer("test", src, dst,
			WithGroupTimeout(time.Second),
			WithDropHandler(dropHandler))

		err := r.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() {
			_ = r.Stop(context.Background())
		})

		err = src.Send(context.Background(), seqMsg(10, "c", 1, 3))
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		err = src.Send(context.Background(), seqMsg(11, "c", 1, 3))
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got, want := getDrops(), int32(1); got != want {
			t.Fatalf("drop count: got %d want %d", got, want)
		}
	})
}

func TestResequencer_GroupTimeout(t *testing.T) {
	t.Parallel()

	t.Run("drops buffered tail when missing position never arrives", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		dropHandler, getDrops := captureDrops()
		dstHandler, getMsgs := captureMessages()

		_, err := dst.Subscribe(dstHandler)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		r := NewResequencer("test", src, dst,
			WithGroupTimeout(50*time.Millisecond),
			WithSweepInterval(10*time.Millisecond),
			WithDropHandler(dropHandler))

		err = r.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() {
			_ = r.Stop(context.Background())
		})

		// Send seq=1 and seq=2 for size=3 — seq=0 never arrives.
		err = src.Send(context.Background(), seqMsg(10, "c", 1, 3))
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		err = src.Send(context.Background(), seqMsg(20, "c", 2, 3))
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

		if got := len(getMsgs()); got != 0 {
			t.Fatalf("dst should not receive emitted messages, got %d", got)
		}
	})
}

func TestResequencer_MaxGroups(t *testing.T) {
	t.Parallel()

	t.Run("drops new correlations beyond cap", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		dropHandler, getDrops := captureDrops()

		r := NewResequencer("test", src, dst,
			WithGroupTimeout(time.Second),
			WithMaxGroups(2),
			WithDropHandler(dropHandler))

		err := r.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() {
			_ = r.Stop(context.Background())
		})

		// Fill 2 correlations (seq=1 of size=3 each — incomplete, stays in flight).
		err = src.Send(context.Background(), seqMsg(1, "a", 1, 3))
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		err = src.Send(context.Background(), seqMsg(2, "b", 1, 3))
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		// Third correlation must be dropped.
		err = src.Send(context.Background(), seqMsg(3, "c", 1, 3))
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got, want := getDrops(), int32(1); got != want {
			t.Fatalf("drop count: got %d want %d", got, want)
		}
	})
}

func TestResequencer_ForwardFailure(t *testing.T) {
	t.Parallel()

	t.Run("reports ErrForwardFailed when destination errs", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()

		boom := errors.New("dst down")
		dst := &failingChannel[int]{err: boom}

		errHandler, getErrs := captureErrors()

		r := NewResequencer("test", src, dst,
			WithGroupTimeout(time.Second),
			WithErrorHandler(errHandler))

		err := r.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() {
			_ = r.Stop(context.Background())
		})

		err = src.Send(context.Background(), seqMsg(1, "c", 0, 1))
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

		if !errors.Is(errs[0], ErrResequencerFailed) {
			t.Fatalf("expected ErrResequencerFailed, got %v", errs[0])
		}

		if !errors.Is(errs[0], boom) {
			t.Fatalf("expected wrapped destination error, got %v", errs[0])
		}
	})
}

func TestResequencer_Lifecycle(t *testing.T) {
	t.Parallel()

	t.Run("Stop drains in-flight groups via DropHandler", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		dropHandler, getDrops := captureDrops()

		r := NewResequencer("test", src, dst,
			WithGroupTimeout(time.Hour),
			WithDropHandler(dropHandler))

		err := r.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		// Buffer 2 of 3 — incomplete, stays in flight.
		err = src.Send(context.Background(), seqMsg(10, "c", 1, 3))
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		err = src.Send(context.Background(), seqMsg(20, "c", 2, 3))
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		err = r.Stop(context.Background())
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

		r := NewResequencer("test", src, dst, WithGroupTimeout(time.Second))

		err = r.Start(context.Background())
		if err != nil {
			t.Fatalf("first start: %v", err)
		}

		err = r.Start(context.Background())
		if err != nil {
			t.Fatalf("second start: %v", err)
		}

		t.Cleanup(func() {
			_ = r.Stop(context.Background())
		})

		err = src.Send(context.Background(), seqMsg(1, "c", 0, 1))
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

		r := NewResequencer("test", src, dst, WithGroupTimeout(time.Second))

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

		r := NewResequencer("test", src, dst, WithGroupTimeout(time.Second))

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

		r := NewResequencer("test", src, dst, WithGroupTimeout(time.Second))

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

func TestResequencer_Options(t *testing.T) {
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
