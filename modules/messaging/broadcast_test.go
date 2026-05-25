package messaging

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewBroadcastChannel(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil channel", func(t *testing.T) {
		t.Parallel()

		ch := NewBroadcastChannel[int]()
		if ch == nil {
			t.Fatal("expected non-nil channel")
		}
	})
}

func TestBroadcastChannel_Send(t *testing.T) {
	t.Parallel()

	t.Run("delivers to every subscriber in parallel", func(t *testing.T) {
		t.Parallel()

		ch := NewBroadcastChannel[int]()

		const subs = 4

		var (
			fired int32
			start sync.WaitGroup
			ready sync.WaitGroup
		)

		start.Add(1)
		ready.Add(subs)

		for range subs {
			_, err := ch.Subscribe(func(_ context.Context, _ Message[int]) error {
				ready.Done()
				start.Wait() // pause until all handlers have entered → proves parallel
				atomic.AddInt32(&fired, 1)
				return nil
			})
			if err != nil {
				t.Fatalf("Subscribe returned %v", err)
			}
		}

		done := make(chan error, 1)
		go func() {
			done <- ch.Send(context.Background(), NewMessage[int](42, nil))
		}()

		// Wait for all handlers to be inside before releasing.
		ready.Wait()
		start.Done()

		err := <-done
		if err != nil {
			t.Fatalf("Send returned %v", err)
		}

		if atomic.LoadInt32(&fired) != subs {
			t.Fatalf("expected %d deliveries, got %d", subs, fired)
		}
	})

	t.Run("waits for all handlers (barrier semantics)", func(t *testing.T) {
		t.Parallel()

		ch := NewBroadcastChannel[int]()

		var slowDone int32

		_, err := ch.Subscribe(func(_ context.Context, _ Message[int]) error {
			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}

		_, err = ch.Subscribe(func(_ context.Context, _ Message[int]) error {
			time.Sleep(20 * time.Millisecond)
			atomic.StoreInt32(&slowDone, 1)
			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}

		err = ch.Send(context.Background(), NewMessage[int](1, nil))
		if err != nil {
			t.Fatalf("Send returned %v", err)
		}

		if atomic.LoadInt32(&slowDone) != 1 {
			t.Fatal("Send returned before slow handler finished — barrier broken")
		}
	})

	t.Run("aggregates errors from all failing handlers (no fail-fast)", func(t *testing.T) {
		t.Parallel()

		ch := NewBroadcastChannel[int]()

		errA := errors.New("err-a")
		errB := errors.New("err-b")

		var fired int32

		_, err := ch.Subscribe(func(_ context.Context, _ Message[int]) error {
			atomic.AddInt32(&fired, 1)
			return errA
		})
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}

		_, err = ch.Subscribe(func(_ context.Context, _ Message[int]) error {
			atomic.AddInt32(&fired, 1)
			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}

		_, err = ch.Subscribe(func(_ context.Context, _ Message[int]) error {
			atomic.AddInt32(&fired, 1)
			return errB
		})
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}

		err = ch.Send(context.Background(), NewMessage[int](1, nil))
		if err == nil {
			t.Fatal("expected aggregated error, got nil")
		}

		if !errors.Is(err, errA) {
			t.Fatalf("expected errA wrapped, got %v", err)
		}

		if !errors.Is(err, errB) {
			t.Fatalf("expected errB wrapped, got %v", err)
		}

		if atomic.LoadInt32(&fired) != 3 {
			t.Fatalf("expected all 3 handlers fired, got %d", fired)
		}
	})

	t.Run("recovers panic per handler and reports as joined error", func(t *testing.T) {
		t.Parallel()

		ch := NewBroadcastChannel[int]()

		var okFired int32

		_, err := ch.Subscribe(func(_ context.Context, _ Message[int]) error {
			panic("boom")
		})
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}

		_, err = ch.Subscribe(func(_ context.Context, _ Message[int]) error {
			atomic.AddInt32(&okFired, 1)
			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}

		err = ch.Send(context.Background(), NewMessage[int](1, nil))
		if !errors.Is(err, ErrHandlerPanic) {
			t.Fatalf("expected ErrHandlerPanic, got %v", err)
		}

		if atomic.LoadInt32(&okFired) != 1 {
			t.Fatal("non-panicking handler should have still fired")
		}
	})

	t.Run("no subscribers returns nil", func(t *testing.T) {
		t.Parallel()

		ch := NewBroadcastChannel[int]()
		err := ch.Send(context.Background(), NewMessage[int](1, nil))
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("returns ErrContextNil on nil ctx", func(t *testing.T) {
		t.Parallel()

		ch := NewBroadcastChannel[int]()
		err := ch.Send(nil, NewMessage[int](1, nil)) //nolint:staticcheck
		if !errors.Is(err, ErrContextNil) {
			t.Fatalf("expected ErrContextNil, got %v", err)
		}
	})
}

func TestBroadcastChannel_Subscribe(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrHandlerNil on nil handler", func(t *testing.T) {
		t.Parallel()

		ch := NewBroadcastChannel[int]()
		_, err := ch.Subscribe(nil)
		if !errors.Is(err, ErrHandlerNil) {
			t.Fatalf("expected ErrHandlerNil, got %v", err)
		}
	})

	t.Run("cancel detaches handler", func(t *testing.T) {
		t.Parallel()

		ch := NewBroadcastChannel[int]()

		var fired int32

		cancel, err := ch.Subscribe(func(_ context.Context, _ Message[int]) error {
			atomic.AddInt32(&fired, 1)
			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}

		err = ch.Send(context.Background(), NewMessage[int](1, nil))
		if err != nil {
			t.Fatalf("Send returned %v", err)
		}

		cancel()
		cancel() // idempotent

		err = ch.Send(context.Background(), NewMessage[int](2, nil))
		if err != nil {
			t.Fatalf("Send returned %v", err)
		}

		if atomic.LoadInt32(&fired) != 1 {
			t.Fatalf("expected fired=1, got %d", fired)
		}
	})
}
