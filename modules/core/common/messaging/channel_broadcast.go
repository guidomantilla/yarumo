package messaging

import (
	"context"
	"errors"
	"sync"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
)

// broadcastChannel implements Channel[T] as a synchronous parallel
// fan-out with barrier semantics: Send spawns one goroutine per
// subscriber, dispatches the message in parallel, and waits for ALL
// handlers to finish (success or failure) before returning. The
// returned error joins every handler's non-nil error via errors.Join.
// There is no fail-fast — every handler runs even if a previous one
// errored or panicked.
//
// broadcastChannel does not own a close lifecycle: handlers are
// spawned per Send and exit when their work completes; no permanent
// goroutines live on the channel.
//
// Use this primitive when the caller needs sync confirmation that
// every subscriber processed the message, but wants parallelism among
// subscribers (cf. PipelineChannel which is sync but serial).
type broadcast[T any] struct {
	mu     sync.RWMutex
	nextID uint64
	order  []uint64
	byID   map[uint64]Handler[T]
}

// NewBroadcastChannel creates a synchronous parallel Channel[T] with
// barrier semantics. Send blocks until every subscriber's handler has
// finished and returns the joined errors (nil when all handlers
// succeed).
func NewBroadcastChannel[T any]() Channel[T] {
	return &broadcast[T]{
		byID: map[uint64]Handler[T]{},
	}
}

// Send dispatches msg to every subscribed handler in parallel and
// waits for all of them to finish. Returns nil on full success, or
// ErrSend wrapping the joined errors of every failing handler
// (errors.Is matches each cause). Panics are recovered per handler
// and surface as ErrHandlerPanic-wrapped errors in the joined set.
// Returns ErrSend(ErrContextNil) when ctx is nil.
func (c *broadcast[T]) Send(ctx context.Context, msg Message[T]) error {
	if ctx == nil {
		return ErrSend(ErrContextNil)
	}

	cassert.NotNil(c, "broadcastChannel is nil")

	handlers := snapshotHandlers(&c.mu, &c.order, c.byID)
	if len(handlers) == 0 {
		return nil
	}

	errs := make([]error, len(handlers))

	var wg sync.WaitGroup

	for i, handler := range handlers {
		wg.Add(1)

		go func(idx int, h Handler[T]) {
			defer wg.Done()

			errs[idx] = invokeHandler(ctx, msg, h)
		}(i, handler)
	}

	wg.Wait()

	joined := errors.Join(errs...)
	if joined == nil {
		return nil
	}

	return ErrSend(joined)
}

// Subscribe registers handler at the end of the subscriber list and
// returns a Cancel that detaches it. Cancel is idempotent. Subscribe
// returns ErrSubscribe(ErrHandlerNil) when handler is nil.
//
// Subscribe order is preserved for the goroutine spawn pattern, but
// because handlers run in parallel the actual completion order is
// non-deterministic. The returned error from Send aggregates failures
// without reference to ordering.
func (c *broadcast[T]) Subscribe(handler Handler[T]) (Cancel, error) {
	cassert.NotNil(c, "broadcastChannel is nil")

	if handler == nil {
		return nil, ErrSubscribe(ErrHandlerNil)
	}

	c.mu.Lock()
	c.nextID++
	id := c.nextID
	c.byID[id] = handler
	c.order = append(c.order, id)
	c.mu.Unlock()

	var once sync.Once

	cancel := func() {
		once.Do(func() {
			c.mu.Lock()
			defer c.mu.Unlock()

			delete(c.byID, id)

			for i, candidate := range c.order {
				if candidate != id {
					continue
				}

				c.order = append(c.order[:i], c.order[i+1:]...)

				break
			}
		})
	}

	return cancel, nil
}
