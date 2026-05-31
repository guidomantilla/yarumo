package scattergather

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

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

// captureDrops returns a thread-safe DropHandler that counts drops
// and a getter returning the current count.
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

// captureMessages returns a Handler[U] that captures every received
// Message[U] payload into the returned slice and a thread-safe
// getter for a defensive copy.
func captureMessages[U any]() (messaging.Handler[U], func() []messaging.Message[U]) {
	var mu sync.Mutex

	captured := []messaging.Message[U]{}

	handler := func(_ context.Context, msg messaging.Message[U]) error {
		mu.Lock()
		defer mu.Unlock()

		captured = append(captured, msg)

		return nil
	}

	get := func() []messaging.Message[U] {
		mu.Lock()
		defer mu.Unlock()

		out := make([]messaging.Message[U], len(captured))
		copy(out, captured)

		return out
	}

	return handler, get
}

// sumAggregate folds a group of int messages into a single
// Message[int] whose payload is the sum of every received payload.
// CorrelationID of the result is inherited from the first message so
// downstream tests can assert per-request output identity.
func sumAggregate(group []messaging.Message[int]) (messaging.Message[int], error) {
	total := 0
	for _, m := range group {
		total += m.Payload
	}

	corrID := ""
	if len(group) > 0 {
		corrID = group[0].Headers.CorrelationID
	}

	return messaging.Message[int]{
		Payload: total,
		Headers: messaging.Headers{CorrelationID: corrID},
	}, nil
}

// reqMsg builds a Message[int] with the given payload and correlation
// id. Used in every test instead of NewMessage so tests are
// independent of the uid generator.
func reqMsg(payload int, correlation string) messaging.Message[int] {
	return messaging.Message[int]{
		Payload: payload,
		Headers: messaging.Headers{CorrelationID: correlation},
	}
}

// startWorker subscribes a Handler[int] to the given worker channel
// that copies the received message (preserving CorrelationID) onto
// replyChan with payloadFn applied to the payload. It returns the
// Cancel so tests can shut down the worker explicitly.
func startWorker(t *testing.T, worker messaging.Channel[int], replyChan messaging.Channel[int], payloadFn func(int) int) messaging.Cancel {
	t.Helper()

	cancel, err := worker.Subscribe(func(ctx context.Context, msg messaging.Message[int]) error {
		reply := messaging.Message[int]{
			Payload: payloadFn(msg.Payload),
			Headers: messaging.Headers{CorrelationID: msg.Headers.CorrelationID},
		}

		return replyChan.Send(ctx, reply)
	})
	if err != nil {
		t.Fatalf("subscribe worker: %v", err)
	}

	return cancel
}

func TestNewScatterGather(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil component", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		reply := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		worker := messaging.NewPipelineChannel[int]()

		selector := func(_ context.Context, _ messaging.Message[int]) ([]string, error) {
			return []string{"w"}, nil
		}

		sg := NewScatterGather("test", src, map[string]messaging.Channel[int]{"w": worker},
			reply, dst, selector, sumAggregate,
			WithGroupTimeout[int](100*time.Millisecond))
		if sg == nil {
			t.Fatal("expected non-nil component")
		}
	})

	t.Run("carries the given name", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		reply := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		worker := messaging.NewPipelineChannel[int]()

		selector := func(_ context.Context, _ messaging.Message[int]) ([]string, error) {
			return []string{"w"}, nil
		}

		sg := NewScatterGather("orders-sg", src, map[string]messaging.Channel[int]{"w": worker},
			reply, dst, selector, sumAggregate,
			WithGroupTimeout[int](100*time.Millisecond))
		if sg.Name() != "orders-sg" {
			t.Fatalf("expected name orders-sg, got %q", sg.Name())
		}
	})

	t.Run("panics without WithGroupTimeout", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected panic when WithGroupTimeout is missing")
			}
		}()

		src := messaging.NewPipelineChannel[int]()
		reply := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		worker := messaging.NewPipelineChannel[int]()

		selector := func(_ context.Context, _ messaging.Message[int]) ([]string, error) {
			return []string{"w"}, nil
		}

		_ = NewScatterGather("test", src, map[string]messaging.Channel[int]{"w": worker},
			reply, dst, selector, sumAggregate)
	})

	t.Run("panics with empty workers map", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected panic on empty workers map")
			}
		}()

		src := messaging.NewPipelineChannel[int]()
		reply := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		selector := func(_ context.Context, _ messaging.Message[int]) ([]string, error) {
			return []string{"w"}, nil
		}

		_ = NewScatterGather("test", src, map[string]messaging.Channel[int]{},
			reply, dst, selector, sumAggregate,
			WithGroupTimeout[int](100*time.Millisecond))
	})
}

