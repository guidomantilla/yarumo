package aggregator

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
// reported error to the returned slice (via mutex), and a getter that
// returns a defensive copy.
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

// collect returns a Handler[[]int] that appends every received payload
// to a slice and a thread-safe getter for the slice copy.
func collect() (messaging.Handler[[]int], func() [][]int) {
	var mu sync.Mutex

	captured := [][]int{}

	handler := func(_ context.Context, msg messaging.Message[[]int]) error {
		mu.Lock()
		defer mu.Unlock()

		captured = append(captured, msg.Payload)

		return nil
	}

	get := func() [][]int {
		mu.Lock()
		defer mu.Unlock()

		out := make([][]int, len(captured))
		copy(out, captured)

		return out
	}

	return handler, get
}

// sumAggregate folds a group of int messages into a single Message[[]int]
// containing every payload in arrival order.
func sumAggregate(group []messaging.Message[int]) (messaging.Message[[]int], error) {
	out := make([]int, 0, len(group))
	for _, m := range group {
		out = append(out, m.Payload)
	}

	return messaging.Message[[]int]{Payload: out}, nil
}

// msgWith builds a Message[int] with the given payload and correlation
// id. Used in every test instead of NewMessage so tests are independent
// of the uid generator.
func msgWith(payload int, correlation string) messaging.Message[int] {
	return messaging.Message[int]{
		Payload: payload,
		Headers: messaging.Headers{CorrelationID: correlation},
	}
}

func TestNewAggregator(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil aggregator", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[[]int]()

		a := NewAggregator("test", src, dst, sumAggregate, WithCompletionSize[int](2))
		if a == nil {
			t.Fatal("expected non-nil aggregator")
		}
	})

	t.Run("carries the given name", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[[]int]()

		a := NewAggregator("orders-aggregator", src, dst, sumAggregate, WithCompletionSize[int](2))
		if a.Name() != "orders-aggregator" {
			t.Fatalf("expected name orders-aggregator, got %q", a.Name())
		}
	})

	t.Run("panics when no completion strategy configured", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected panic when no completion strategy configured")
			}
		}()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[[]int]()

		_ = NewAggregator("test", src, dst, sumAggregate)
	})
}

func TestAggregator_SizeCompletion(t *testing.T) {
	t.Parallel()

	t.Run("releases group when size reached", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[[]int]()

		sink, getCaptured := collect()

		_, err := dst.Subscribe(sink)
		if err != nil {
			t.Fatalf("subscribe dst: %v", err)
		}

		a := NewAggregator("test", src, dst, sumAggregate, WithCompletionSize[int](3))

		err = a.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		ctx := context.Background()

		for _, p := range []int{1, 2, 3} {
			err = src.Send(ctx, msgWith(p, "g1"))
			if err != nil {
				t.Fatalf("send %d: %v", p, err)
			}
		}

		captured := getCaptured()
		if len(captured) != 1 {
			t.Fatalf("expected 1 aggregate forwarded, got %d", len(captured))
		}

		got := captured[0]
		if len(got) != 3 || got[0] != 1 || got[1] != 2 || got[2] != 3 {
			t.Fatalf("expected [1 2 3], got %v", got)
		}
	})
}

func TestAggregator_MultipleCorrelations(t *testing.T) {
	t.Parallel()

	t.Run("groups complete independently", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[[]int]()

		sink, getCaptured := collect()

		_, err := dst.Subscribe(sink)
		if err != nil {
			t.Fatalf("subscribe dst: %v", err)
		}

		a := NewAggregator("test", src, dst, sumAggregate, WithCompletionSize[int](2))

		err = a.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		ctx := context.Background()

		err = src.Send(ctx, msgWith(1, "a"))
		if err != nil {
			t.Fatalf("send a/1: %v", err)
		}

		err = src.Send(ctx, msgWith(10, "b"))
		if err != nil {
			t.Fatalf("send b/10: %v", err)
		}

		err = src.Send(ctx, msgWith(2, "a"))
		if err != nil {
			t.Fatalf("send a/2: %v", err)
		}

		err = src.Send(ctx, msgWith(20, "b"))
		if err != nil {
			t.Fatalf("send b/20: %v", err)
		}

		captured := getCaptured()
		if len(captured) != 2 {
			t.Fatalf("expected 2 aggregates forwarded, got %d", len(captured))
		}

		// Order between groups is deterministic because the source is a
		// PipelineChannel (sync dispatch) — group "a" completes first.
		if captured[0][0] != 1 || captured[0][1] != 2 {
			t.Fatalf("first aggregate (group a): expected [1 2], got %v", captured[0])
		}

		if captured[1][0] != 10 || captured[1][1] != 20 {
			t.Fatalf("second aggregate (group b): expected [10 20], got %v", captured[1])
		}
	})
}

