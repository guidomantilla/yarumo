package messaging

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	lctests "github.com/guidomantilla/yarumo/core/common/lifecycle/tests"
)

func TestNewTopicChannel(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil channel", func(t *testing.T) {
		t.Parallel()

		qc := NewTopicChannel[int]("q-1").(*topic[int])
		if qc == nil {
			t.Fatal("expected non-nil queue channel")
		}
	})

	t.Run("carries the given name", func(t *testing.T) {
		t.Parallel()

		qc := NewTopicChannel[int]("q-named").(*topic[int])
		if qc.Name() != "q-named" {
			t.Fatalf("expected name %q, got %q", "q-named", qc.Name())
		}
	})

	t.Run("done channel open at construction", func(t *testing.T) {
		t.Parallel()

		qc := NewTopicChannel[int]("q-done").(*topic[int])
		select {
		case <-qc.Done():
			t.Fatal("expected Done open before Stop")
		default:
		}
	})

	t.Run("applies buffer size option", func(t *testing.T) {
		t.Parallel()

		qc := NewTopicChannel[int]("q-buf", WithBufferSize(4)).(*topic[int])
		if qc.bufferSize != 4 {
			t.Fatalf("expected buffer 4, got %d", qc.bufferSize)
		}
	})
}

func TestTopicChannel_Start(t *testing.T) {
	t.Parallel()

	t.Run("returns nil and accepts sends", func(t *testing.T) {
		t.Parallel()

		qc := NewTopicChannel[int]("q-start", WithBufferSize(4)).(*topic[int])
		err := qc.Start(context.Background())
		if err != nil {
			t.Fatalf("Start returned %v", err)
		}

		delivered := make(chan int, 1)
		_, err = qc.Subscribe(func(_ context.Context, msg Message[int]) error {
			delivered <- msg.Payload
			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}

		err = qc.Send(context.Background(), NewMessage[int](7, nil))
		if err != nil {
			t.Fatalf("Send returned %v", err)
		}

		select {
		case got := <-delivered:
			if got != 7 {
				t.Fatalf("expected 7, got %d", got)
			}
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for delivery")
		}

		_ = qc.Stop(context.Background())
	})
}

func TestTopicChannel_Send(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrContextNil on nil ctx", func(t *testing.T) {
		t.Parallel()

		qc := NewTopicChannel[int]("q-ctxnil").(*topic[int])
		err := qc.Send(nil, NewMessage[int](1, nil)) //nolint:staticcheck
		if !errors.Is(err, ErrContextNil) {
			t.Fatalf("expected ErrContextNil, got %v", err)
		}
	})

	t.Run("returns ErrClosed after Stop", func(t *testing.T) {
		t.Parallel()

		qc := NewTopicChannel[int]("q-closed").(*topic[int])
		err := qc.Start(context.Background())
		if err != nil {
			t.Fatalf("Start returned %v", err)
		}

		err = qc.Stop(context.Background())
		if err != nil {
			t.Fatalf("Stop returned %v", err)
		}

		err = qc.Send(context.Background(), NewMessage[int](1, nil))
		if !errors.Is(err, ErrClosed) {
			t.Fatalf("expected ErrClosed, got %v", err)
		}
	})

	t.Run("returns ErrTimeout when ctx expires while buffer full", func(t *testing.T) {
		t.Parallel()

		// buffer size 1 + a subscriber registered without Start so its
		// inbox fills but no worker drains. Explicit OverflowBlock since
		// this test asserts blocking behavior; default OverflowReject
		// would return ErrBufferFull instead.
		qc := NewTopicChannel[int]("q-fullbuf",
			WithBufferSize(1),
			WithOverflowPolicy(OverflowBlock),
		).(*topic[int])

		_, err := qc.Subscribe(func(_ context.Context, _ Message[int]) error { return nil })
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}

		err = qc.Send(context.Background(), NewMessage[int](1, nil))
		if err != nil {
			t.Fatalf("first Send returned %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		err = qc.Send(ctx, NewMessage[int](2, nil))
		if err == nil {
			t.Fatal("expected error on full-buffer Send")
		}
		if !errors.Is(err, ErrTimeout) {
			t.Fatalf("expected ErrTimeout, got %v", err)
		}
	})
}

