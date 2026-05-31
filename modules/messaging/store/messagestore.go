package store

import (
	"context"
	"sync"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	cpointer "github.com/guidomantilla/yarumo/core/common/pointer"
	"github.com/guidomantilla/yarumo/messaging"
)

// inMemoryMessageStore is the canonical MessageStore[T] backed by a
// mutex-guarded map. It is purely passive — no goroutines, no
// lifecycle. The store overwrites on Put and is idempotent on Delete.
type inMemoryMessageStore[T any] struct {
	mu   sync.RWMutex
	data map[string]messaging.Message[T]
}

// NewInMemoryMessageStore constructs an in-memory MessageStore[T].
// The returned store is ready to use immediately; it holds no
// external resources and does not implement lifecycle.Component.
// Heavy-dep backends (Redis, Postgres, S3, …) live in
// extension/messaging/store/<backend>/.
func NewInMemoryMessageStore[T any]() MessageStore[T] {
	return &inMemoryMessageStore[T]{
		data: map[string]messaging.Message[T]{},
	}
}

// Put stores msg under key. It overwrites any value previously
// stored at the same key. ctx is honored only for its cancellation
// signal — if ctx is already expired the call returns the
// corresponding ctx error wrapped in ErrStore.
func (s *inMemoryMessageStore[T]) Put(ctx context.Context, key string, msg messaging.Message[T]) error {
	cassert.NotNil(s, "in-memory message store is nil")

	err := ctx.Err()
	if err != nil {
		return ErrStore(err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = msg

	return nil
}

// Get returns the message previously stored at key. It returns
// ErrNotFound when the key does not exist, matchable via
// errors.Is(err, ErrStoreNotFound).
func (s *inMemoryMessageStore[T]) Get(ctx context.Context, key string) (messaging.Message[T], error) {
	cassert.NotNil(s, "in-memory message store is nil")

	err := ctx.Err()
	if err != nil {
		return cpointer.Zero[messaging.Message[T]](), ErrStore(err)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	msg, ok := s.data[key]
	if !ok {
		return cpointer.Zero[messaging.Message[T]](), ErrNotFound()
	}

	return msg, nil
}

// Delete removes the message stored at key. Delete is idempotent: it
// returns nil whether or not the key was present before the call.
func (s *inMemoryMessageStore[T]) Delete(ctx context.Context, key string) error {
	cassert.NotNil(s, "in-memory message store is nil")

	err := ctx.Err()
	if err != nil {
		return ErrStore(err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, key)

	return nil
}
