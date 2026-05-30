package messaging

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewPollableChannel(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil channel", func(t *testing.T) {
		t.Parallel()

		ch := NewPollableChannel[int]()
		if ch == nil {
			t.Fatal("NewPollableChannel returned nil")
		}
	})

	t.Run("applies buffer size option", func(t *testing.T) {
		t.Parallel()

		ch := NewPollableChannel[int](WithBufferSize(2)).(*pollable[int])
		if cap(ch.buf) != 2 {
			t.Fatalf("expected buffer cap 2, got %d", cap(ch.buf))
		}
	})

	t.Run("default buffer size when option absent", func(t *testing.T) {
		t.Parallel()

		ch := NewPollableChannel[int]().(*pollable[int])
		if cap(ch.buf) != defaultBufferSize {
			t.Fatalf("expected buffer cap %d, got %d", defaultBufferSize, cap(ch.buf))
		}
	})
}

func TestPollableChannel_Send(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrContextNil on nil ctx", func(t *testing.T) {
		t.Parallel()

		ch := NewPollableChannel[int]()

		//nolint:staticcheck // intentional nil ctx to validate guard
		err := ch.Send(nil, NewMessage[int](1, nil))
		if !errors.Is(err, ErrContextNil) {
			t.Fatalf("expected ErrContextNil, got %v", err)
		}
	})

	t.Run("returns ErrClosed after Close", func(t *testing.T) {
		t.Parallel()

		ch := NewPollableChannel[int]()

		err := ch.Close()
		if err != nil {
			t.Fatalf("Close returned %v", err)
		}

		err = ch.Send(context.Background(), NewMessage[int](1, nil))
		if !errors.Is(err, ErrClosed) {
			t.Fatalf("expected ErrClosed, got %v", err)
		}
	})

	t.Run("returns ErrTimeout when ctx expires with buffer full", func(t *testing.T) {
		t.Parallel()

		ch := NewPollableChannel[int](WithBufferSize(1))

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

		if !errors.Is(err, ErrSendFailed) {
			t.Fatalf("expected ErrSendFailed, got %v", err)
		}
	})

	t.Run("succeeds when buffer has capacity", func(t *testing.T) {
		t.Parallel()

		ch := NewPollableChannel[int](WithBufferSize(4))

		err := ch.Send(context.Background(), NewMessage[int](7, nil))
		if err != nil {
			t.Fatalf("Send returned %v", err)
		}
	})
}