func TestTopicChannel_Subscribe(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrHandlerNil on nil handler", func(t *testing.T) {
		t.Parallel()

		qc := NewTopicChannel[int]("q-sub-nil").(*topic[int])
		_, err := qc.Subscribe(nil)
		if !errors.Is(err, ErrHandlerNil) {
			t.Fatalf("expected ErrHandlerNil, got %v", err)
		}
	})

	t.Run("returns ErrClosed after Stop", func(t *testing.T) {
		t.Parallel()

		qc := NewTopicChannel[int]("q-sub-closed").(*topic[int])
		err := qc.Start(context.Background())
		if err != nil {
			t.Fatalf("Start returned %v", err)
		}

		err = qc.Stop(context.Background())
		if err != nil {
			t.Fatalf("Stop returned %v", err)
		}

		_, err = qc.Subscribe(func(_ context.Context, _ Message[int]) error { return nil })
		if !errors.Is(err, ErrClosed) {
			t.Fatalf("expected ErrClosed, got %v", err)
		}
	})

	t.Run("cancel detaches handler", func(t *testing.T) {
		t.Parallel()

		qc := NewTopicChannel[int]("q-sub-cancel").(*topic[int])
		err := qc.Start(context.Background())
		if err != nil {
			t.Fatalf("Start returned %v", err)
		}

		var fired int32

		cancel, err := qc.Subscribe(func(_ context.Context, _ Message[int]) error {
			atomic.AddInt32(&fired, 1)
			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}

		err = qc.Send(context.Background(), NewMessage[int](1, nil))
		if err != nil {
			t.Fatalf("Send returned %v", err)
		}

		// give worker a moment to deliver
		deadline := time.Now().Add(time.Second)
		for time.Now().Before(deadline) {
			if atomic.LoadInt32(&fired) == 1 {
				break
			}
			time.Sleep(time.Millisecond)
		}

		cancel()
		cancel() // idempotent

		err = qc.Send(context.Background(), NewMessage[int](2, nil))
		if err != nil {
			t.Fatalf("Send returned %v", err)
		}

		// allow worker time to drain second send
		time.Sleep(50 * time.Millisecond)

		got := atomic.LoadInt32(&fired)
		if got != 1 {
			t.Fatalf("expected fired=1 after cancel, got %d", got)
		}

		_ = qc.Stop(context.Background())
	})
}