func TestScatterGather_HappyPath(t *testing.T) {
	t.Parallel()

	t.Run("scatters to 3 workers and gathers 1 aggregated reply", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		reply := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		wA := messaging.NewPipelineChannel[int]()
		wB := messaging.NewPipelineChannel[int]()
		wC := messaging.NewPipelineChannel[int]()

		_ = startWorker(t, wA, reply, func(p int) int { return p + 1 })
		_ = startWorker(t, wB, reply, func(p int) int { return p + 2 })
		_ = startWorker(t, wC, reply, func(p int) int { return p + 3 })

		sinkHandler, getMsgs := captureMessages[int]()

		_, err := dst.Subscribe(sinkHandler)
		if err != nil {
			t.Fatalf("subscribe dst: %v", err)
		}

		selector := func(_ context.Context, _ messaging.Message[int]) ([]string, error) {
			return []string{"a", "b", "c"}, nil
		}

		sg := NewScatterGather("test", src, map[string]messaging.Channel[int]{"a": wA, "b": wB, "c": wC},
			reply, dst, selector, sumAggregate,
			WithGroupTimeout[int](200*time.Millisecond))

		err = sg.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), reqMsg(10, "req-1"))
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		got := getMsgs()
		if len(got) != 1 {
			t.Fatalf("expected 1 aggregated message, got %d", len(got))
		}

		// payload = (10+1) + (10+2) + (10+3) = 36.
		if got[0].Payload != 36 {
			t.Fatalf("expected payload 36, got %d", got[0].Payload)
		}

		if got[0].Headers.CorrelationID != "req-1" {
			t.Fatalf("expected CorrelationID req-1, got %q", got[0].Headers.CorrelationID)
		}
	})
}

func TestScatterGather_MultipleConcurrentScatters(t *testing.T) {
	t.Parallel()

	t.Run("5 distinct requests gather independently", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		reply := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		wA := messaging.NewPipelineChannel[int]()
		wB := messaging.NewPipelineChannel[int]()

		_ = startWorker(t, wA, reply, func(p int) int { return p })
		_ = startWorker(t, wB, reply, func(p int) int { return p * 2 })

		sinkHandler, getMsgs := captureMessages[int]()

		_, err := dst.Subscribe(sinkHandler)
		if err != nil {
			t.Fatalf("subscribe dst: %v", err)
		}

		selector := func(_ context.Context, _ messaging.Message[int]) ([]string, error) {
			return []string{"a", "b"}, nil
		}

		sg := NewScatterGather("test", src, map[string]messaging.Channel[int]{"a": wA, "b": wB},
			reply, dst, selector, sumAggregate,
			WithGroupTimeout[int](200*time.Millisecond))

		err = sg.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		for i := 1; i <= 5; i++ {
			err = src.Send(context.Background(), reqMsg(i, fmt.Sprintf("req-%d", i)))
			if err != nil {
				t.Fatalf("send req-%d: %v", i, err)
			}
		}

		got := getMsgs()
		if len(got) != 5 {
			t.Fatalf("expected 5 aggregated messages, got %d", len(got))
		}

		// Every aggregated result must carry one of the original
		// CorrelationIDs, with payload = i + 2*i = 3*i.
		seen := map[string]int{}
		for _, m := range got {
			seen[m.Headers.CorrelationID] = m.Payload
		}

		for i := 1; i <= 5; i++ {
			key := fmt.Sprintf("req-%d", i)

			want := 3 * i

			payload, ok := seen[key]
			if !ok {
				t.Fatalf("missing aggregated reply for %s", key)
			}

			if payload != want {
				t.Fatalf("%s: expected payload %d, got %d", key, want, payload)
			}
		}
	})
}

