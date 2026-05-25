package messaging

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
)

func TestNewPipelineChannel(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil channel", func(t *testing.T) {
		t.Parallel()

		ch := NewPipelineChannel[int]()
		if ch == nil {
			t.Fatal("expected non-nil channel")
		}
	})
}

func TestPipelineChannel_Send(t *testing.T) {
	t.Parallel()

	t.Run("delivers to single subscriber", func(t *testing.T) {
		t.Parallel()

		ch := NewPipelineChannel[int]()

		var got int
		_, err := ch.Subscribe(func(_ context.Context, msg Message[int]) error {
			got = msg.Payload
			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}

		err = ch.Send(context.Background(), NewMessage[int](42, nil))
		if err != nil {
			t.Fatalf("Send returned %v", err)
		}

		if got != 42 {
			t.Fatalf("expected payload 42, got %d", got)
		}
	})

	t.Run("delivers to multiple subscribers", func(t *testing.T) {
		t.Parallel()

		ch := NewPipelineChannel[string]()

		var count int32
		for range 3 {
			_, err := ch.Subscribe(func(_ context.Context, _ Message[string]) error {
				atomic.AddInt32(&count, 1)
				return nil
			})
			if err != nil {
				t.Fatalf("Subscribe returned %v", err)
			}
		}

		err := ch.Send(context.Background(), NewMessage[string]("hello", nil))
		if err != nil {
			t.Fatalf("Send returned %v", err)
		}

		if atomic.LoadInt32(&count) != 3 {
			t.Fatalf("expected 3 deliveries, got %d", count)
		}
	})

	t.Run("returns first handler error", func(t *testing.T) {
		t.Parallel()

		ch := NewPipelineChannel[int]()

		boom := errors.New("boom")
		_, err := ch.Subscribe(func(_ context.Context, _ Message[int]) error {
			return boom
		})
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}

		err = ch.Send(context.Background(), NewMessage[int](1, nil))
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, boom) {
			t.Fatalf("expected error to wrap boom, got %v", err)
		}

		if !errors.Is(err, ErrSendFailed) {
			t.Fatalf("expected error to wrap ErrSendFailed, got %v", err)
		}
	})

	t.Run("returns ErrContextNil on nil ctx", func(t *testing.T) {
		t.Parallel()

		ch := NewPipelineChannel[int]()
		err := ch.Send(nil, NewMessage[int](1, nil)) //nolint:staticcheck
		if !errors.Is(err, ErrContextNil) {
			t.Fatalf("expected ErrContextNil, got %v", err)
		}
	})

	t.Run("no subscribers returns nil", func(t *testing.T) {
		t.Parallel()

		ch := NewPipelineChannel[int]()
		err := ch.Send(context.Background(), NewMessage[int](1, nil))
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})
}

func TestPipelineChannel_Subscribe(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrHandlerNil on nil handler", func(t *testing.T) {
		t.Parallel()

		ch := NewPipelineChannel[int]()
		_, err := ch.Subscribe(nil)
		if !errors.Is(err, ErrHandlerNil) {
			t.Fatalf("expected ErrHandlerNil, got %v", err)
		}
	})

	t.Run("cancel detaches handler", func(t *testing.T) {
		t.Parallel()

		ch := NewPipelineChannel[int]()

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

		err = ch.Send(context.Background(), NewMessage[int](2, nil))
		if err != nil {
			t.Fatalf("Send returned %v", err)
		}

		got := atomic.LoadInt32(&fired)
		if got != 1 {
			t.Fatalf("expected handler fired once, got %d", got)
		}
	})

	t.Run("cancel is idempotent", func(t *testing.T) {
		t.Parallel()

		ch := NewPipelineChannel[int]()
		cancel, err := ch.Subscribe(func(_ context.Context, _ Message[int]) error { return nil })
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}

		cancel()
		cancel()
		cancel()
	})
}

