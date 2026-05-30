package messaging

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/guidomantilla/yarumo/core/common/lifecycle"
)

func TestNewQueueChannel(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil channel", func(t *testing.T) {
		t.Parallel()

		qc := NewQueueChannel[int]("q-1").(*queue[int])
		if qc == nil {
			t.Fatal("expected non-nil channel")
		}
	})

	t.Run("applies worker count option", func(t *testing.T) {
		t.Parallel()

		qc := NewQueueChannel[int]("q-pool", WithWorkerCount(4)).(*queue[int])
		if qc.workerCount != 4 {
			t.Fatalf("expected workerCount=4, got %d", qc.workerCount)
		}
	})
}

func TestQueueChannel_RoundRobin(t *testing.T) {
	t.Parallel()

	qc := NewQueueChannel[int]("q-rr", WithBufferSize(16), WithDrainTimeout(time.Second)).(*queue[int])

	const subs = 3

	var (
		mu      sync.Mutex
		fired   = make(map[int]int)
		signal  = make(chan struct{}, 32)
		labels  = []int{0, 1, 2}
		handlers = make([]func(context.Context, Message[int]) error, subs)
	)

	for i, label := range labels {
		handlers[i] = func(_ context.Context, _ Message[int]) error {
			mu.Lock()
			fired[label]++
			mu.Unlock()

			signal <- struct{}{}

			return nil
		}
	}

	for i := range subs {
		_, err := qc.Subscribe(handlers[i])
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}
	}

	errChan := make(chan error, 1)

	closeFn, err := lifecycle.Build(context.Background(), qc, errChan)
	if err != nil {
		t.Fatalf("lifecycle.Build returned %v", err)
	}

	t.Cleanup(func() { closeFn(context.Background(), time.Second) })

	const sends = 9

	for i := range sends {
		err = qc.Send(context.Background(), NewMessage[int](i, nil))
		if err != nil {
			t.Fatalf("Send returned %v", err)
		}
	}

	for range sends {
		<-signal
	}

	closeFn(context.Background(), time.Second)

	mu.Lock()
	defer mu.Unlock()

	// With round-robin and 9 messages over 3 subs, each sub should
	// fire exactly 3 times.
	for label, count := range fired {
		if count != 3 {
			t.Errorf("sub %d fired %d times, expected 3 (round-robin distribution)", label, count)
		}
	}
}