func TestScatterGather_PartialReplyTimeout(t *testing.T) {
	t.Parallel()

	t.Run("only 2 of 3 workers reply within timeout — DropHandler fires", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		reply := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		wA := messaging.NewPipelineChannel[int]()
		wB := messaging.NewPipelineChannel[int]()
		wSilent := messaging.NewPipelineChannel[int]()

		_ = startWorker(t, wA, reply, func(p int) int { return p })
		_ = startWorker(t, wB, reply, func(p int) int { return p })

		// wSilent: subscribe a no-op handler so Send succeeds but no reply ever lands.
		_, err := wSilent.Subscribe(func(_ context.Context, _ messaging.Message[int]) error {
			return nil
		})
		if err != nil {
			t.Fatalf("subscribe silent: %v", err)
		}

		sinkHandler, getMsgs := captureMessages[int]()

		_, err = dst.Subscribe(sinkHandler)
		if err != nil {
			t.Fatalf("subscribe dst: %v", err)
		}

		dropHandler, getDrops := captureDrops()

		selector := func(_ context.Context, _ messaging.Message[int]) ([]string, error) {
			return []string{"a", "b", "silent"}, nil
		}

		sg := NewScatterGather("test", src,
			map[string]messaging.Channel[int]{"a": wA, "b": wB, "silent": wSilent},
			reply, dst, selector, sumAggregate,
			WithGroupTimeout[int](50*time.Millisecond),
			WithDropHandler[int](dropHandler))

		err = sg.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), reqMsg(7, "stuck"))
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) {
			if getDrops() >= 1 {
				break
			}

			time.Sleep(10 * time.Millisecond)
		}

		if got := getDrops(); got < 1 {
			t.Fatalf("expected at least 1 drop fired for partial gather, got %d", got)
		}

		if got := getMsgs(); len(got) != 0 {
			t.Fatalf("expected 0 aggregated messages forwarded, got %d", len(got))
		}
	})
}

func TestScatterGather_EmptySelector(t *testing.T) {
	t.Parallel()

	t.Run("empty selector result fires DropHandler — no scatter", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		reply := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		worker := messaging.NewPipelineChannel[int]()

		workerCalls := int32(0)

		_, err := worker.Subscribe(func(_ context.Context, _ messaging.Message[int]) error {
			atomic.AddInt32(&workerCalls, 1)

			return nil
		})
		if err != nil {
			t.Fatalf("subscribe worker: %v", err)
		}

		selector := func(_ context.Context, _ messaging.Message[int]) ([]string, error) {
			return []string{}, nil
		}

		dropHandler, getDrops := captureDrops()
		errHandler, getErrs := captureErrors()

		sg := NewScatterGather("test", src, map[string]messaging.Channel[int]{"w": worker},
			reply, dst, selector, sumAggregate,
			WithGroupTimeout[int](200*time.Millisecond),
			WithDropHandler[int](dropHandler),
			WithErrorHandler[int](errHandler))

		err = sg.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), reqMsg(1, "empty"))
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got, want := getDrops(), int32(1); got != want {
			t.Fatalf("expected %d drops, got %d", want, got)
		}

		if got := atomic.LoadInt32(&workerCalls); got != 0 {
			t.Fatalf("worker should not be called on empty selector, got %d calls", got)
		}

		if errs := getErrs(); len(errs) != 0 {
			t.Fatalf("empty selector should not fire ErrorHandler, got %d errors: %v", len(errs), errs)
		}
	})
}

