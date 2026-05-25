package messaging

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/guidomantilla/yarumo/common/lifecycle"
	lctests "github.com/guidomantilla/yarumo/common/lifecycle/tests"
)

func TestNewTopicChannel(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil channel", func(t *testing.T) {
		t.Parallel()

		qc := NewTopicChannel[int]("q-1")
		if qc == nil {
			t.Fatal("expected non-nil queue channel")
		}
	})

	t.Run("carries the given name", func(t *testing.T) {
		t.Parallel()

		qc := NewTopicChannel[int]("q-named")
		if qc.Name() != "q-named" {
			t.Fatalf("expected name %q, got %q", "q-named", qc.Name())
		}
	})

	t.Run("done channel open at construction", func(t *testing.T) {
		t.Parallel()

		qc := NewTopicChannel[int]("q-done")
		select {
		case <-qc.Done():
			t.Fatal("expected Done open before Stop")
		default:
		}
	})

	t.Run("applies buffer size option", func(t *testing.T) {
		t.Parallel()

		qc := NewTopicChannel[int]("q-buf", WithBufferSize(4))
		if qc.bufferSize != 4 {
			t.Fatalf("expected buffer 4, got %d", qc.bufferSize)
		}
	})
}

func TestTopicChannel_Start(t *testing.T) {
	t.Parallel()

	t.Run("returns nil and accepts sends", func(t *testing.T) {
		t.Parallel()

		qc := NewTopicChannel[int]("q-start", WithBufferSize(4))
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

		qc := NewTopicChannel[int]("q-ctxnil")
		err := qc.Send(nil, NewMessage[int](1, nil)) //nolint:staticcheck
		if !errors.Is(err, ErrContextNil) {
			t.Fatalf("expected ErrContextNil, got %v", err)
		}
	})

	t.Run("returns ErrClosed after Stop", func(t *testing.T) {
		t.Parallel()

		qc := NewTopicChannel[int]("q-closed")
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

		// buffer size 1, no Start so the worker never drains
		qc := NewTopicChannel[int]("q-fullbuf", WithBufferSize(1))

		err := qc.Send(context.Background(), NewMessage[int](1, nil))
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

		qc := NewTopicChannel[int]("q-sub-nil")
		_, err := qc.Subscribe(nil)
		if !errors.Is(err, ErrHandlerNil) {
			t.Fatalf("expected ErrHandlerNil, got %v", err)
		}
	})

	t.Run("returns ErrClosed after Stop", func(t *testing.T) {
		t.Parallel()

		qc := NewTopicChannel[int]("q-sub-closed")
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

		qc := NewTopicChannel[int]("q-sub-cancel")
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

		qc := NewTopicChannel[int]("q-drain", WithBufferSize(16), WithDrainTimeout(time.Second))

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

		qc := NewTopicChannel[int]("q-drain-timeout", WithBufferSize(8), WithDrainTimeout(10*time.Millisecond))

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

		qc := NewTopicChannel[int]("q-stop-idemp")
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

	qc := NewTopicChannel[int]("q-lct")
	err := qc.Start(context.Background())
	if err != nil {
		t.Fatalf("Start returned %v", err)
	}

	lctests.AssertIdempotentStop(t, qc)
}

func TestBuildTopicChannel(t *testing.T) {
	t.Parallel()

	t.Run("returns CloseFn and starts worker", func(t *testing.T) {
		t.Parallel()

		qc := NewTopicChannel[int]("q-build", WithBufferSize(4), WithDrainTimeout(time.Second))

		var delivered int32
		_, err := qc.Subscribe(func(_ context.Context, _ Message[int]) error {
			atomic.AddInt32(&delivered, 1)
			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}

		errChan := make(chan error, 1)
		closeFn, err := BuildTopicChannel(context.Background(), qc, errChan)
		if err != nil {
			t.Fatalf("BuildTopicChannel returned %v", err)
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

func TestTopicChannel_DispatchOrder(t *testing.T) {
	t.Parallel()

	qc := NewTopicChannel[int]("q-order", WithBufferSize(4), WithDrainTimeout(time.Second))

	var (
		mu     sync.Mutex
		order  []int
		signal = make(chan struct{}, 1)
	)

	subscribe := func(id int) {
		_, err := qc.Subscribe(func(_ context.Context, _ Message[int]) error {
			mu.Lock()
			order = append(order, id)
			isLast := len(order) == 4
			mu.Unlock()

			if isLast {
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

	closeFn, err := BuildTopicChannel(context.Background(), qc, errChan)
	if err != nil {
		t.Fatalf("BuildTopicChannel returned %v", err)
	}

	defer closeFn(context.Background(), time.Second)

	err = qc.Send(context.Background(), NewMessage[int](42, nil))
	if err != nil {
		t.Fatalf("Send returned %v", err)
	}

	<-signal

	want := []int{0, 1, 2, 3}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("handler %d fired in wrong order: %v", i, order)
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

	qc := NewTopicChannel[int]("q-panic", WithBufferSize(4), WithDrainTimeout(time.Second), WithErrorHandler(errHook))

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

	closeFn, err := BuildTopicChannel(context.Background(), qc, errChan)
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

	qc := NewTopicChannel[int]("q-errhook", WithBufferSize(4), WithDrainTimeout(time.Second), WithErrorHandler(errHook))

	_, err := qc.Subscribe(func(_ context.Context, _ Message[int]) error {
		return boom
	})
	if err != nil {
		t.Fatalf("Subscribe returned %v", err)
	}

	errChan := make(chan error, 1)

	closeFn, err := BuildTopicChannel(context.Background(), qc, errChan)
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

	qc := NewTopicChannel[int]("q-concur", WithBufferSize(64), WithDrainTimeout(time.Second))

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