func TestAggregator_TimeoutCompletion(t *testing.T) {
	t.Parallel()

	t.Run("sweeper releases idle partial group", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[[]int]()

		sink, getCaptured := collect()

		_, err := dst.Subscribe(sink)
		if err != nil {
			t.Fatalf("subscribe dst: %v", err)
		}

		a := NewAggregator("test", src, dst, sumAggregate,
			WithCompletionSize[int](5),
			WithGroupTimeout[int](50*time.Millisecond))

		err = a.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		defer func() {
			_ = a.Stop(context.Background())
		}()

		ctx := context.Background()

		err = src.Send(ctx, msgWith(7, "slow"))
		if err != nil {
			t.Fatalf("send 7: %v", err)
		}

		err = src.Send(ctx, msgWith(8, "slow"))
		if err != nil {
			t.Fatalf("send 8: %v", err)
		}

		deadline := time.Now().Add(500 * time.Millisecond)

		for time.Now().Before(deadline) {
			if len(getCaptured()) >= 1 {
				break
			}

			time.Sleep(10 * time.Millisecond)
		}

		captured := getCaptured()
		if len(captured) != 1 {
			t.Fatalf("expected 1 aggregate released by sweeper, got %d", len(captured))
		}

		got := captured[0]
		if len(got) != 2 || got[0] != 7 || got[1] != 8 {
			t.Fatalf("expected [7 8] sweeper release, got %v", got)
		}
	})
}

func TestAggregator_EmptyCorrelation(t *testing.T) {
	t.Parallel()

	t.Run("DropHandler fires when correlation key is empty", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[[]int]()

		dropHandler, getDrops := captureDrops()
		errHandler, getErrs := captureErrors()

		a := NewAggregator("test", src, dst, sumAggregate,
			WithCompletionSize[int](2),
			WithErrorHandler[int](errHandler),
			WithDropHandler[int](dropHandler))

		err := a.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		// Message has empty Headers.CorrelationID by construction.
		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got, want := getDrops(), int32(1); got != want {
			t.Fatalf("DropHandler should fire once for empty correlation, got %d", got)
		}

		if errs := getErrs(); len(errs) != 0 {
			t.Fatalf("ErrorHandler should not fire on empty correlation, got %v", errs)
		}
	})
}

func TestAggregator_MaxGroupsExceeded(t *testing.T) {
	t.Parallel()

	t.Run("third correlation reports ErrMaxGroupsExceeded", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[[]int]()

		errHandler, getErrs := captureErrors()

		a := NewAggregator("test", src, dst, sumAggregate,
			WithCompletionSize[int](10),
			WithMaxGroups[int](2),
			WithErrorHandler[int](errHandler))

		err := a.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		ctx := context.Background()

		err = src.Send(ctx, msgWith(1, "a"))
		if err != nil {
			t.Fatalf("send a: %v", err)
		}

		err = src.Send(ctx, msgWith(2, "b"))
		if err != nil {
			t.Fatalf("send b: %v", err)
		}

		err = src.Send(ctx, msgWith(3, "c"))
		if err != nil {
			t.Fatalf("send c: %v", err)
		}

		errs := getErrs()
		if len(errs) != 1 {
			t.Fatalf("expected 1 captured error, got %d: %v", len(errs), errs)
		}

		if !errors.Is(errs[0], ErrMaxGroupsExceeded) {
			t.Fatalf("expected ErrMaxGroupsExceeded, got %v", errs[0])
		}

		if !errors.Is(errs[0], ErrAggregateFailed) {
			t.Fatalf("expected ErrAggregateFailed wrapper, got %v", errs[0])
		}
	})
}