func TestQueueChannel_PointToPoint(t *testing.T) {
	t.Parallel()

	qc := NewQueueChannel[int]("q-p2p", WithBufferSize(4), WithDrainTimeout(time.Second)).(*queue[int])

	var subA, subB int32

	done := make(chan struct{}, 1)

	_, err := qc.Subscribe(func(_ context.Context, _ Message[int]) error {
		atomic.AddInt32(&subA, 1)
		done <- struct{}{}

		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe returned %v", err)
	}

	_, err = qc.Subscribe(func(_ context.Context, _ Message[int]) error {
		atomic.AddInt32(&subB, 1)
		done <- struct{}{}

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

	t.Cleanup(func() { closeFn(context.Background(), time.Second) })

	// Send one msg; exactly one subscriber should fire (not both).
	err = qc.Send(context.Background(), NewMessage[int](1, nil))
	if err != nil {
		t.Fatalf("Send returned %v", err)
	}

	<-done

	time.Sleep(20 * time.Millisecond) // give a window for accidental double-fire

	total := atomic.LoadInt32(&subA) + atomic.LoadInt32(&subB)
	if total != 1 {
		t.Fatalf("expected exactly one subscriber to fire (point-to-point), got total=%d (a=%d b=%d)", total, subA, subB)
	}
}

func TestQueueChannel_NoSubscribersHook(t *testing.T) {
	t.Parallel()

	var hookErr error

	hookFired := make(chan struct{}, 1)
	hook := func(_ context.Context, _ any, err error) {
		hookErr = err

		select {
		case hookFired <- struct{}{}:
		default:
		}
	}

	qc := NewQueueChannel[int]("q-empty",
		WithBufferSize(4),
		WithDrainTimeout(time.Second),
		WithErrorHandler(hook),
	).(*queue[int])

	errChan := make(chan error, 1)

	closeFn, err := lifecycle.Build(context.Background(), qc, errChan)
	if err != nil {
		t.Fatalf("lifecycle.Build returned %v", err)
	}

	t.Cleanup(func() { closeFn(context.Background(), time.Second) })

	err = qc.Send(context.Background(), NewMessage[int](1, nil))
	if err != nil {
		t.Fatalf("Send returned %v", err)
	}

	<-hookFired

	if !errors.Is(hookErr, ErrNoSubscribers) {
		t.Fatalf("expected ErrNoSubscribers, got %v", hookErr)
	}
}

func TestQueueChannel_PanicRecovery(t *testing.T) {
	t.Parallel()

	var hookErr error

	hookFired := make(chan struct{}, 1)
	hook := func(_ context.Context, _ any, err error) {
		hookErr = err

		hookFired <- struct{}{}
	}

	qc := NewQueueChannel[int]("q-panic",
		WithBufferSize(4),
		WithDrainTimeout(time.Second),
		WithErrorHandler(hook),
	).(*queue[int])

	var nextFired int32

	_, err := qc.Subscribe(func(_ context.Context, _ Message[int]) error {
		panic("kaboom")
	})
	if err != nil {
		t.Fatalf("Subscribe returned %v", err)
	}

	_, err = qc.Subscribe(func(_ context.Context, _ Message[int]) error {
		atomic.AddInt32(&nextFired, 1)
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

	t.Cleanup(func() { closeFn(context.Background(), time.Second) })

	err = qc.Send(context.Background(), NewMessage[int](1, nil))
	if err != nil {
		t.Fatalf("Send returned %v", err)
	}

	<-hookFired

	if !errors.Is(hookErr, ErrHandlerPanic) {
		t.Fatalf("expected ErrHandlerPanic, got %v", hookErr)
	}

	// Round-robin should pick the non-panicking sub next.
	err = qc.Send(context.Background(), NewMessage[int](2, nil))
	if err != nil {
		t.Fatalf("Send returned %v", err)
	}

	closeFn(context.Background(), time.Second)

	if atomic.LoadInt32(&nextFired) != 1 {
		t.Fatal("subsequent message should have routed to the non-panicking sub")
	}
}

func TestQueueChannel_Send(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrContextNil on nil ctx", func(t *testing.T) {
		t.Parallel()

		qc := NewQueueChannel[int]("q-ctx").(*queue[int])
		err := qc.Send(nil, NewMessage[int](1, nil)) //nolint:staticcheck
		if !errors.Is(err, ErrContextNil) {
			t.Fatalf("expected ErrContextNil, got %v", err)
		}
	})

	t.Run("returns ErrClosed after Stop", func(t *testing.T) {
		t.Parallel()

		qc := NewQueueChannel[int]("q-stopped").(*queue[int])
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
}

func TestQueueChannel_Subscribe(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrHandlerNil on nil handler", func(t *testing.T) {
		t.Parallel()

		qc := NewQueueChannel[int]("q-sub-nil").(*queue[int])
		_, err := qc.Subscribe(nil)
		if !errors.Is(err, ErrHandlerNil) {
			t.Fatalf("expected ErrHandlerNil, got %v", err)
		}
	})
}

func TestQueueChannel_StopIsIdempotent(t *testing.T) {
	t.Parallel()

	qc := NewQueueChannel[int]("q-stop", WithDrainTimeout(time.Second)).(*queue[int])
	_ = qc.Start(context.Background())

	err := qc.Stop(context.Background())
	if err != nil {
		t.Fatalf("first Stop returned %v", err)
	}

	err = qc.Stop(context.Background())
	if err != nil {
		t.Fatalf("second Stop returned %v", err)
	}
}

func TestQueueChannel_OverflowReject_ReturnsErrBufferFull(t *testing.T) {
	t.Parallel()

	ch := NewQueueChannel[int]("q-reject", WithBufferSize(1)).(*queue[int])

	err := ch.Send(context.Background(), NewMessage[int](1, nil))
	if err != nil {
		t.Fatalf("first Send returned %v", err)
	}

	err = ch.Send(context.Background(), NewMessage[int](2, nil))
	if !errors.Is(err, ErrBufferFull) {
		t.Fatalf("expected ErrBufferFull, got %v", err)
	}
}

func TestQueueChannel_OverflowBlock_BlocksUntilCtxExpires(t *testing.T) {
	t.Parallel()

	ch := NewQueueChannel[int]("q-block",
		WithBufferSize(1),
		WithOverflowPolicy(OverflowBlock),
	).(*queue[int])

	err := ch.Send(context.Background(), NewMessage[int](1, nil))
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

func TestQueueChannel_OverflowDropNewest_DropsNewMessage(t *testing.T) {
	t.Parallel()

	var captured []Message[int]
	hook := func(_ context.Context, msg any, _ error) {
		captured = append(captured, msg.(Message[int]))
	}

	ch := NewQueueChannel[int]("q-newest",
		WithBufferSize(1),
		WithOverflowPolicy(OverflowDropNewest),
		WithErrorHandler(hook),
	).(*queue[int])

	_ = ch.Send(context.Background(), NewMessage[int](1, nil))

	err := ch.Send(context.Background(), NewMessage[int](2, nil))
	if err != nil {
		t.Fatalf("DropNewest Send should return nil, got %v", err)
	}

	if len(captured) != 1 || captured[0].Payload != 2 {
		t.Fatalf("expected new msg (2) dropped, got %+v", captured)
	}
}

func TestQueueChannel_OverflowDropOldest_EvictsAndAcceptsNew(t *testing.T) {
	t.Parallel()

	var captured []Message[int]
	hook := func(_ context.Context, msg any, _ error) {
		captured = append(captured, msg.(Message[int]))
	}

	ch := NewQueueChannel[int]("q-oldest",
		WithBufferSize(2),
		WithOverflowPolicy(OverflowDropOldest),
		WithErrorHandler(hook),
	).(*queue[int])

	_ = ch.Send(context.Background(), NewMessage[int](1, nil))
	_ = ch.Send(context.Background(), NewMessage[int](2, nil))

	err := ch.Send(context.Background(), NewMessage[int](3, nil))
	if err != nil {
		t.Fatalf("DropOldest Send should return nil, got %v", err)
	}

	if len(captured) != 1 || captured[0].Payload != 1 {
		t.Fatalf("expected oldest (1) evicted, got %+v", captured)
	}
}

func TestQueueChannel_ConcurrentSubscribeStop_NoLeak(t *testing.T) {
	t.Parallel()

	// Per YA-0171 the closed check + handler registration must be
	// atomic so a Subscribe racing with Stop either bails with
	// ErrClosed or registers before Stop closes the inbound channel.
	qc := NewQueueChannel[int]("q-race",
		WithBufferSize(8),
		WithDrainTimeout(time.Second),
	).(*queue[int])

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

	go func() {
		_ = qc.Stop(context.Background())
	}()

	wg.Wait()

	_ = qc.Stop(context.Background())

	select {
	case <-qc.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("Done never closed after concurrent Subscribe + Stop")
	}
}

func TestQueueChannel_WithDLQChannel_PublishesOnHandlerError(t *testing.T) {
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

	wantErr := errors.New("queue boom")

	ch := NewQueueChannel[int]("q-dlq",
		WithBufferSize(8),
		WithDrainTimeout(time.Second),
		WithDLQChannel(dlq),
	).(*queue[int])

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

	err = ch.Send(context.Background(), NewMessage[int](99, nil))
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
	if dl.Original.Payload != 99 {
		t.Fatalf("expected DeadLetter.Original.Payload=99, got %d", dl.Original.Payload)
	}
	if !errors.Is(dl.LastError, wantErr) {
		t.Fatalf("expected DeadLetter.LastError=%v, got %v", wantErr, dl.LastError)
	}
}

func TestQueueChannel_Stop_DrainTimeoutDefaultWhenFieldZero(t *testing.T) {
	t.Parallel()

	qc := &queue[int]{
		name:         "q-zero-drain",
		bufferSize:   1,
		workerCount:  1,
		drainTimeout: 0,
		inbound:      make(chan envelope[int], 1),
		done:         make(chan struct{}),
		byID:         map[uint64]Handler[int]{},
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
