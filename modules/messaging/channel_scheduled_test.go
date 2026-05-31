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

func TestNewScheduledChannel(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil channel", func(t *testing.T) {
		t.Parallel()

		sc := NewScheduledChannel[int]("sc-1").(*scheduled[int])
		if sc == nil {
			t.Fatal("NewScheduledChannel returned nil")
		}
	})

	t.Run("carries the given name", func(t *testing.T) {
		t.Parallel()

		sc := NewScheduledChannel[int]("sc-named").(*scheduled[int])
		if sc.Name() != "sc-named" {
			t.Fatalf("expected name %q, got %q", "sc-named", sc.Name())
		}
	})

	t.Run("done channel open at construction", func(t *testing.T) {
		t.Parallel()

		sc := NewScheduledChannel[int]("sc-done").(*scheduled[int])
		select {
		case <-sc.Done():
			t.Fatal("expected Done open before Stop")
		default:
		}
	})
}

func TestScheduledChannel_Send(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrContextNil on nil ctx", func(t *testing.T) {
		t.Parallel()

		sc := NewScheduledChannel[int]("sc-ctxnil")

		//nolint:staticcheck // intentional nil ctx to validate guard
		err := sc.Send(nil, NewMessage[int](1, nil))
		if !errors.Is(err, ErrContextNil) {
			t.Fatalf("expected ErrContextNil, got %v", err)
		}
	})

	t.Run("returns ErrClosed after Stop", func(t *testing.T) {
		t.Parallel()

		sc := NewScheduledChannel[int]("sc-closed").(*scheduled[int])

		err := sc.Start(context.Background())
		if err != nil {
			t.Fatalf("Start returned %v", err)
		}

		err = sc.Stop(context.Background())
		if err != nil {
			t.Fatalf("Stop returned %v", err)
		}

		err = sc.Send(context.Background(), NewMessage[int](1, nil))
		if !errors.Is(err, ErrClosed) {
			t.Fatalf("expected ErrClosed, got %v", err)
		}
	})

	t.Run("delivers immediately when sent via Send", func(t *testing.T) {
		t.Parallel()

		sc := NewScheduledChannel[int]("sc-immediate").(*scheduled[int])

		delivered := make(chan int, 1)
		_, err := sc.Subscribe(func(_ context.Context, msg Message[int]) error {
			delivered <- msg.Payload

			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}

		err = sc.Start(context.Background())
		if err != nil {
			t.Fatalf("Start returned %v", err)
		}

		err = sc.Send(context.Background(), NewMessage[int](42, nil))
		if err != nil {
			t.Fatalf("Send returned %v", err)
		}

		select {
		case got := <-delivered:
			if got != 42 {
				t.Fatalf("expected 42, got %d", got)
			}
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for delivery")
		}

		_ = sc.Stop(context.Background())
	})
}

func TestScheduledChannel_SendAt(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrContextNil on nil ctx", func(t *testing.T) {
		t.Parallel()

		sc := NewScheduledChannel[int]("sc-at-ctxnil")

		//nolint:staticcheck // intentional nil ctx to validate guard
		err := sc.SendAt(nil, time.Now(), NewMessage[int](1, nil))
		if !errors.Is(err, ErrContextNil) {
			t.Fatalf("expected ErrContextNil, got %v", err)
		}
	})

	t.Run("past deliverAt delivers immediately", func(t *testing.T) {
		t.Parallel()

		sc := NewScheduledChannel[int]("sc-at-past").(*scheduled[int])

		delivered := make(chan int, 1)
		_, err := sc.Subscribe(func(_ context.Context, msg Message[int]) error {
			delivered <- msg.Payload

			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}

		err = sc.Start(context.Background())
		if err != nil {
			t.Fatalf("Start returned %v", err)
		}

		past := time.Now().Add(-time.Hour)
		err = sc.SendAt(context.Background(), past, NewMessage[int](99, nil))
		if err != nil {
			t.Fatalf("SendAt returned %v", err)
		}

		select {
		case got := <-delivered:
			if got != 99 {
				t.Fatalf("expected 99, got %d", got)
			}
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for past-deliverAt delivery")
		}

		_ = sc.Stop(context.Background())
	})
}