func TestAggregator_AggregateFnError(t *testing.T) {
	t.Parallel()

	t.Run("wraps AggregateFn error as ErrAggregateFnFailed", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[[]int]()

		boom := errors.New("aggregate boom")

		failing := func(_ []messaging.Message[int]) (messaging.Message[[]int], error) {
			return messaging.Message[[]int]{}, boom
		}

		errHandler, getErrs := captureErrors()

		a := NewAggregator("test", src, dst, failing,
			WithCompletionSize[int](1),
			WithErrorHandler[int](errHandler))

		err := a.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), msgWith(1, "g"))
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		errs := getErrs()
		if len(errs) != 1 {
			t.Fatalf("expected 1 captured error, got %d", len(errs))
		}

		if !errors.Is(errs[0], ErrAggregateFnFailed) {
			t.Fatalf("expected ErrAggregateFnFailed, got %v", errs[0])
		}

		if !errors.Is(errs[0], boom) {
			t.Fatalf("expected wrapped origin error, got %v", errs[0])
		}
	})

	t.Run("wraps AggregateFn panic as ErrAggregateFnFailed", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[[]int]()

		panicking := func(_ []messaging.Message[int]) (messaging.Message[[]int], error) {
			panic("kaboom")
		}

		errHandler, getErrs := captureErrors()

		a := NewAggregator("test", src, dst, panicking,
			WithCompletionSize[int](1),
			WithErrorHandler[int](errHandler))

		err := a.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), msgWith(1, "g"))
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		errs := getErrs()
		if len(errs) != 1 {
			t.Fatalf("expected 1 captured error, got %d", len(errs))
		}

		if !errors.Is(errs[0], ErrAggregateFnFailed) {
			t.Fatalf("expected ErrAggregateFnFailed wrapping panic, got %v", errs[0])
		}
	})
}

func TestAggregator_ForwardFailure(t *testing.T) {
	t.Parallel()

	t.Run("reports ErrForwardFailed when dst Send errs", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()

		boom := errors.New("dst down")
		dst := &failingChannel[[]int]{err: boom}

		errHandler, getErrs := captureErrors()

		a := NewAggregator("test", src, dst, sumAggregate,
			WithCompletionSize[int](1),
			WithErrorHandler[int](errHandler))

		err := a.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), msgWith(1, "g"))
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

func TestAggregator_StopDrains(t *testing.T) {
	t.Parallel()

	t.Run("Stop releases remaining incomplete groups", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[[]int]()

		sink, getCaptured := collect()

		_, err := dst.Subscribe(sink)
		if err != nil {
			t.Fatalf("subscribe dst: %v", err)
		}

		a := NewAggregator("test", src, dst, sumAggregate, WithCompletionSize[int](5))

		err = a.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		ctx := context.Background()

		err = src.Send(ctx, msgWith(1, "g"))
		if err != nil {
			t.Fatalf("send 1: %v", err)
		}

		err = src.Send(ctx, msgWith(2, "g"))
		if err != nil {
			t.Fatalf("send 2: %v", err)
		}

		// Before Stop the partial group has not been released yet.
		if got := getCaptured(); len(got) != 0 {
			t.Fatalf("partial group should not be released before Stop, got %d", len(got))
		}

		err = a.Stop(ctx)
		if err != nil {
			t.Fatalf("stop: %v", err)
		}

		captured := getCaptured()
		if len(captured) != 1 {
			t.Fatalf("expected 1 aggregate drained on Stop, got %d", len(captured))
		}

		got := captured[0]
		if len(got) != 2 || got[0] != 1 || got[1] != 2 {
			t.Fatalf("expected drained aggregate [1 2], got %v", got)
		}
	})
}