func TestPollableChannel_Receive(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrContextNil on nil ctx", func(t *testing.T) {
		t.Parallel()

		ch := NewPollableChannel[int]()

		//nolint:staticcheck // intentional nil ctx to validate guard
		_, err := ch.Receive(nil)
		if !errors.Is(err, ErrContextNil) {
			t.Fatalf("expected ErrContextNil, got %v", err)
		}
	})

	t.Run("returns sent message on happy path", func(t *testing.T) {
		t.Parallel()

		ch := NewPollableChannel[int](WithBufferSize(2))

		err := ch.Send(context.Background(), NewMessage[int](42, nil))
		if err != nil {
			t.Fatalf("Send returned %v", err)
		}

		msg, err := ch.Receive(context.Background())
		if err != nil {
			t.Fatalf("Receive returned %v", err)
		}

		if msg.Payload != 42 {
			t.Fatalf("expected payload 42, got %d", msg.Payload)
		}
	})

	t.Run("returns ErrTimeout when ctx expires while empty", func(t *testing.T) {
		t.Parallel()

		ch := NewPollableChannel[int]()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		_, err := ch.Receive(ctx)
		if !errors.Is(err, ErrTimeout) {
			t.Fatalf("expected ErrTimeout, got %v", err)
		}

		if !errors.Is(err, ErrReceiveFailed) {
			t.Fatalf("expected ErrReceiveFailed, got %v", err)
		}
	})

	t.Run("returns ErrChannelClosed after Close when buffer drained", func(t *testing.T) {
		t.Parallel()

		ch := NewPollableChannel[int]()

		err := ch.Close()
		if err != nil {
			t.Fatalf("Close returned %v", err)
		}

		_, err = ch.Receive(context.Background())
		if !errors.Is(err, ErrChannelClosed) {
			t.Fatalf("expected ErrChannelClosed, got %v", err)
		}

		if !errors.Is(err, ErrReceiveFailed) {
			t.Fatalf("expected ErrReceiveFailed, got %v", err)
		}
	})

	t.Run("drains buffered messages before reporting closed", func(t *testing.T) {
		t.Parallel()

		ch := NewPollableChannel[int](WithBufferSize(4))

		err := ch.Send(context.Background(), NewMessage[int](1, nil))
		if err != nil {
			t.Fatalf("Send returned %v", err)
		}

		err = ch.Send(context.Background(), NewMessage[int](2, nil))
		if err != nil {
			t.Fatalf("Send returned %v", err)
		}

		err = ch.Close()
		if err != nil {
			t.Fatalf("Close returned %v", err)
		}

		msg, err := ch.Receive(context.Background())
		if err != nil {
			t.Fatalf("Receive returned %v after Close", err)
		}

		if msg.Payload != 1 {
			t.Fatalf("expected payload 1, got %d", msg.Payload)
		}

		msg, err = ch.Receive(context.Background())
		if err != nil {
			t.Fatalf("Receive returned %v after Close", err)
		}

		if msg.Payload != 2 {
			t.Fatalf("expected payload 2, got %d", msg.Payload)
		}

		_, err = ch.Receive(context.Background())
		if !errors.Is(err, ErrChannelClosed) {
			t.Fatalf("expected ErrChannelClosed once drained, got %v", err)
		}
	})

	t.Run("concurrent producers and single consumer round-trip", func(t *testing.T) {
		t.Parallel()

		const producers = 8
		const perProducer = 25
		ch := NewPollableChannel[int](WithBufferSize(32))

		var wg sync.WaitGroup
		ctx := context.Background()

		for p := range producers {
			wg.Go(func() {
				for i := range perProducer {
					err := ch.Send(ctx, NewMessage[int](p*perProducer+i, nil))
					if err != nil {
						t.Errorf("Send returned %v", err)

						return
					}
				}
			})
		}

		seen := make(map[int]bool, producers*perProducer)
		var consumed atomic.Int32

		done := make(chan struct{})
		go func() {
			defer close(done)
			for {
				rcvCtx, cancel := context.WithTimeout(ctx, time.Second)
				msg, err := ch.Receive(rcvCtx)
				cancel()
				if err != nil {
					return
				}
				seen[msg.Payload] = true
				if consumed.Add(1) == int32(producers*perProducer) {
					return
				}
			}
		}()

		wg.Wait()

		<-done

		if len(seen) != producers*perProducer {
			t.Fatalf("expected %d unique messages, got %d", producers*perProducer, len(seen))
		}
	})

	t.Run("multiple receivers each get one message", func(t *testing.T) {
		t.Parallel()

		ch := NewPollableChannel[int](WithBufferSize(4))

		const receivers = 3

		var ready sync.WaitGroup
		ready.Add(receivers)

		var got atomic.Int32
		done := make(chan struct{}, receivers)

		for range receivers {
			go func() {
				ready.Done()
				_, err := ch.Receive(context.Background())
				if err != nil {
					t.Errorf("Receive returned %v", err)
				}
				got.Add(1)
				done <- struct{}{}
			}()
		}

		ready.Wait()

		// Allow receivers to park on Receive before producing.
		time.Sleep(10 * time.Millisecond)

		ctx := context.Background()
		for i := range receivers {
			err := ch.Send(ctx, NewMessage[int](i, nil))
			if err != nil {
				t.Fatalf("Send returned %v", err)
			}
		}

		for range receivers {
			select {
			case <-done:
			case <-time.After(time.Second):
				t.Fatal("timed out waiting for receiver")
			}
		}

		if got.Load() != receivers {
			t.Fatalf("expected %d deliveries, got %d", receivers, got.Load())
		}
	})
}

func TestPollableChannel_Close(t *testing.T) {
	t.Parallel()

	t.Run("idempotent across multiple calls", func(t *testing.T) {
		t.Parallel()

		ch := NewPollableChannel[int]()

		err := ch.Close()
		if err != nil {
			t.Fatalf("first Close returned %v", err)
		}

		err = ch.Close()
		if err != nil {
			t.Fatalf("second Close returned %v", err)
		}
	})

	t.Run("unblocks parked receivers", func(t *testing.T) {
		t.Parallel()

		ch := NewPollableChannel[int]()

		var wg sync.WaitGroup
		errCh := make(chan error, 1)

		wg.Go(func() {
			_, err := ch.Receive(context.Background())
			errCh <- err
		})

		// Allow the receiver to park.
		time.Sleep(10 * time.Millisecond)

		err := ch.Close()
		if err != nil {
			t.Fatalf("Close returned %v", err)
		}

		wg.Wait()

		err = <-errCh
		if !errors.Is(err, ErrChannelClosed) {
			t.Fatalf("expected ErrChannelClosed, got %v", err)
		}
	})
}
