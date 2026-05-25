package messaging

import (
	"context"
	"sync"

	cassert "github.com/guidomantilla/yarumo/common/assert"
)

// directChannel implements Channel[T] with synchronous in-goroutine
// dispatch: Send invokes every subscribed handler on the caller's
// goroutine, in registration order. The first handler error is
// returned via ErrSend; subsequent handlers are not invoked.
//
// directChannel does not own a close lifecycle: there is nothing to
// drain and no goroutines to stop. Channels that need a graceful
// shutdown use the QueueChannel variant.
type directChannel[T any] struct {
	mu       sync.RWMutex
	nextID   uint64
	handlers map[uint64]Handler[T]
}

// NewDirectChannel creates a synchronous Channel[T]. Subscribe attaches
// handlers; Send invokes each attached handler in the caller's
// goroutine. The channel is safe for concurrent use.
func NewDirectChannel[T any]() Channel[T] {
	return &directChannel[T]{
		handlers: map[uint64]Handler[T]{},
	}
}

// Send dispatches msg to all currently subscribed handlers
// synchronously, in the caller's goroutine. It returns the first
// non-nil handler error wrapped in ErrSend, or nil when all handlers
// returned nil. Returns ErrSend(ErrContextNil) when ctx is nil.
func (c *directChannel[T]) Send(ctx context.Context, msg Message[T]) error {
	cassert.NotNil(c, "directChannel is nil")

	if ctx == nil {
		return ErrSend(ErrContextNil)
	}

	snapshot := c.snapshot()

	for _, handler := range snapshot {
		err := handler(ctx, msg)
		if err != nil {
			return ErrSend(err)
		}
	}

	return nil
}

// Subscribe registers handler and returns a Cancel that detaches it.
// Cancel is idempotent. Subscribe returns ErrSubscribe(ErrHandlerNil)
// when handler is nil.
func (c *directChannel[T]) Subscribe(handler Handler[T]) (Cancel, error) {
	cassert.NotNil(c, "directChannel is nil")

	if handler == nil {
		return nil, ErrSubscribe(ErrHandlerNil)
	}

	c.mu.Lock()
	c.nextID++
	id := c.nextID
	c.handlers[id] = handler
	c.mu.Unlock()

	var once sync.Once

	cancel := func() {
		once.Do(func() {
			c.mu.Lock()
			delete(c.handlers, id)
			c.mu.Unlock()
		})
	}

	return cancel, nil
}

// snapshot returns a stable copy of the handler list. It exists so
// the Send dispatch loop holds the read lock for as little time as
// possible — handlers should not be invoked while holding any lock.
func (c *directChannel[T]) snapshot() []Handler[T] {
	c.mu.RLock()
	defer c.mu.RUnlock()

	out := make([]Handler[T], 0, len(c.handlers))
	for _, handler := range c.handlers {
		out = append(out, handler)
	}

	return out
}