func TestPipelineChannel_DispatchOrder(t *testing.T) {
	t.Parallel()

	ch := NewPipelineChannel[int]()

	var order []int

	var mu sync.Mutex

	subscribe := func(id int) {
		_, err := ch.Subscribe(func(_ context.Context, _ Message[int]) error {
			mu.Lock()
			order = append(order, id)
			mu.Unlock()

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

	err := ch.Send(context.Background(), NewMessage[int](42, nil))
	if err != nil {
		t.Fatalf("Send returned %v", err)
	}

	want := []int{0, 1, 2, 3}
	if len(order) != len(want) {
		t.Fatalf("expected %d handlers fired, got %d", len(want), len(order))
	}

	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("step %d fired in wrong order: %v", i, order)
		}
	}
}

func TestPipelineChannel_ChainErrorTrace(t *testing.T) {
	t.Parallel()

	ch := NewPipelineChannel[int]()

	boom := errors.New("boom")

	_, err := ch.Subscribe(func(_ context.Context, _ Message[int]) error { return nil })
	if err != nil {
		t.Fatalf("Subscribe returned %v", err)
	}

	_, err = ch.Subscribe(func(_ context.Context, _ Message[int]) error { return nil })
	if err != nil {
		t.Fatalf("Subscribe returned %v", err)
	}

	_, err = ch.Subscribe(func(_ context.Context, _ Message[int]) error { return boom })
	if err != nil {
		t.Fatalf("Subscribe returned %v", err)
	}

	_, err = ch.Subscribe(func(_ context.Context, _ Message[int]) error {
		t.Fatal("step 3 should have been skipped")
		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe returned %v", err)
	}

	err = ch.Send(context.Background(), NewMessage[int](1, nil))
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var chainErr *ChainError

	if !errors.As(err, &chainErr) {
		t.Fatalf("expected *ChainError, got %T: %v", err, err)
	}

	if chainErr.Failed != 2 {
		t.Fatalf("expected Failed=2, got %d", chainErr.Failed)
	}

	if len(chainErr.Steps) != 4 {
		t.Fatalf("expected 4 steps in trace, got %d", len(chainErr.Steps))
	}

	want := []StepStatus{StepStatusOK, StepStatusOK, StepStatusError, StepStatusSkipped}
	for i, w := range want {
		if chainErr.Steps[i].Status != w {
			t.Fatalf("step %d: expected status %q, got %q", i, w, chainErr.Steps[i].Status)
		}
	}

	if !errors.Is(err, boom) {
		t.Fatalf("Unwrap chain should surface boom, got %v", err)
	}

	if !errors.Is(err, ErrChainFailed) {
		t.Fatalf("expected ErrChainFailed wrapped, got %v", err)
	}
}

func TestPipelineChannel_PanicRecovery(t *testing.T) {
	t.Parallel()

	ch := NewPipelineChannel[int]()

	_, err := ch.Subscribe(func(_ context.Context, _ Message[int]) error { return nil })
	if err != nil {
		t.Fatalf("Subscribe returned %v", err)
	}

	_, err = ch.Subscribe(func(_ context.Context, _ Message[int]) error {
		panic("kaboom")
	})
	if err != nil {
		t.Fatalf("Subscribe returned %v", err)
	}

	_, err = ch.Subscribe(func(_ context.Context, _ Message[int]) error {
		t.Fatal("step 2 should have been skipped after panic")
		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe returned %v", err)
	}

	err = ch.Send(context.Background(), NewMessage[int](1, nil))
	if err == nil {
		t.Fatal("expected error from panicking handler, got nil")
	}

	var chainErr *ChainError

	if !errors.As(err, &chainErr) {
		t.Fatalf("expected *ChainError, got %T", err)
	}

	if chainErr.Failed != 1 {
		t.Fatalf("expected Failed=1, got %d", chainErr.Failed)
	}

	if chainErr.Steps[1].Status != StepStatusPanic {
		t.Fatalf("expected step 1 panic, got %q", chainErr.Steps[1].Status)
	}

	if !errors.Is(err, ErrHandlerPanic) {
		t.Fatalf("expected ErrHandlerPanic wrapped, got %v", err)
	}
}

func TestPipelineChannel_Concurrent(t *testing.T) {
	t.Parallel()

	ch := NewPipelineChannel[int]()

	var fired int64
	const sends = 100
	const subs = 5

	for range subs {
		_, err := ch.Subscribe(func(_ context.Context, _ Message[int]) error {
			atomic.AddInt64(&fired, 1)
			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}
	}

	var wg sync.WaitGroup
	wg.Add(sends)
	for i := range sends {
		go func(n int) {
			defer wg.Done()
			err := ch.Send(context.Background(), NewMessage[int](n, nil))
			if err != nil {
				t.Errorf("Send returned %v", err)
			}
		}(i)
	}

	wg.Wait()

	got := atomic.LoadInt64(&fired)
	want := int64(sends * subs)
	if got != want {
		t.Fatalf("expected %d deliveries, got %d", want, got)
	}
}