func TestScatterGather_UnknownWorkerKey(t *testing.T) {
	t.Parallel()

	t.Run("unknown selector key reports through ErrorHandler", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		reply := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		worker := messaging.NewPipelineChannel[int]()

		selector := func(_ context.Context, _ messaging.Message[int]) ([]string, error) {
			return []string{"unknown"}, nil
		}

		errHandler, getErrs := captureErrors()

		sg := NewScatterGather("test", src, map[string]messaging.Channel[int]{"w": worker},
			reply, dst, selector, sumAggregate,
			WithGroupTimeout[int](200*time.Millisecond),
			WithErrorHandler[int](errHandler))

		err := sg.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), reqMsg(1, "bad-key"))
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		errs := getErrs()
		if len(errs) == 0 {
			t.Fatal("expected error for unknown worker key")
		}

		if !errors.Is(errs[0], ErrScatterGatherFailed) {
			t.Fatalf("expected ErrScatterGatherFailed, got %v", errs[0])
		}

		if !errors.Is(errs[0], ErrScatterFailed) {
			t.Fatalf("expected ErrScatterFailed wrap, got %v", errs[0])
		}
	})
}

func TestScatterGather_OrphanCorrelationOnReply(t *testing.T) {
	t.Parallel()

	t.Run("worker reply with unknown correlation gathers but never completes — drops on timeout", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		reply := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		worker := messaging.NewPipelineChannel[int]()

		dropHandler, getDrops := captureDrops()
		errHandler, getErrs := captureErrors()

		selector := func(_ context.Context, _ messaging.Message[int]) ([]string, error) {
			return []string{"w"}, nil
		}

		sg := NewScatterGather("test", src, map[string]messaging.Channel[int]{"w": worker},
			reply, dst, selector, sumAggregate,
			WithGroupTimeout[int](50*time.Millisecond),
			WithDropHandler[int](dropHandler),
			WithErrorHandler[int](errHandler))

		err := sg.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		// Inject a reply on replyChan with a CorrelationID no scatter ever
		// produced. The Aggregator opens a group for it; without any
		// matching expected entry, completion never fires; the timeout
		// sweeper releases the partial group; our wrappedAggregate
		// returns errPartialDrop which routes to DropHandler.
		err = reply.Send(context.Background(), reqMsg(99, "ghost"))
		if err != nil {
			t.Fatalf("send ghost reply: %v", err)
		}

		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) {
			if getDrops() >= 1 {
				break
			}

			time.Sleep(10 * time.Millisecond)
		}

		if got := getDrops(); got < 1 {
			t.Fatalf("expected ghost reply to drop, got %d drops", got)
		}

		if errs := getErrs(); len(errs) != 0 {
			t.Fatalf("ghost reply should not fire ErrorHandler, got %d errors: %v", len(errs), errs)
		}
	})
}

func TestScatterGather_MaxConcurrentScatters(t *testing.T) {
	t.Parallel()

	t.Run("exceeding cap fires ErrorHandler with ErrMaxScattersExceeded", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		reply := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		// Worker that never replies — keeps the expected entry alive
		// so the cap is reached after the second send.
		worker := messaging.NewPipelineChannel[int]()

		_, err := worker.Subscribe(func(_ context.Context, _ messaging.Message[int]) error {
			return nil
		})
		if err != nil {
			t.Fatalf("subscribe worker: %v", err)
		}

		selector := func(_ context.Context, _ messaging.Message[int]) ([]string, error) {
			return []string{"w"}, nil
		}

		errHandler, getErrs := captureErrors()

		sg := NewScatterGather("test", src, map[string]messaging.Channel[int]{"w": worker},
			reply, dst, selector, sumAggregate,
			WithGroupTimeout[int](500*time.Millisecond),
			WithMaxConcurrentScatters[int](1),
			WithErrorHandler[int](errHandler))

		err = sg.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		// First request consumes the only slot.
		err = src.Send(context.Background(), reqMsg(1, "ok"))
		if err != nil {
			t.Fatalf("send ok: %v", err)
		}

		// Second request exceeds the cap.
		err = src.Send(context.Background(), reqMsg(2, "over"))
		if err != nil {
			t.Fatalf("send over: %v", err)
		}

		errs := getErrs()
		if len(errs) == 0 {
			t.Fatal("expected ErrMaxScattersExceeded report")
		}

		found := false
		for _, e := range errs {
			if errors.Is(e, ErrMaxScattersExceeded) {
				found = true
				break
			}
		}

		if !found {
			t.Fatalf("expected ErrMaxScattersExceeded in captured errors, got %v", errs)
		}
	})
}

