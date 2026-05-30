package messaging

import (
	"context"
	"sync"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
)

// pipeline implements Channel[T] as a Transactional Handler
// Chain: Send invokes every subscribed handler synchronously, in the
// caller's goroutine, in registration order. The first handler to
// return a non-nil error (or to panic) aborts the chain — subsequent
// steps are reported as skipped in the returned *ChainError trace.
//
// pipeline does not own a close lifecycle: there is nothing to
// drain and no goroutines to stop. Channels that need a graceful
// shutdown use the TopicChannel variant.
type pipeline[T any] struct {
	mu     sync.RWMutex
	nextID uint64
	order  []uint64
	byID   map[uint64]Handler[T]
}

// NewPipelineChannel creates a synchronous Channel[T] that dispatches
// messages through subscribers in Subscribe order, fail-fast, with a
// per-step trace exposed via *ChainError on failure.
//
// Use this primitive for in-process side effects that must complete
// (or fail together) before the caller proceeds — audit logs joining
// the caller's transaction, cache invalidation, metrics that must be
// flushed before the response, or a "bridge to async" step that hands
// the message off to a TopicChannel.
func NewPipelineChannel[T any]() Channel[T] {
	return &pipeline[T]{
		byID: map[uint64]Handler[T]{},
	}
}

// Send dispatches msg sequentially through every subscribed handler
// in registration order. The dispatch is fail-fast: the first step
// that returns a non-nil error (or panics) aborts the chain.
//
// Returns:
//   - nil when every step completed without error.
//   - *ChainError wrapped in ErrSend when at least one step failed.
//     The ChainError carries the full step trace so callers can render
//     which steps ran, which one broke, and which never executed.
//   - ErrSend(ErrContextNil) when ctx is nil.
//
// Panics inside handlers are recovered, converted to a StepResult
// with Status StepStatusPanic, and reported through the same
// *ChainError flow. They never propagate to the caller.
func (c *pipeline[T]) Send(ctx context.Context, msg Message[T]) error {
	if ctx == nil {
		return ErrSend(ErrContextNil)
	}

	cassert.NotNil(c, "pipeline is nil")

	handlers := snapshotHandlers(&c.mu, &c.order, c.byID)

	steps := make([]StepResult, len(handlers))
	failed := -1

	for i, handler := range handlers {
		if failed >= 0 {
			steps[i] = StepResult{Index: i, Status: StepStatusSkipped}

			continue
		}

		steps[i] = invokeStep(ctx, msg, i, handler)
		if steps[i].Status != StepStatusOK {
			failed = i
		}
	}

	if failed < 0 {
		return nil
	}

	return ErrSend(&ChainError{Steps: steps, Failed: failed}, ErrChainFailed)
}

// Subscribe registers handler at the end of the chain and returns a
// Cancel that detaches it. Cancel is idempotent. Subscribe returns
// ErrSubscribe(ErrHandlerNil) when handler is nil.
//
// Handlers run in Subscribe order: the first Subscribe call becomes
// step 0, the second becomes step 1, and so on. Cancelling a handler
// does not renumber the remaining steps — subsequent traces simply
// omit the cancelled handler.
func (c *pipeline[T]) Subscribe(handler Handler[T]) (Cancel, error) {
	cassert.NotNil(c, "pipeline is nil")

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