func TestScheduledChannel_SendAfter(t *testing.T) {
	t.Parallel()

	t.Run("delays delivery by at least the requested duration", func(t *testing.T) {
		t.Parallel()

		sc := NewScheduledChannel[int]("sc-after").(*scheduled[int])

		delivered := make(chan time.Time, 1)
		_, err := sc.Subscribe(func(_ context.Context, _ Message[int]) error {
			delivered <- time.Now()

			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}

		err = sc.Start(context.Background())
		if err != nil {
			t.Fatalf("Start returned %v", err)
		}

		const delay = 60 * time.Millisecond
		start := time.Now()
		err = sc.SendAfter(context.Background(), delay, NewMessage[int](1, nil))
		if err != nil {
			t.Fatalf("SendAfter returned %v", err)
		}

		select {
		case at := <-delivered:
			elapsed := at.Sub(start)
			if elapsed < delay {
				t.Fatalf("delivered after %v, want >= %v", elapsed, delay)
			}
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for delayed delivery")
		}

		_ = sc.Stop(context.Background())
	})

	t.Run("non-positive delay delivers immediately", func(t *testing.T) {
		t.Parallel()

		sc := NewScheduledChannel[int]("sc-after-zero").(*scheduled[int])

		delivered := make(chan int, 1)
		_, err := sc.Subscribe(func(_ context.Context, msg Message[int]) error {
			delivered <- msg.Payload

			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}

		err = sc.Start(context.Background())
		if err != nil {
			t.Fatalf("Start returned %v", err)
		}

		err = sc.SendAfter(context.Background(), 0, NewMessage[int](7, nil))
		if err != nil {
			t.Fatalf("SendAfter returned %v", err)
		}

		select {
		case got := <-delivered:
			if got != 7 {
				t.Fatalf("expected 7, got %d", got)
			}
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for zero-delay delivery")
		}

		_ = sc.Stop(context.Background())
	})

	t.Run("delivers in scheduled order regardless of send order", func(t *testing.T) {
		t.Parallel()

		sc := NewScheduledChannel[int]("sc-order").(*scheduled[int])

		const items = 4
		delivered := make(chan int, items)
		_, err := sc.Subscribe(func(_ context.Context, msg Message[int]) error {
			delivered <- msg.Payload

			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}

		err = sc.Start(context.Background())
		if err != nil {
			t.Fatalf("Start returned %v", err)
		}

		// Schedule in reverse delivery order: payload N gets delay (items-N)*step.
		const step = 30 * time.Millisecond
		ctx := context.Background()
		for i := 1; i <= items; i++ {
			delay := time.Duration(items-i+1) * step
			err = sc.SendAfter(ctx, delay, NewMessage[int](i, nil))
			if err != nil {
				t.Fatalf("SendAfter returned %v", err)
			}
		}

		got := make([]int, 0, items)
		deadline := time.After(time.Duration(items+2) * step)
		for len(got) < items {
			select {
			case v := <-delivered:
				got = append(got, v)
			case <-deadline:
				t.Fatalf("timed out collecting deliveries, got %v", got)
			}
		}

		// Expected delivery order is by deliverAt ascending:
		// item with shortest delay first. We scheduled delay(i) =
		// (items-i+1)*step, so smaller delay ↔ larger i. Order is
		// items, items-1, ..., 1.
		for idx, val := range got {
			expected := items - idx
			if val != expected {
				t.Fatalf("got order %v, expected descending from %d", got, items)
			}
		}

		_ = sc.Stop(context.Background())
	})
}

func TestScheduledChannel_Subscribe(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrHandlerNil on nil handler", func(t *testing.T) {
		t.Parallel()

		sc := NewScheduledChannel[int]("sc-sub-nil")
		_, err := sc.Subscribe(nil)
		if !errors.Is(err, ErrHandlerNil) {
			t.Fatalf("expected ErrHandlerNil, got %v", err)
		}
	})

	t.Run("returns ErrClosed after Stop", func(t *testing.T) {
		t.Parallel()

		sc := NewScheduledChannel[int]("sc-sub-closed").(*scheduled[int])

		err := sc.Start(context.Background())
		if err != nil {
			t.Fatalf("Start returned %v", err)
		}

		err = sc.Stop(context.Background())
		if err != nil {
			t.Fatalf("Stop returned %v", err)
		}

		_, err = sc.Subscribe(func(_ context.Context, _ Message[int]) error { return nil })
		if !errors.Is(err, ErrClosed) {
			t.Fatalf("expected ErrClosed, got %v", err)
		}
	})

	t.Run("cancel detaches handler", func(t *testing.T) {
		t.Parallel()

		sc := NewScheduledChannel[int]("sc-sub-cancel").(*scheduled[int])

		var fired atomic.Int32
		cancel, err := sc.Subscribe(func(_ context.Context, _ Message[int]) error {
			fired.Add(1)

			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}

		err = sc.Start(context.Background())
		if err != nil {
			t.Fatalf("Start returned %v", err)
		}

		err = sc.Send(context.Background(), NewMessage[int](1, nil))
		if err != nil {
			t.Fatalf("Send returned %v", err)
		}

		// Allow first delivery.
		deadline := time.Now().Add(time.Second)
		for time.Now().Before(deadline) {
			if fired.Load() == 1 {
				break
			}
			time.Sleep(time.Millisecond)
		}

		cancel()
		cancel() // idempotent

		err = sc.Send(context.Background(), NewMessage[int](2, nil))
		if err != nil {
			t.Fatalf("Send returned %v", err)
		}

		time.Sleep(50 * time.Millisecond)

		got := fired.Load()
		if got != 1 {
			t.Fatalf("expected fired=1 after cancel, got %d", got)
		}

		_ = sc.Stop(context.Background())
	})

	t.Run("fan-out to multiple subscribers", func(t *testing.T) {
		t.Parallel()

		sc := NewScheduledChannel[int]("sc-fanout").(*scheduled[int])

		var a, b atomic.Int32
		_, err := sc.Subscribe(func(_ context.Context, _ Message[int]) error {
			a.Add(1)

			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe A returned %v", err)
		}

		_, err = sc.Subscribe(func(_ context.Context, _ Message[int]) error {
			b.Add(1)

			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe B returned %v", err)
		}

		err = sc.Start(context.Background())
		if err != nil {
			t.Fatalf("Start returned %v", err)
		}

		err = sc.Send(context.Background(), NewMessage[int](1, nil))
		if err != nil {
			t.Fatalf("Send returned %v", err)
		}

		deadline := time.Now().Add(time.Second)
		for time.Now().Before(deadline) {
			if a.Load() == 1 && b.Load() == 1 {
				break
			}
			time.Sleep(time.Millisecond)
		}

		if a.Load() != 1 || b.Load() != 1 {
			t.Fatalf("expected both subs fired once, got a=%d b=%d", a.Load(), b.Load())
		}

		_ = sc.Stop(context.Background())
	})
}

func TestScheduledChannel_Stop(t *testing.T) {
	t.Parallel()

	t.Run("is idempotent across multiple calls", func(t *testing.T) {
		t.Parallel()

		sc := NewScheduledChannel[int]("sc-stop-idemp").(*scheduled[int])

		err := sc.Start(context.Background())
		if err != nil {
			t.Fatalf("Start returned %v", err)
		}

		err = sc.Stop(context.Background())
		if err != nil {
			t.Fatalf("first Stop returned %v", err)
		}

		err = sc.Stop(context.Background())
		if err != nil {
			t.Fatalf("second Stop returned %v", err)
		}
	})

	t.Run("returns ErrShutdownTimeout when drain exceeds bound", func(t *testing.T) {
		t.Parallel()

		// Construct with a tiny drain timeout, then poison the worker
		// by replacing workerCancel with a no-op so Stop cannot kill
		// the goroutine. The scheduler keeps running, drain times out.
		sc := NewScheduledChannel[int]("sc-drain-timeout",
			WithDrainTimeout(10*time.Millisecond),
		).(*scheduled[int])

		err := sc.Start(context.Background())
		if err != nil {
			t.Fatalf("Start returned %v", err)
		}

		// Replace workerCancel with a no-op so the scheduler never
		// observes ctx cancellation. This is a deliberate poisoning
		// of an internal field — the only stable way to drive the
		// drain-timeout path without holding the worker live via a
		// real handler. Restore after the assertion so the test
		// cleanup can actually stop the goroutine.
		sc.subsMu.Lock()
		realCancel := sc.workerCancel
		sc.workerCancel = func() {}
		sc.subsMu.Unlock()

		err = sc.Stop(context.Background())
		if err == nil {
			t.Fatal("expected drain-timeout error")
		}

		if !errors.Is(err, lifecycle.ErrShutdownTimeout) {
			t.Fatalf("expected ErrShutdownTimeout, got %v", err)
		}

		// Cleanup: actually shut down the scheduler so the goroutine exits.
		realCancel()
		<-sc.Done()
	})
}

func TestScheduledChannel_StopIsIdempotent(t *testing.T) {
	t.Parallel()

	sc := NewScheduledChannel[int]("sc-lct").(*scheduled[int])
	err := sc.Start(context.Background())
	if err != nil {
		t.Fatalf("Start returned %v", err)
	}

	lctests.AssertIdempotentStop(t, sc)
}

func TestScheduledChannel_LifecycleIntegration(t *testing.T) {
	t.Parallel()

	t.Run("lifecycle.Build wires the worker and CloseFn drains", func(t *testing.T) {
		t.Parallel()

		sc := NewScheduledChannel[int]("sc-lifecycle",
			WithDrainTimeout(time.Second),
		).(*scheduled[int])

		var delivered atomic.Int32
		_, err := sc.Subscribe(func(_ context.Context, _ Message[int]) error {
			delivered.Add(1)

			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}

		errChan := make(chan error, 1)
		closeFn, err := lifecycle.Build(context.Background(), sc, errChan)
		if err != nil {
			t.Fatalf("lifecycle.Build returned %v", err)
		}

		err = sc.Send(context.Background(), NewMessage[int](1, nil))
		if err != nil {
			t.Fatalf("Send returned %v", err)
		}

		// Wait until the worker has had a chance to dispatch.
		deadline := time.Now().Add(time.Second)
		for time.Now().Before(deadline) {
			if delivered.Load() == 1 {
				break
			}
			time.Sleep(time.Millisecond)
		}

		closeFn(context.Background(), time.Second)

		got := delivered.Load()
		if got != 1 {
			t.Fatalf("expected 1 delivery, got %d", got)
		}
	})
}

func TestScheduledChannel_ConcurrentSendAt(t *testing.T) {
	t.Parallel()

	sc := NewScheduledChannel[int]("sc-conc").(*scheduled[int])

	const producers = 8
	const perProducer = 25
	const totalMsgs = producers * perProducer

	var delivered atomic.Int32
	_, err := sc.Subscribe(func(_ context.Context, _ Message[int]) error {
		delivered.Add(1)

		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe returned %v", err)
	}

	err = sc.Start(context.Background())
	if err != nil {
		t.Fatalf("Start returned %v", err)
	}

	var wg sync.WaitGroup
	ctx := context.Background()
	baseDelay := 20 * time.Millisecond

	for p := range producers {
		wg.Go(func() {
			for i := range perProducer {
				delay := baseDelay + time.Duration(i)*time.Microsecond
				err := sc.SendAfter(ctx, delay, NewMessage[int](p*perProducer+i, nil))
				if err != nil {
					t.Errorf("SendAfter returned %v", err)

					return
				}
			}
		})
	}

	wg.Wait()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if delivered.Load() == totalMsgs {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	got := delivered.Load()
	if got != totalMsgs {
		t.Fatalf("expected %d deliveries, got %d", totalMsgs, got)
	}

	_ = sc.Stop(context.Background())
}