func TestScatterGather_Stop(t *testing.T) {
	t.Parallel()

	t.Run("Stop drains in-flight gather without panic and closes Done", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		reply := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		worker := messaging.NewPipelineChannel[int]()

		_, err := worker.Subscribe(func(_ context.Context, _ messaging.Message[int]) error {
			return nil
		})
		if err != nil {
			t.Fatalf("subscribe worker: %v", err)
		}

		selector := func(_ context.Context, _ messaging.Message[int]) ([]string, error) {
			return []string{"w"}, nil
		}

		dropHandler, _ := captureDrops()

		sg := NewScatterGather("test", src, map[string]messaging.Channel[int]{"w": worker},
			reply, dst, selector, sumAggregate,
			WithGroupTimeout[int](500*time.Millisecond),
			WithDropHandler[int](dropHandler))

		err = sg.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		// In-flight request with no reply.
		err = src.Send(context.Background(), reqMsg(1, "pending"))
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		err = sg.Stop(context.Background())
		if err != nil {
			t.Fatalf("stop: %v", err)
		}

		select {
		case <-sg.Done():
		case <-time.After(2 * time.Second):
			t.Fatal("Done was not closed within timeout")
		}
	})

	t.Run("Stop is idempotent", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		reply := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		worker := messaging.NewPipelineChannel[int]()

		selector := func(_ context.Context, _ messaging.Message[int]) ([]string, error) {
			return []string{"w"}, nil
		}

		sg := NewScatterGather("test", src, map[string]messaging.Channel[int]{"w": worker},
			reply, dst, selector, sumAggregate,
			WithGroupTimeout[int](100*time.Millisecond))

		err := sg.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = sg.Stop(context.Background())
		if err != nil {
			t.Fatalf("first stop: %v", err)
		}

		err = sg.Stop(context.Background())
		if err != nil {
			t.Fatalf("second stop: %v", err)
		}
	})
}

