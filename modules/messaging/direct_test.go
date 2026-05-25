package messaging

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
)

func TestNewDirectChannel(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil channel", func(t *testing.T) {
		t.Parallel()

		ch := NewDirectChannel[int]()
		if ch == nil {
			t.Fatal("expected non-nil channel")
		}
	})
}

func TestDirectChannel_Send(t *testing.T) {
	t.Parallel()

	t.Run("delivers to single subscriber", func(t *testing.T) {
		t.Parallel()

		ch := NewDirectChannel[int]()

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

		ch := NewDirectChannel[string]()

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

		ch := NewDirectChannel[int]()

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

		ch := NewDirectChannel[int]()
		err := ch.Send(nil, NewMessage[int](1, nil)) //nolint:staticcheck
		if !errors.Is(err, ErrContextNil) {
			t.Fatalf("expected ErrContextNil, got %v", err)
		}
	})

	t.Run("no subscribers returns nil", func(t *testing.T) {
		t.Parallel()

		ch := NewDirectChannel[int]()
		err := ch.Send(context.Background(), NewMessage[int](1, nil))
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})
}

func TestDirectChannel_Subscribe(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrHandlerNil on nil handler", func(t *testing.T) {
		t.Parallel()

		ch := NewDirectChannel[int]()
		_, err := ch.Subscribe(nil)
		if !errors.Is(err, ErrHandlerNil) {
			t.Fatalf("expected ErrHandlerNil, got %v", err)
		}
	})

	t.Run("cancel detaches handler", func(t *testing.T) {
		t.Parallel()

		ch := NewDirectChannel[int]()

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

		ch := NewDirectChannel[int]()
		cancel, err := ch.Subscribe(func(_ context.Context, _ Message[int]) error { return nil })
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}

		cancel()
		cancel()
		cancel()
	})
}

func TestDirectChannel_Concurrent(t *testing.T) {
	t.Parallel()

	ch := NewDirectChannel[int]()

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
