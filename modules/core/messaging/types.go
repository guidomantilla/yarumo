// Package messaging provides typed in-process messaging primitives.
//
// The package centers on a generic Channel[T] abstraction that delivers
// Message[T] envelopes to registered handlers. Four channel
// implementations are provided in-process:
//
//   - PipelineChannel[T]: synchronous, sequential fan-out in the
//     caller's goroutine; fail-fast with a per-step ChainError trace.
//     Use for transactional handler chains where steps must commit
//     or abort together.
//   - BroadcastChannel[T]: synchronous, parallel fan-out — Send
//     spawns one goroutine per subscriber and waits at a barrier for
//     all of them to finish. Returns the joined errors of every
//     failing handler. Use when the caller needs sync confirmation
//     of all subscribers, but wants parallelism between them.
//   - TopicChannel[T]: asynchronous, buffered fan-out via a worker
//     goroutine. Send returns immediately; the worker dispatches each
//     message to every subscriber serially. Implements common/
//     lifecycle.Component with graceful drain on Stop.
//   - QueueChannel[T]: asynchronous, point-to-point distribution —
//     each message is delivered to EXACTLY ONE subscriber via round-
//     robin among the registered handlers. Buffered with a worker
//     pool (WithWorkerCount). Implements lifecycle.Component.
//     Use for work distribution among equivalent workers.
//
// Concurrency: all public types in this package are safe for concurrent
// use by multiple goroutines.
//
// Scope: in-process only. Broker drivers and outbox patterns will be
// added later under the same Channel[T] shape; this module owns no
// external transport dependencies beyond the standard library.
package messaging

import (
	"context"
)

var (
	_ Channel[any] = (*pipeline[any])(nil)
	_ Channel[any] = (*broadcast[any])(nil)
	_ Channel[any] = (*topic[any])(nil)
	_ Channel[any] = (*queue[any])(nil)
)

// Handler is the function type for a message handler. The Handler
// receives the propagated context and the typed Message envelope and
// returns an error to signal failure. PipelineChannel propagates the
// error to the Send caller; TopicChannel logs and continues.
type Handler[T any] func(ctx context.Context, msg Message[T]) error

// Cancel is the function type returned by Subscribe. Invoking Cancel
// detaches the handler from the channel. Cancel is idempotent: calling
// it more than once is safe and is a no-op after the first call.
type Cancel func()

// Channel defines the contract for an in-process typed message channel.
//
// Implementations dispatch published Message[T] envelopes to all
// subscribed handlers. The dispatch flavor (synchronous in-caller-
// goroutine vs. asynchronous via a worker) is implementation-defined
// and documented on each concrete type.
//
// Implementations must be safe for concurrent use by multiple
// goroutines. Send must return ErrClosed (matched via errors.Is) when
// invoked after the channel has been closed; Subscribe must return the
// same error on closed channels.
type Channel[T any] interface {
	// Send dispatches msg to all currently subscribed handlers. The
	// returned error reflects the dispatch outcome per implementation:
	// PipelineChannel propagates the first handler error; TopicChannel
	// returns ErrClosed if the worker is no longer accepting work, or
	// nil after successful enqueue. ctx propagates to each Handler.
	Send(ctx context.Context, msg Message[T]) error
	// Subscribe registers handler and returns a Cancel that detaches
	// it. Cancel is idempotent. Subscribe returns an error if the
	// channel is closed or if handler is nil.
	Subscribe(handler Handler[T]) (Cancel, error)
}