func TestScatterGather_ConcurrentProducers(t *testing.T) {
	t.Parallel()

	t.Run("50 goroutines each issuing 1 request — all complete cleanly", func(t *testing.T) {
		t.Parallel()

		// Use TopicChannel for src so concurrent Send calls do not
		// serialize behind a Pipeline handler.
		src := messaging.NewTopicChannel[int]("scatter-src", messaging.WithBufferSize(200))
		reply := messaging.NewTopicChannel[int]("gather-reply", messaging.WithBufferSize(500))
		dst := messaging.NewTopicChannel[int]("gather-dst", messaging.WithBufferSize(200))

		wA := messaging.NewTopicChannel[int]("worker-a", messaging.WithBufferSize(200))
		wB := messaging.NewTopicChannel[int]("worker-b", messaging.WithBufferSize(200))

		// Start the topic channels.
		startTopic := func(t *testing.T, ch messaging.Channel[int]) {
			t.Helper()

			c, ok := ch.(interface {
				Start(context.Context) error
				Stop(context.Context) error
			})
			if !ok {
				t.Fatalf("channel does not implement lifecycle")
			}

			err := c.Start(context.Background())
			if err != nil {
				t.Fatalf("start topic: %v", err)
			}

			t.Cleanup(func() {
				stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				_ = c.Stop(stopCtx)
			})
		}

		startTopic(t, src)
		startTopic(t, reply)
		startTopic(t, dst)
		startTopic(t, wA)
		startTopic(t, wB)

		_ = startWorker(t, wA, reply, func(p int) int { return p })
		_ = startWorker(t, wB, reply, func(p int) int { return p * 2 })

		sinkHandler, getMsgs := captureMessages[int]()

		_, err := dst.Subscribe(sinkHandler)
		if err != nil {
			t.Fatalf("subscribe dst: %v", err)
		}

		selector := func(_ context.Context, _ messaging.Message[int]) ([]string, error) {
			return []string{"a", "b"}, nil
		}

		sg := NewScatterGather("test", src, map[string]messaging.Channel[int]{"a": wA, "b": wB},
			reply, dst, selector, sumAggregate,
			WithGroupTimeout[int](2*time.Second))

		err = sg.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() {
			stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_ = sg.Stop(stopCtx)
		})

		const n = 50

		var wg sync.WaitGroup

		for i := 1; i <= n; i++ {
			wg.Add(1)

			go func(idx int) {
				defer wg.Done()

				_ = src.Send(context.Background(), reqMsg(idx, fmt.Sprintf("req-%d", idx)))
			}(i)
		}

		wg.Wait()

		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			if len(getMsgs()) >= n {
				break
			}

			time.Sleep(20 * time.Millisecond)
		}

		got := getMsgs()
		if len(got) != n {
			t.Fatalf("expected %d aggregated messages, got %d", n, len(got))
		}

		// Verify uniqueness of correlation ids in the result.
		seen := map[string]struct{}{}
		for _, m := range got {
			seen[m.Headers.CorrelationID] = struct{}{}
		}

		if len(seen) != n {
			t.Fatalf("expected %d distinct CorrelationIDs in results, got %d", n, len(seen))
		}
	})
}

func TestScatterGather_ReplyWithEmptyCorrelation(t *testing.T) {
	t.Parallel()

	t.Run("reply with empty CorrelationID routes via aggregator drop", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		reply := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		worker := messaging.NewPipelineChannel[int]()

		dropHandler, getDrops := captureDrops()

		selector := func(_ context.Context, _ messaging.Message[int]) ([]string, error) {
			return []string{"w"}, nil
		}

		sg := NewScatterGather("test", src, map[string]messaging.Channel[int]{"w": worker},
			reply, dst, selector, sumAggregate,
			WithGroupTimeout[int](100*time.Millisecond),
			WithDropHandler[int](dropHandler))

		err := sg.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		// Send reply with empty CorrelationID — the internal Aggregator's
		// default CorrelationFn returns "" which routes through the
		// Aggregator's DropHandler (our aggregatorDrop).
		err = reply.Send(context.Background(), messaging.Message[int]{Payload: 99})
		if err != nil {
			t.Fatalf("send reply: %v", err)
		}

		if got, want := getDrops(), int32(1); got != want {
			t.Fatalf("expected %d drops, got %d", want, got)
		}
	})
}

func TestScatterGather_AggregateError(t *testing.T) {
	t.Parallel()

	t.Run("user AggregateFn error routes via ErrorHandler with ErrGatherFailed", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		reply := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		worker := messaging.NewPipelineChannel[int]()

		_ = startWorker(t, worker, reply, func(p int) int { return p })

		boom := errors.New("aggregate boom")

		failingAggregate := func(_ []messaging.Message[int]) (messaging.Message[int], error) {
			return messaging.Message[int]{}, boom
		}

		errHandler, getErrs := captureErrors()

		selector := func(_ context.Context, _ messaging.Message[int]) ([]string, error) {
			return []string{"w"}, nil
		}

		sg := NewScatterGather("test", src, map[string]messaging.Channel[int]{"w": worker},
			reply, dst, selector, failingAggregate,
			WithGroupTimeout[int](200*time.Millisecond),
			WithErrorHandler[int](errHandler))

		err := sg.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), reqMsg(1, "agg-fail"))
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		errs := getErrs()
		if len(errs) == 0 {
			t.Fatal("expected error from failing aggregate")
		}

		if !errors.Is(errs[0], ErrGatherFailed) {
			t.Fatalf("expected ErrGatherFailed, got %v", errs[0])
		}

		if !errors.Is(errs[0], boom) {
			t.Fatalf("expected wrapped aggregate error, got %v", errs[0])
		}
	})
}

