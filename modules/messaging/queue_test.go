package messaging

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/guidomantilla/yarumo/common/lifecycle"
)

func TestNewQueueChannel(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil channel", func(t *testing.T) {
		t.Parallel()

		qc := NewQueueChannel[int]("q-1")
		if qc == nil {
			t.Fatal("expected non-nil channel")
		}
	})

	t.Run("applies worker count option", func(t *testing.T) {
		t.Parallel()

		qc := NewQueueChannel[int]("q-pool", WithWorkerCount(4))
		if qc.workerCount != 4 {
			t.Fatalf("expected workerCount=4, got %d", qc.workerCount)
		}
	})
}

func TestQueueChannel_RoundRobin(t *testing.T) {
	t.Parallel()

	qc := NewQueueChannel[int]("q-rr", WithBufferSize(16), WithDrainTimeout(time.Second))

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

	qc := NewQueueChannel[int]("q-p2p", WithBufferSize(4), WithDrainTimeout(time.Second))

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
	)

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
	)

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

		qc := NewQueueChannel[int]("q-ctx")
		err := qc.Send(nil, NewMessage[int](1, nil)) //nolint:staticcheck
		if !errors.Is(err, ErrContextNil) {
			t.Fatalf("expected ErrContextNil, got %v", err)
		}
	})

	t.Run("returns ErrClosed after Stop", func(t *testing.T) {
		t.Parallel()

		qc := NewQueueChannel[int]("q-stopped")
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

		qc := NewQueueChannel[int]("q-sub-nil")
		_, err := qc.Subscribe(nil)
		if !errors.Is(err, ErrHandlerNil) {
			t.Fatalf("expected ErrHandlerNil, got %v", err)
		}
	})
}

func TestQueueChannel_StopIsIdempotent(t *testing.T) {
	t.Parallel()

	qc := NewQueueChannel[int]("q-stop", WithDrainTimeout(time.Second))
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
