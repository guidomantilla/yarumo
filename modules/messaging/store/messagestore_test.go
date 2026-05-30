package store

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"

	"github.com/guidomantilla/yarumo/messaging"
)

func TestNewInMemoryMessageStore(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil store", func(t *testing.T) {
		t.Parallel()

		s := NewInMemoryMessageStore[int]()
		if s == nil {
			t.Fatal("expected non-nil store")
		}
	})
}

func TestInMemoryMessageStore_Put(t *testing.T) {
	t.Parallel()

	t.Run("stores a message under key", func(t *testing.T) {
		t.Parallel()

		s := NewInMemoryMessageStore[int]()

		err := s.Put(context.Background(), "k", messaging.Message[int]{Payload: 42})
		if err != nil {
			t.Fatalf("put: %v", err)
		}

		got, err := s.Get(context.Background(), "k")
		if err != nil {
			t.Fatalf("get: %v", err)
		}

		if got.Payload != 42 {
			t.Fatalf("payload: got %d want 42", got.Payload)
		}
	})

	t.Run("overwrites previous value at same key", func(t *testing.T) {
		t.Parallel()

		s := NewInMemoryMessageStore[int]()

		err := s.Put(context.Background(), "k", messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("first put: %v", err)
		}

		err = s.Put(context.Background(), "k", messaging.Message[int]{Payload: 2})
		if err != nil {
			t.Fatalf("second put: %v", err)
		}

		got, err := s.Get(context.Background(), "k")
		if err != nil {
			t.Fatalf("get: %v", err)
		}

		if got.Payload != 2 {
			t.Fatalf("payload: got %d want 2 (overwrite expected)", got.Payload)
		}
	})

	t.Run("preserves headers end-to-end", func(t *testing.T) {
		t.Parallel()

		s := NewInMemoryMessageStore[string]()

		msg := messaging.Message[string]{
			Payload: "hello",
			Headers: messaging.Headers{
				MessageID:     "msg-1",
				CorrelationID: "corr-1",
				Type:          "greeting",
			},
		}

		err := s.Put(context.Background(), "k", msg)
		if err != nil {
			t.Fatalf("put: %v", err)
		}

		got, err := s.Get(context.Background(), "k")
		if err != nil {
			t.Fatalf("get: %v", err)
		}

		if got.Headers.MessageID != "msg-1" {
			t.Fatalf("MessageID: got %q", got.Headers.MessageID)
		}

		if got.Headers.CorrelationID != "corr-1" {
			t.Fatalf("CorrelationID: got %q", got.Headers.CorrelationID)
		}

		if got.Headers.Type != "greeting" {
			t.Fatalf("Type: got %q", got.Headers.Type)
		}
	})

	t.Run("returns ErrStore when ctx is already cancelled", func(t *testing.T) {
		t.Parallel()

		s := NewInMemoryMessageStore[int]()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := s.Put(ctx, "k", messaging.Message[int]{Payload: 1})
		if !errors.Is(err, ErrStoreFailed) {
			t.Fatalf("expected ErrStoreFailed, got %v", err)
		}

		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected wrapped context.Canceled, got %v", err)
		}
	})
}

func TestInMemoryMessageStore_Get(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrNotFound when key absent", func(t *testing.T) {
		t.Parallel()

		s := NewInMemoryMessageStore[int]()

		_, err := s.Get(context.Background(), "missing")
		if !errors.Is(err, ErrStoreNotFound) {
			t.Fatalf("expected ErrStoreNotFound, got %v", err)
		}
	})

	t.Run("returns the stored message after Put", func(t *testing.T) {
		t.Parallel()

		s := NewInMemoryMessageStore[int]()

		err := s.Put(context.Background(), "k", messaging.Message[int]{Payload: 7})
		if err != nil {
			t.Fatalf("put: %v", err)
		}

		got, err := s.Get(context.Background(), "k")
		if err != nil {
			t.Fatalf("get: %v", err)
		}

		if got.Payload != 7 {
			t.Fatalf("payload: got %d want 7", got.Payload)
		}
	})

	t.Run("returns ErrStore when ctx is already cancelled", func(t *testing.T) {
		t.Parallel()

		s := NewInMemoryMessageStore[int]()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := s.Get(ctx, "k")
		if !errors.Is(err, ErrStoreFailed) {
			t.Fatalf("expected ErrStoreFailed, got %v", err)
		}
	})
}

func TestInMemoryMessageStore_Delete(t *testing.T) {
	t.Parallel()

	t.Run("removes the message at key", func(t *testing.T) {
		t.Parallel()

		s := NewInMemoryMessageStore[int]()

		err := s.Put(context.Background(), "k", messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("put: %v", err)
		}

		err = s.Delete(context.Background(), "k")
		if err != nil {
			t.Fatalf("delete: %v", err)
		}

		_, err = s.Get(context.Background(), "k")
		if !errors.Is(err, ErrStoreNotFound) {
			t.Fatalf("expected ErrStoreNotFound after delete, got %v", err)
		}
	})

	t.Run("is idempotent on missing key", func(t *testing.T) {
		t.Parallel()

		s := NewInMemoryMessageStore[int]()

		err := s.Delete(context.Background(), "missing")
		if err != nil {
			t.Fatalf("expected nil on missing key, got %v", err)
		}

		err = s.Delete(context.Background(), "missing")
		if err != nil {
			t.Fatalf("expected nil on second delete, got %v", err)
		}
	})

	t.Run("returns ErrStore when ctx is already cancelled", func(t *testing.T) {
		t.Parallel()

		s := NewInMemoryMessageStore[int]()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := s.Delete(ctx, "k")
		if !errors.Is(err, ErrStoreFailed) {
			t.Fatalf("expected ErrStoreFailed, got %v", err)
		}
	})
}

func TestInMemoryMessageStore_Concurrent(t *testing.T) {
	t.Parallel()

	t.Run("Put + Get + Delete are race-free", func(t *testing.T) {
		t.Parallel()

		s := NewInMemoryMessageStore[int]()

		const workers = 16
		const ops = 200

		var wg sync.WaitGroup

		for w := range workers {
			wg.Add(1)

			go func(id int) {
				defer wg.Done()

				for i := range ops {
					key := strconv.Itoa(id) + "-" + strconv.Itoa(i)

					err := s.Put(context.Background(), key, messaging.Message[int]{Payload: i})
					if err != nil {
						t.Errorf("put: %v", err)

						return
					}

					_, err = s.Get(context.Background(), key)
					if err != nil {
						t.Errorf("get: %v", err)

						return
					}

					err = s.Delete(context.Background(), key)
					if err != nil {
						t.Errorf("delete: %v", err)

						return
					}
				}
			}(w)
		}

		wg.Wait()
	})
}