func TestScatterGather_NilHandlersTolerated(t *testing.T) {
	t.Parallel()

	t.Run("default nil drop handler stays silent on empty selector", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		reply := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		worker := messaging.NewPipelineChannel[int]()

		selector := func(_ context.Context, _ messaging.Message[int]) ([]string, error) {
			return []string{}, nil
		}

		// No WithDropHandler — recipientListDrop runs with sg.dropHandler == nil.
		sg := NewScatterGather("test", src, map[string]messaging.Channel[int]{"w": worker},
			reply, dst, selector, sumAggregate,
			WithGroupTimeout[int](100*time.Millisecond))

		err := sg.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), reqMsg(1, "silent"))
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		// Stop should still close Done cleanly.
		err = sg.Stop(context.Background())
		if err != nil {
			t.Fatalf("stop: %v", err)
		}

		select {
		case <-sg.Done():
		case <-time.After(time.Second):
			t.Fatal("Done not closed")
		}
	})

	t.Run("default nil drop handler stays silent on empty-correlation reply", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		reply := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		worker := messaging.NewPipelineChannel[int]()

		selector := func(_ context.Context, _ messaging.Message[int]) ([]string, error) {
			return []string{"w"}, nil
		}

		sg := NewScatterGather("test", src, map[string]messaging.Channel[int]{"w": worker},
			reply, dst, selector, sumAggregate,
			WithGroupTimeout[int](100*time.Millisecond))

		err := sg.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = reply.Send(context.Background(), messaging.Message[int]{Payload: 99})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		err = sg.Stop(context.Background())
		if err != nil {
			t.Fatalf("stop: %v", err)
		}
	})

	t.Run("orphan sweeper does not panic when dropHandler is nil", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		reply := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		worker := messaging.NewPipelineChannel[int]()

		_, err := worker.Subscribe(func(_ context.Context, _ messaging.Message[int]) error {
			return nil
		})
		if err != nil {
			t.Fatalf("subscribe worker: %v", err)
		}

		selector := func(_ context.Context, _ messaging.Message[int]) ([]string, error) {
			return []string{"w"}, nil
		}

		sg := NewScatterGather("test", src, map[string]messaging.Channel[int]{"w": worker},
			reply, dst, selector, sumAggregate,
			WithGroupTimeout[int](20*time.Millisecond))

		err = sg.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), reqMsg(1, "orphan"))
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		// Wait long enough for orphan sweep to fire (~2 × tick + ttl).
		time.Sleep(150 * time.Millisecond)

		err = sg.Stop(context.Background())
		if err != nil {
			t.Fatalf("stop: %v", err)
		}
	})
}

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("applies all defaults when no options passed", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions[int]()
		if opts.errorHandler == nil {
			t.Fatal("default error handler should be installed")
		}

		if opts.maxConcurrentScatters != defaultMaxConcurrentScatters {
			t.Fatalf("expected default cap %d, got %d", defaultMaxConcurrentScatters, opts.maxConcurrentScatters)
		}

		if opts.dropHandler != nil {
			t.Fatal("default drop handler should be nil (silent)")
		}

		if opts.groupTimeout != 0 {
			t.Fatalf("default groupTimeout should be zero, got %v", opts.groupTimeout)
		}
	})

	t.Run("WithGroupTimeout positive value installs duration", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithGroupTimeout[int](250 * time.Millisecond))
		if opts.groupTimeout != 250*time.Millisecond {
			t.Fatalf("expected 250ms, got %v", opts.groupTimeout)
		}
	})

	t.Run("WithGroupTimeout zero is ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithGroupTimeout[int](0))
		if opts.groupTimeout != 0 {
			t.Fatalf("zero duration should be ignored, got %v", opts.groupTimeout)
		}
	})

	t.Run("WithMaxConcurrentScatters non-positive is ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMaxConcurrentScatters[int](0))
		if opts.maxConcurrentScatters != defaultMaxConcurrentScatters {
			t.Fatalf("zero should be ignored; default kept, got %d", opts.maxConcurrentScatters)
		}
	})
}