func TestAggregator_ConcurrentProducers(t *testing.T) {
	t.Parallel()

	t.Run("concurrent Send for same correlation completes once", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewTopicChannel[int]("src", messaging.WithBufferSize(64))
		dst := messaging.NewPipelineChannel[[]int]()

		sink, getCaptured := collect()

		_, err := dst.Subscribe(sink)
		if err != nil {
			t.Fatalf("subscribe dst: %v", err)
		}

		const total = 10

		a := NewAggregator("test", src, dst, sumAggregate, WithCompletionSize[int](total))

		err = src.(lifecycle.Component).Start(context.Background())
		if err != nil {
			t.Fatalf("start src: %v", err)
		}

		err = a.Start(context.Background())
		if err != nil {
			t.Fatalf("start aggregator: %v", err)
		}

		defer func() {
			_ = a.Stop(context.Background())
			_ = src.(lifecycle.Component).Stop(context.Background())
		}()

		ctx := context.Background()

		var wg sync.WaitGroup
		for i := range total {
			wg.Add(1)

			go func(payload int) {
				defer wg.Done()

				err := src.Send(ctx, msgWith(payload, "concurrent"))
				if err != nil {
					t.Errorf("send %d: %v", payload, err)
				}
			}(i)
		}

		wg.Wait()

		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) {
			if len(getCaptured()) >= 1 {
				break
			}

			time.Sleep(5 * time.Millisecond)
		}

		captured := getCaptured()
		if len(captured) != 1 {
			t.Fatalf("expected exactly 1 aggregate from concurrent producers, got %d", len(captured))
		}

		if len(captured[0]) != total {
			t.Fatalf("expected aggregate of %d messages, got %d", total, len(captured[0]))
		}
	})
}

func TestAggregator_StopDetaches(t *testing.T) {
	t.Parallel()

	t.Run("post-Stop messages do not fire handle", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[[]int]()

		sink, getCaptured := collect()

		_, err := dst.Subscribe(sink)
		if err != nil {
			t.Fatalf("subscribe dst: %v", err)
		}

		a := NewAggregator("test", src, dst, sumAggregate, WithCompletionSize[int](2))

		err = a.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), msgWith(1, "g"))
		if err != nil {
			t.Fatalf("send pre-stop: %v", err)
		}

		err = a.Stop(context.Background())
		if err != nil {
			t.Fatalf("stop: %v", err)
		}

		// Pre-stop accumulated [1] was drained as a partial aggregate on
		// Stop — record that count and confirm post-Stop sends do not
		// add new aggregates.
		base := len(getCaptured())

		err = src.Send(context.Background(), msgWith(2, "g"))
		if err != nil {
			t.Fatalf("send post-stop: %v", err)
		}

		err = src.Send(context.Background(), msgWith(3, "g"))
		if err != nil {
			t.Fatalf("send post-stop: %v", err)
		}

		if got := len(getCaptured()); got != base {
			t.Fatalf("post-Stop messages should not trigger aggregation: base %d, after %d", base, got)
		}
	})
}

func TestAggregator_Lifecycle(t *testing.T) {
	t.Parallel()

	t.Run("Start is idempotent", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[[]int]()

		sink, getCaptured := collect()

		_, err := dst.Subscribe(sink)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		a := NewAggregator("test", src, dst, sumAggregate, WithCompletionSize[int](1))

		err = a.Start(context.Background())
		if err != nil {
			t.Fatalf("first start: %v", err)
		}

		err = a.Start(context.Background())
		if err != nil {
			t.Fatalf("second start: %v", err)
		}

		err = src.Send(context.Background(), msgWith(1, "g"))
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got := len(getCaptured()); got != 1 {
			t.Fatalf("dst should receive once even with double Start, got %d", got)
		}
	})

	t.Run("Stop is idempotent", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[[]int]()

		a := NewAggregator("test", src, dst, sumAggregate, WithCompletionSize[int](2))

		err := a.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = a.Stop(context.Background())
		if err != nil {
			t.Fatalf("first stop: %v", err)
		}

		err = a.Stop(context.Background())
		if err != nil {
			t.Fatalf("second stop: %v", err)
		}
	})

	t.Run("Done closes after Stop", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[[]int]()

		a := NewAggregator("test", src, dst, sumAggregate, WithCompletionSize[int](2))

		err := a.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		select {
		case <-a.Done():
			t.Fatal("Done closed before Stop")
		default:
		}

		err = a.Stop(context.Background())
		if err != nil {
			t.Fatalf("stop: %v", err)
		}

		select {
		case <-a.Done():
		default:
			t.Fatal("Done not closed after Stop")
		}
	})

	t.Run("Stop with expired ctx returns ErrShutdownTimeout", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[[]int]()

		a := NewAggregator("test", src, dst, sumAggregate, WithCompletionSize[int](2))

		err := a.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err = a.Stop(ctx)
		if !errors.Is(err, lifecycle.ErrShutdownTimeout) {
			t.Fatalf("expected ErrShutdownTimeout, got %v", err)
		}
	})
}