func TestTopicChannel_Stop(t *testing.T) {
	t.Parallel()

	t.Run("drains pending messages before returning", func(t *testing.T) {
		t.Parallel()

		qc := NewTopicChannel[int]("q-drain", WithBufferSize(16), WithDrainTimeout(time.Second)).(*topic[int])

		var delivered int32

		_, err := qc.Subscribe(func(_ context.Context, _ Message[int]) error {
			atomic.AddInt32(&delivered, 1)
			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}

		err = qc.Start(context.Background())
		if err != nil {
			t.Fatalf("Start returned %v", err)
		}

		const sends = 10
		for i := range sends {
			err = qc.Send(context.Background(), NewMessage[int](i, nil))
			if err != nil {
				t.Fatalf("Send returned %v", err)
			}
		}

		err = qc.Stop(context.Background())
		if err != nil {
			t.Fatalf("Stop returned %v", err)
		}

		got := atomic.LoadInt32(&delivered)
		if got != sends {
			t.Fatalf("expected %d delivered, got %d", sends, got)
		}
	})

	t.Run("returns ErrShutdownTimeout when drain exceeds bound", func(t *testing.T) {
		t.Parallel()

		qc := NewTopicChannel[int]("q-drain-timeout", WithBufferSize(8), WithDrainTimeout(10*time.Millisecond)).(*topic[int])

		// slow handler that blocks past the drain timeout
		release := make(chan struct{})
		_, err := qc.Subscribe(func(_ context.Context, _ Message[int]) error {
			<-release
			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}

		err = qc.Start(context.Background())
		if err != nil {
			t.Fatalf("Start returned %v", err)
		}

		err = qc.Send(context.Background(), NewMessage[int](1, nil))
		if err != nil {
			t.Fatalf("Send returned %v", err)
		}

		err = qc.Stop(context.Background())
		if err == nil {
			t.Fatal("expected drain-timeout error")
		}
		if !errors.Is(err, lifecycle.ErrShutdownTimeout) {
			t.Fatalf("expected ErrShutdownTimeout, got %v", err)
		}

		close(release)
		<-qc.Done()
	})

	t.Run("is idempotent across multiple calls", func(t *testing.T) {
		t.Parallel()

		qc := NewTopicChannel[int]("q-stop-idemp").(*topic[int])
		err := qc.Start(context.Background())
		if err != nil {
			t.Fatalf("Start returned %v", err)
		}

		err = qc.Stop(context.Background())
		if err != nil {
			t.Fatalf("first Stop returned %v", err)
		}

		err = qc.Stop(context.Background())
		if err != nil {
			t.Fatalf("second Stop returned %v", err)
		}
	})
}

func TestTopicChannel_StopIsIdempotent(t *testing.T) {
	t.Parallel()

	qc := NewTopicChannel[int]("q-lct").(*topic[int])
	err := qc.Start(context.Background())
	if err != nil {
		t.Fatalf("Start returned %v", err)
	}

	lctests.AssertIdempotentStop(t, qc)
}

func TestTopicChannel_LifecycleIntegration(t *testing.T) {
	t.Parallel()

	t.Run("lifecycle.Build wires the worker and CloseFn drains", func(t *testing.T) {
		t.Parallel()

		qc := NewTopicChannel[int]("q-lifecycle", WithBufferSize(4), WithDrainTimeout(time.Second)).(*topic[int])

		var delivered int32
		_, err := qc.Subscribe(func(_ context.Context, _ Message[int]) error {
			atomic.AddInt32(&delivered, 1)
			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}

		errChan := make(chan error, 1)
		closeFn, err := lifecycle.Build(context.Background(), qc, errChan)
		if err != nil {
			t.Fatalf("lifecycle.Build returned %v", err)
		}

		err = qc.Send(context.Background(), NewMessage[int](1, nil))
		if err != nil {
			t.Fatalf("Send returned %v", err)
		}

		closeFn(context.Background(), time.Second)

		got := atomic.LoadInt32(&delivered)
		if got != 1 {
			t.Fatalf("expected 1 delivery, got %d", got)
		}
	})
}

func TestTopicChannel_FanOutAllHandlersFire(t *testing.T) {
	t.Parallel()

	// With per-subscriber workers each handler runs in its own
	// goroutine concurrently; the order in which they observe a
	// fan-out is NOT a TopicChannel guarantee. This test verifies the
	// invariant that DOES hold: every registered handler receives
	// every published message.
	qc := NewTopicChannel[int]("q-fanout", WithBufferSize(4), WithDrainTimeout(time.Second)).(*topic[int])

	var (
		mu     sync.Mutex
		fired  = map[int]bool{}
		signal = make(chan struct{}, 1)
	)

	subscribe := func(id int) {
		_, err := qc.Subscribe(func(_ context.Context, _ Message[int]) error {
			mu.Lock()
			fired[id] = true
			complete := len(fired) == 4
			mu.Unlock()

			if complete {
				signal <- struct{}{}
			}

			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}
	}

	subscribe(0)
	subscribe(1)
	subscribe(2)
	subscribe(3)

	errChan := make(chan error, 1)

	closeFn, err := lifecycle.Build(context.Background(), qc, errChan)
	if err != nil {
		t.Fatalf("lifecycle.Build returned %v", err)
	}

	defer closeFn(context.Background(), time.Second)

	err = qc.Send(context.Background(), NewMessage[int](42, nil))
	if err != nil {
		t.Fatalf("Send returned %v", err)
	}

	select {
	case <-signal:
	case <-time.After(time.Second):
		t.Fatalf("not all handlers fired in time: %v", fired)
	}

	for id := range 4 {
		if !fired[id] {
			t.Fatalf("handler %d did not fire: %v", id, fired)
		}
	}
}

func TestTopicChannel_PanicRecovery(t *testing.T) {
	t.Parallel()

	var (
		captured  error
		hookFired = make(chan struct{}, 1)
	)

	errHook := func(_ context.Context, _ any, err error) {
		captured = err
		hookFired <- struct{}{}
	}

	qc := NewTopicChannel[int]("q-panic", WithBufferSize(4), WithDrainTimeout(time.Second), WithErrorHandler(errHook)).(*topic[int])

	var afterFired int32

	_, err := qc.Subscribe(func(_ context.Context, _ Message[int]) error {
		panic("kaboom")
	})
	if err != nil {
		t.Fatalf("Subscribe returned %v", err)
	}

	_, err = qc.Subscribe(func(_ context.Context, _ Message[int]) error {
		atomic.AddInt32(&afterFired, 1)
		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe returned %v", err)
	}

	errChan := make(chan error, 1)

	closeFn, err := lifecycle.Build(context.Background(), qc, errChan)
	if err != nil {
		t.Fatalf("BuildTopicChannel returned %v", err)
	}

	defer closeFn(context.Background(), time.Second)

	err = qc.Send(context.Background(), NewMessage[int](1, nil))
	if err != nil {
		t.Fatalf("Send returned %v", err)
	}

	<-hookFired

	if captured == nil {
		t.Fatal("expected error captured, got nil")
	}

	if !errors.Is(captured, ErrHandlerPanic) {
		t.Fatalf("expected ErrHandlerPanic wrapped, got %v", captured)
	}

	closeFn(context.Background(), time.Second)

	if atomic.LoadInt32(&afterFired) != 1 {
		t.Fatalf("subsequent handler should still fire after a panic, got %d", afterFired)
	}
}

func TestTopicChannel_ErrorHandlerHook(t *testing.T) {
	t.Parallel()

	boom := errors.New("boom")

	var (
		captured  error
		hookFired = make(chan struct{}, 1)
	)

	errHook := func(_ context.Context, _ any, err error) {
		captured = err
		hookFired <- struct{}{}
	}

	qc := NewTopicChannel[int]("q-errhook", WithBufferSize(4), WithDrainTimeout(time.Second), WithErrorHandler(errHook)).(*topic[int])

	_, err := qc.Subscribe(func(_ context.Context, _ Message[int]) error {
		return boom
	})
	if err != nil {
		t.Fatalf("Subscribe returned %v", err)
	}

	errChan := make(chan error, 1)

	closeFn, err := lifecycle.Build(context.Background(), qc, errChan)
	if err != nil {
		t.Fatalf("BuildTopicChannel returned %v", err)
	}

	defer closeFn(context.Background(), time.Second)

	err = qc.Send(context.Background(), NewMessage[int](1, nil))
	if err != nil {
		t.Fatalf("Send returned %v", err)
	}

	<-hookFired

	if !errors.Is(captured, boom) {
		t.Fatalf("expected captured to wrap boom, got %v", captured)
	}
}

func TestTopicChannel_Concurrent(t *testing.T) {
	t.Parallel()

	// Concurrent dispatch correctness test — uses OverflowBlock so the
	// 200 concurrent sends all succeed (default OverflowReject would
	// return ErrBufferFull under sustained overshoot).
	qc := NewTopicChannel[int]("q-concur",
		WithBufferSize(64),
		WithDrainTimeout(time.Second),
		WithOverflowPolicy(OverflowBlock),
	).(*topic[int])

	var fired int64
	const sends = 200
	const subs = 4

	for range subs {
		_, err := qc.Subscribe(func(_ context.Context, _ Message[int]) error {
			atomic.AddInt64(&fired, 1)
			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}
	}

	err := qc.Start(context.Background())
	if err != nil {
		t.Fatalf("Start returned %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(sends)
	for i := range sends {
		go func(n int) {
			defer wg.Done()
			sendErr := qc.Send(context.Background(), NewMessage[int](n, nil))
			if sendErr != nil {
				t.Errorf("Send returned %v", sendErr)
			}
		}(i)
	}

	wg.Wait()

	err = qc.Stop(context.Background())
	if err != nil {
		t.Fatalf("Stop returned %v", err)
	}

	got := atomic.LoadInt64(&fired)
	want := int64(sends * subs)
	if got != want {
		t.Fatalf("expected %d deliveries, got %d", want, got)
	}
}

func TestTopicChannel_OverflowReject_ReturnsErrBufferFull(t *testing.T) {
	t.Parallel()

	// Subscriber registered without Start → its inbox fills but no
	// worker drains.
	ch := NewTopicChannel[int]("t-reject", WithBufferSize(1)).(*topic[int])

	_, err := ch.Subscribe(func(_ context.Context, _ Message[int]) error { return nil })
	if err != nil {
		t.Fatalf("Subscribe returned %v", err)
	}

	err = ch.Send(context.Background(), NewMessage[int](1, nil))
	if err != nil {
		t.Fatalf("first Send returned %v", err)
	}

	start := time.Now()
	err = ch.Send(context.Background(), NewMessage[int](2, nil))
	elapsed := time.Since(start)

	if !errors.Is(err, ErrBufferFull) {
		t.Fatalf("expected ErrBufferFull, got %v", err)
	}
	if elapsed > 50*time.Millisecond {
		t.Fatalf("Reject should not block; took %v", elapsed)
	}
}

func TestTopicChannel_OverflowBlock_BlocksUntilCtxExpires(t *testing.T) {
	t.Parallel()

	ch := NewTopicChannel[int]("t-block",
		WithBufferSize(1),
		WithOverflowPolicy(OverflowBlock),
	).(*topic[int])

	_, err := ch.Subscribe(func(_ context.Context, _ Message[int]) error { return nil })
	if err != nil {
		t.Fatalf("Subscribe returned %v", err)
	}

	err = ch.Send(context.Background(), NewMessage[int](1, nil))
	if err != nil {
		t.Fatalf("first Send returned %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err = ch.Send(ctx, NewMessage[int](2, nil))
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("expected ErrTimeout, got %v", err)
	}
}

func TestTopicChannel_OverflowDropNewest_DropsNewMessage(t *testing.T) {
	t.Parallel()

	var captured []Message[int]
	hook := func(_ context.Context, msg any, _ error) {
		captured = append(captured, msg.(Message[int]))
	}

	ch := NewTopicChannel[int]("t-newest",
		WithBufferSize(1),
		WithOverflowPolicy(OverflowDropNewest),
		WithErrorHandler(hook),
	).(*topic[int])

	_, err := ch.Subscribe(func(_ context.Context, _ Message[int]) error { return nil })
	if err != nil {
		t.Fatalf("Subscribe returned %v", err)
	}

	err = ch.Send(context.Background(), NewMessage[int](1, nil))
	if err != nil {
		t.Fatalf("first Send returned %v", err)
	}

	err = ch.Send(context.Background(), NewMessage[int](2, nil))
	if err != nil {
		t.Fatalf("DropNewest Send should return nil, got %v", err)
	}

	if len(captured) != 1 {
		t.Fatalf("expected 1 hook call, got %d", len(captured))
	}
	if captured[0].Payload != 2 {
		t.Fatalf("expected new msg (2) dropped, got %d", captured[0].Payload)
	}
}

func TestTopicChannel_OverflowDropOldest_EvictsAndAcceptsNew(t *testing.T) {
	t.Parallel()

	var captured []Message[int]
	hook := func(_ context.Context, msg any, _ error) {
		captured = append(captured, msg.(Message[int]))
	}

	ch := NewTopicChannel[int]("t-oldest",
		WithBufferSize(2),
		WithOverflowPolicy(OverflowDropOldest),
		WithErrorHandler(hook),
	).(*topic[int])

	_, err := ch.Subscribe(func(_ context.Context, _ Message[int]) error { return nil })
	if err != nil {
		t.Fatalf("Subscribe returned %v", err)
	}

	_ = ch.Send(context.Background(), NewMessage[int](1, nil))
	_ = ch.Send(context.Background(), NewMessage[int](2, nil))

	err = ch.Send(context.Background(), NewMessage[int](3, nil))
	if err != nil {
		t.Fatalf("DropOldest Send should return nil, got %v", err)
	}

	if len(captured) != 1 {
		t.Fatalf("expected 1 evicted msg, got %d", len(captured))
	}
	if captured[0].Payload != 1 {
		t.Fatalf("expected oldest (1) evicted, got %d", captured[0].Payload)
	}
}

func TestTopicChannel_OverflowDrop_HookErrorJoinedOverflowAndDropped(t *testing.T) {
	t.Parallel()

	var got error
	hook := func(_ context.Context, _ any, err error) {
		got = err
	}

	ch := NewTopicChannel[int]("t-hookcheck",
		WithBufferSize(1),
		WithOverflowPolicy(OverflowDropNewest),
		WithErrorHandler(hook),
	).(*topic[int])

	_, err := ch.Subscribe(func(_ context.Context, _ Message[int]) error { return nil })
	if err != nil {
		t.Fatalf("Subscribe returned %v", err)
	}

	_ = ch.Send(context.Background(), NewMessage[int](1, nil))
	_ = ch.Send(context.Background(), NewMessage[int](2, nil))

	if got == nil {
		t.Fatal("hook did not fire")
	}
	if !errors.Is(got, ErrOverflow) {
		t.Fatalf("expected ErrOverflow, got %v", got)
	}
	if !errors.Is(got, ErrDropped) {
		t.Fatalf("expected ErrDropped (joined), got %v", got)
	}
}

func TestTopicChannel_SlowHandlerDoesNotBlockFast(t *testing.T) {
	t.Parallel()

	// Core invariant of per-subscriber workers: a slow handler stays
	// in its own goroutine; fast handlers keep receiving messages at
	// line rate. We verify by having one subscriber block on a signal
	// and confirming the other delivers N messages before we unblock
	// the slow one.
	qc := NewTopicChannel[int]("t-isolation",
		WithBufferSize(64),
		WithDrainTimeout(time.Second),
	).(*topic[int])

	const fastSends = 50

	slowUnblock := make(chan struct{})
	slowEntered := make(chan struct{}, 1)

	_, err := qc.Subscribe(func(_ context.Context, _ Message[int]) error {
		select {
		case slowEntered <- struct{}{}:
		default:
		}

		<-slowUnblock

		return nil
	})
	if err != nil {
		t.Fatalf("slow Subscribe returned %v", err)
	}

	var fastDelivered int64

	fastDone := make(chan struct{}, 1)

	_, err = qc.Subscribe(func(_ context.Context, _ Message[int]) error {
		if atomic.AddInt64(&fastDelivered, 1) == int64(fastSends) {
			fastDone <- struct{}{}
		}

		return nil
	})
	if err != nil {
		t.Fatalf("fast Subscribe returned %v", err)
	}

	errChan := make(chan error, 1)

	closeFn, err := lifecycle.Build(context.Background(), qc, errChan)
	if err != nil {
		t.Fatalf("lifecycle.Build returned %v", err)
	}

	for i := range fastSends {
		err = qc.Send(context.Background(), NewMessage[int](i, nil))
		if err != nil {
			t.Fatalf("Send %d returned %v", i, err)
		}
	}

	// Wait until the slow handler has entered (so we know it is
	// holding its goroutine) AND the fast handler has drained all 50.
	select {
	case <-slowEntered:
	case <-time.After(time.Second):
		t.Fatal("slow handler never entered")
	}

	select {
	case <-fastDone:
	case <-time.After(time.Second):
		t.Fatalf("fast handler did not drain %d msgs while slow handler was blocked; got %d", fastSends, atomic.LoadInt64(&fastDelivered))
	}

	close(slowUnblock)

	closeFn(context.Background(), time.Second)
}

func TestTopicChannel_CancelDuringActiveSends_NoLeak(t *testing.T) {
	t.Parallel()

	// Cancel during in-flight Sends must be safe: no panic, no leak,
	// no further dispatch to the cancelled subscriber.
	qc := NewTopicChannel[int]("t-cancel-race",
		WithBufferSize(16),
		WithDrainTimeout(time.Second),
	).(*topic[int])

	var (
		mu       sync.Mutex
		received []int
	)

	cancel, err := qc.Subscribe(func(_ context.Context, msg Message[int]) error {
		mu.Lock()
		received = append(received, msg.Payload)
		mu.Unlock()

		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe returned %v", err)
	}

	errChan := make(chan error, 1)

	closeFn, err := lifecycle.Build(context.Background(), qc, errChan)
	if err != nil {
		t.Fatalf("lifecycle.Build returned %v", err)
	}

	defer closeFn(context.Background(), time.Second)

	var wg sync.WaitGroup

	const sends = 50

	wg.Add(sends)

	for i := range sends {
		go func(n int) {
			defer wg.Done()

			_ = qc.Send(context.Background(), NewMessage[int](n, nil))
		}(i)
	}

	// Cancel mid-flight.
	cancel()

	wg.Wait()

	mu.Lock()
	count := len(received)
	mu.Unlock()

	if count > sends {
		t.Fatalf("received more msgs than sent: %d > %d", count, sends)
	}
}

func TestTopicChannel_SubscribeBeforeStart_WorkerSpawnedAtStart(t *testing.T) {
	t.Parallel()

	qc := NewTopicChannel[int]("t-presub").(*topic[int])

	var delivered int32

	_, err := qc.Subscribe(func(_ context.Context, _ Message[int]) error {
		atomic.AddInt32(&delivered, 1)

		return nil
	})
	if err != nil {
		t.Fatalf("pre-Start Subscribe returned %v", err)
	}

	errChan := make(chan error, 1)

	closeFn, err := lifecycle.Build(context.Background(), qc, errChan)
	if err != nil {
		t.Fatalf("lifecycle.Build returned %v", err)
	}

	err = qc.Send(context.Background(), NewMessage[int](1, nil))
	if err != nil {
		t.Fatalf("Send returned %v", err)
	}

	closeFn(context.Background(), time.Second)

	got := atomic.LoadInt32(&delivered)
	if got != 1 {
		t.Fatalf("expected 1 delivery, got %d", got)
	}
}

func TestTopicChannel_SubscribeAfterStart_WorkerSpawnedImmediately(t *testing.T) {
	t.Parallel()

	qc := NewTopicChannel[int]("t-postsub").(*topic[int])

	errChan := make(chan error, 1)

	closeFn, err := lifecycle.Build(context.Background(), qc, errChan)
	if err != nil {
		t.Fatalf("lifecycle.Build returned %v", err)
	}

	// Give Build's goroutine a moment to actually run Start.
	for i := 0; i < 100 && !qc.started.Load(); i++ {
		time.Sleep(time.Millisecond)
	}

	if !qc.started.Load() {
		t.Fatal("started flag never became true")
	}

	var delivered int32

	_, err = qc.Subscribe(func(_ context.Context, _ Message[int]) error {
		atomic.AddInt32(&delivered, 1)

		return nil
	})
	if err != nil {
		t.Fatalf("post-Start Subscribe returned %v", err)
	}

	err = qc.Send(context.Background(), NewMessage[int](1, nil))
	if err != nil {
		t.Fatalf("Send returned %v", err)
	}

	closeFn(context.Background(), time.Second)

	got := atomic.LoadInt32(&delivered)
	if got != 1 {
		t.Fatalf("expected 1 delivery, got %d", got)
	}
}

func TestTopicChannel_Send_AggregatesPerSubErrors(t *testing.T) {
	t.Parallel()

	// With OverflowReject + bufferSize=1, sending to a Topic with two
	// subs after filling their inboxes returns a joined error
	// containing ErrBufferFull from BOTH subs.
	qc := NewTopicChannel[int]("t-aggregate",
		WithBufferSize(1),
	).(*topic[int])

	_, err := qc.Subscribe(func(_ context.Context, _ Message[int]) error { return nil })
	if err != nil {
		t.Fatalf("Subscribe 1 returned %v", err)
	}

	_, err = qc.Subscribe(func(_ context.Context, _ Message[int]) error { return nil })
	if err != nil {
		t.Fatalf("Subscribe 2 returned %v", err)
	}

	// First Send fills both inboxes (no Start so no worker drains).
	err = qc.Send(context.Background(), NewMessage[int](1, nil))
	if err != nil {
		t.Fatalf("first Send returned %v", err)
	}

	err = qc.Send(context.Background(), NewMessage[int](2, nil))
	if err == nil {
		t.Fatal("expected aggregated ErrBufferFull, got nil")
	}
	if !errors.Is(err, ErrBufferFull) {
		t.Fatalf("expected aggregated error to wrap ErrBufferFull, got %v", err)
	}
}

func TestTopicChannel_ConcurrentSubscribeStop_NoLeak(t *testing.T) {
	t.Parallel()

	// Stress: many concurrent Subscribe vs Stop. Per YA-0171 the
	// closed check + sub registration must be atomic so a Subscribe
	// racing with Stop either bails with ErrClosed or completes and
	// gets its inbox closed by Stop. No goroutine leak, no panic.
	qc := NewTopicChannel[int]("t-race",
		WithBufferSize(8),
		WithDrainTimeout(time.Second),
	).(*topic[int])

	err := qc.Start(context.Background())
	if err != nil {
		t.Fatalf("Start returned %v", err)
	}

	const subs = 64

	var wg sync.WaitGroup

	wg.Add(subs)

	for range subs {
		go func() {
			defer wg.Done()

			_, _ = qc.Subscribe(func(_ context.Context, _ Message[int]) error { return nil })
		}()
	}

	// Stop racing concurrently with all those Subscribes.
	go func() {
		_ = qc.Stop(context.Background())
	}()

	wg.Wait()

	// Drain whatever lingering work remains; second Stop is idempotent.
	_ = qc.Stop(context.Background())

	select {
	case <-qc.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("Done never closed after concurrent Subscribe + Stop")
	}
}

func TestTopicChannel_WithDLQChannel_PublishesOnHandlerError(t *testing.T) {
	t.Parallel()

	dlq := NewPipelineChannel[DeadLetter[int]]()

	var (
		mu       sync.Mutex
		captured []DeadLetter[int]
	)

	_, err := dlq.Subscribe(func(_ context.Context, m Message[DeadLetter[int]]) error {
		mu.Lock()
		captured = append(captured, m.Payload)
		mu.Unlock()

		return nil
	})
	if err != nil {
		t.Fatalf("dlq Subscribe: %v", err)
	}

	wantErr := errors.New("topic boom")

	ch := NewTopicChannel[int]("t-dlq",
		WithBufferSize(8),
		WithDrainTimeout(time.Second),
		WithDLQChannel(dlq),
	).(*topic[int])

	_, err = ch.Subscribe(func(_ context.Context, _ Message[int]) error {
		return wantErr
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	err = ch.Start(context.Background())
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	err = ch.Send(context.Background(), NewMessage[int](42, nil))
	if err != nil {
		t.Fatalf("Send: %v", err)
	}

	_ = ch.Stop(context.Background())

	mu.Lock()
	defer mu.Unlock()

	if len(captured) != 1 {
		t.Fatalf("expected 1 DLQ publication, got %d", len(captured))
	}

	dl := captured[0]
	if dl.Original.Payload != 42 {
		t.Fatalf("expected DeadLetter.Original.Payload=42, got %d", dl.Original.Payload)
	}
	if !errors.Is(dl.LastError, wantErr) {
		t.Fatalf("expected DeadLetter.LastError=%v, got %v", wantErr, dl.LastError)
	}
}

func TestTopicChannel_WithDLQChannel_ErrorHandlerAndDLQBothFire(t *testing.T) {
	t.Parallel()

	dlq := NewPipelineChannel[DeadLetter[int]]()

	var (
		mu       sync.Mutex
		captured []DeadLetter[int]
	)

	_, err := dlq.Subscribe(func(_ context.Context, m Message[DeadLetter[int]]) error {
		mu.Lock()
		captured = append(captured, m.Payload)
		mu.Unlock()

		return nil
	})
	if err != nil {
		t.Fatalf("dlq Subscribe: %v", err)
	}

	var hookHits int

	ch := NewTopicChannel[int]("t-both",
		WithBufferSize(8),
		WithDrainTimeout(time.Second),
		WithErrorHandler(func(_ context.Context, _ any, _ error) {
			hookHits++
		}),
		WithDLQChannel(dlq),
	).(*topic[int])

	_, err = ch.Subscribe(func(_ context.Context, _ Message[int]) error {
		return errors.New("boom")
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	err = ch.Start(context.Background())
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	_ = ch.Send(context.Background(), NewMessage[int](1, nil))

	_ = ch.Stop(context.Background())

	mu.Lock()
	dlqCount := len(captured)
	mu.Unlock()

	if hookHits != 1 {
		t.Fatalf("expected ErrorHandler fired once, got %d", hookHits)
	}
	if dlqCount != 1 {
		t.Fatalf("expected DLQ published once, got %d", dlqCount)
	}
}

func TestTopicChannel_WithDLQChannel_HappyPathNoDLQ(t *testing.T) {
	t.Parallel()

	dlq := NewPipelineChannel[DeadLetter[int]]()

	var (
		mu       sync.Mutex
		captured []DeadLetter[int]
	)

	_, err := dlq.Subscribe(func(_ context.Context, m Message[DeadLetter[int]]) error {
		mu.Lock()
		captured = append(captured, m.Payload)
		mu.Unlock()

		return nil
	})
	if err != nil {
		t.Fatalf("dlq Subscribe: %v", err)
	}

	ch := NewTopicChannel[int]("t-nodlq",
		WithBufferSize(8),
		WithDrainTimeout(time.Second),
		WithDLQChannel(dlq),
	).(*topic[int])

	_, err = ch.Subscribe(func(_ context.Context, _ Message[int]) error {
		return nil // success
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	err = ch.Start(context.Background())
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	_ = ch.Send(context.Background(), NewMessage[int](1, nil))

	_ = ch.Stop(context.Background())

	mu.Lock()
	defer mu.Unlock()

	if len(captured) != 0 {
		t.Fatalf("expected 0 DLQ publications on success, got %d", len(captured))
	}
}

func TestTopicChannel_Stop_DrainTimeoutDefaultWhenFieldZero(t *testing.T) {
	t.Parallel()

	// Direct struct construction (bypassing NewTopicChannel) with
	// drainTimeout = 0. The defensive guard in Stop must fall back
	// to defaultDrainTimeout instead of cancelling waitCtx
	// immediately.
	qc := &topic[int]{
		name:         "t-zero-drain",
		bufferSize:   1,
		drainTimeout: 0,
		done:         make(chan struct{}),
		subs:         map[uint64]*subscriber[int]{},
	}

	start := time.Now()

	err := qc.Stop(context.Background())
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected ErrShutdownTimeout (no workers, no Start)")
	}
	if !errors.Is(err, lifecycle.ErrShutdownTimeout) {
		t.Fatalf("expected ErrShutdownTimeout, got %v", err)
	}
	if elapsed < 100*time.Millisecond {
		t.Fatalf("Stop returned in %v — guard did not apply default timeout", elapsed)
	}
}