func TestScatterGather_StartIdempotent(t *testing.T) {
	t.Parallel()

	t.Run("second Start returns nil without re-subscribing", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		reply := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		worker := messaging.NewPipelineChannel[int]()

		selector := func(_ context.Context, _ messaging.Message[int]) ([]string, error) {
			return []string{"w"}, nil
		}

		sg := NewScatterGather("test", src, map[string]messaging.Channel[int]{"w": worker},
			reply, dst, selector, sumAggregate,
			WithGroupTimeout[int](100*time.Millisecond))

		err := sg.Start(context.Background())
		if err != nil {
			t.Fatalf("first start: %v", err)
		}

		err = sg.Start(context.Background())
		if err != nil {
			t.Fatalf("second start: %v", err)
		}

		err = sg.Stop(context.Background())
		if err != nil {
			t.Fatalf("stop: %v", err)
		}
	})
}

func TestScatterGather_SelectorPanic(t *testing.T) {
	t.Parallel()

	t.Run("user selector panic routes through ErrorHandler as ErrScatterFailed", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		reply := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		worker := messaging.NewPipelineChannel[int]()

		selector := func(_ context.Context, _ messaging.Message[int]) ([]string, error) {
			panic("boom selector")
		}

		errHandler, getErrs := captureErrors()

		sg := NewScatterGather("test", src, map[string]messaging.Channel[int]{"w": worker},
			reply, dst, selector, sumAggregate,
			WithGroupTimeout[int](100*time.Millisecond),
			WithErrorHandler[int](errHandler))

		err := sg.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), reqMsg(1, "panicky"))
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		errs := getErrs()
		if len(errs) == 0 {
			t.Fatal("expected error from selector panic")
		}

		if !errors.Is(errs[0], ErrScatterFailed) {
			t.Fatalf("expected ErrScatterFailed wrap, got %v", errs[0])
		}
	})
}

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("includes type prefix", func(t *testing.T) {
		t.Parallel()

		err := ErrScatterGather(ErrMaxScattersExceeded)

		var domainErr *Error
		if !errors.As(err, &domainErr) {
			t.Fatal("expected *Error from ErrScatterGather")
		}

		got := domainErr.Error()
		if got == "" {
			t.Fatal("expected non-empty Error string")
		}

		// must include the scattergather type token
		if !contains(got, ScatterGatherType) {
			t.Fatalf("expected Error to contain %q, got %q", ScatterGatherType, got)
		}
	})
}

// contains is a tiny substring helper to avoid importing strings.
func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}

	return false
}

func TestErrScatterGather(t *testing.T) {
	t.Parallel()

	t.Run("wraps causes with ErrScatterGatherFailed", func(t *testing.T) {
		t.Parallel()

		err := ErrScatterGather(ErrMaxScattersExceeded)
		if !errors.Is(err, ErrScatterGatherFailed) {
			t.Fatal("expected ErrScatterGatherFailed in chain")
		}

		if !errors.Is(err, ErrMaxScattersExceeded) {
			t.Fatal("expected wrapped cause in chain")
		}
	})

	t.Run("Error string includes type", func(t *testing.T) {
		t.Parallel()

		var domainErr *Error

		err := ErrScatterGather(ErrScatterFailed)
		if !errors.As(err, &domainErr) {
			t.Fatal("expected *Error from ErrScatterGather")
		}

		if got := domainErr.Type; got != ScatterGatherType {
			t.Fatalf("expected Type %q, got %q", ScatterGatherType, got)
		}
	})
}