func TestAggregator_PredicateCompletion(t *testing.T) {
	t.Parallel()

	t.Run("predicate-based completion releases when fn returns true", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[[]int]()

		sink, getCaptured := collect()

		_, err := dst.Subscribe(sink)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		predicate := func(group []messaging.Message[int]) bool {
			// release when last payload is 0 (sentinel "END" marker).
			return len(group) > 0 && group[len(group)-1].Payload == 0
		}

		a := NewAggregator("test", src, dst, sumAggregate, WithCompletionFn[int](predicate))

		err = a.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		ctx := context.Background()

		for _, p := range []int{4, 5, 6, 0} {
			err = src.Send(ctx, msgWith(p, "g"))
			if err != nil {
				t.Fatalf("send %d: %v", p, err)
			}
		}

		captured := getCaptured()
		if len(captured) != 1 {
			t.Fatalf("expected 1 aggregate on END marker, got %d", len(captured))
		}

		got := captured[0]
		if len(got) != 4 || got[3] != 0 {
			t.Fatalf("expected [4 5 6 0], got %v", got)
		}
	})
}

func TestAggregator_Options(t *testing.T) {
	t.Parallel()

	t.Run("WithCorrelationFn(nil) is a no-op", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions[int](WithCorrelationFn[int](nil))
		if opts.correlation == nil {
			t.Fatal("expected default correlation preserved on nil arg")
		}
	})

	t.Run("WithCompletionFn(nil) is a no-op", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions[int](WithCompletionFn[int](nil))
		if opts.completion != nil {
			t.Fatal("expected no completion fn installed for nil arg")
		}
	})

	t.Run("WithCompletionSize(0) is a no-op", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions[int](WithCompletionSize[int](0))
		if opts.completionSize != 0 {
			t.Fatal("expected completionSize unchanged on non-positive arg")
		}
	})

	t.Run("WithGroupTimeout(0) is a no-op", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions[int](WithGroupTimeout[int](0))
		if opts.groupTimeout != 0 {
			t.Fatal("expected groupTimeout unchanged on non-positive arg")
		}
	})

	t.Run("WithMaxGroups(0) is a no-op", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions[int](WithMaxGroups[int](0))
		if opts.maxGroups != defaultMaxGroups {
			t.Fatalf("expected default maxGroups preserved, got %d", opts.maxGroups)
		}
	})

	t.Run("WithErrorHandler(nil) is a no-op", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions[int](WithErrorHandler[int](nil))
		if opts.errorHandler == nil {
			t.Fatal("expected default error handler preserved on nil arg")
		}
	})

	t.Run("WithDropHandler(nil) is a no-op", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions[int](WithDropHandler[int](nil))
		if opts.dropHandler != nil {
			t.Fatal("expected nil drop handler preserved on nil arg")
		}
	})

	t.Run("defaults install messaging.DefaultErrorHandler and headers correlation", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions[int]()
		if opts.errorHandler == nil {
			t.Fatal("default error handler should be installed")
		}

		if opts.correlation == nil {
			t.Fatal("default correlation extractor should be installed")
		}

		msg := messaging.Message[int]{Headers: messaging.Headers{CorrelationID: "abc"}}
		if got := opts.correlation(msg); got != "abc" {
			t.Fatalf("default correlation should read Headers.CorrelationID, got %q", got)
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
