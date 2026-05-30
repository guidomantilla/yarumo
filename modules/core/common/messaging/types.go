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
// # Context propagation
//
// The ctx the Handler receives depends on the channel's dispatch
// model:
//
//   - Synchronous channels (PipelineChannel, BroadcastChannel) run
//     handlers in the caller's goroutine and pass the ctx given to
//     Send straight through. Deadline, cancellation and values from
//     the publisher all reach the handler.
//   - Asynchronous channels (TopicChannel, QueueChannel) decouple
//     publisher and consumer lifetimes. The handler ctx merges the
//     worker's lifecycle ctx (the one passed to Start) with the
//     publisher's Send ctx: Done / Deadline / Err follow the worker
//     so publisher cancellation does NOT abort in-flight handlers,
//     but Value lookups fall through to the Send ctx so trace span,
//     correlation id and slogctx attributes propagate from the
//     publisher to the handler. Use Headers.CorrelationID for
//     request-to-handler correlation when ctx-based cancellation
//     propagation is undesirable for the async pattern.
//
// # Overflow policy (async channels)
//
// TopicChannel and QueueChannel honor a configurable OverflowPolicy
// (see WithOverflowPolicy) selecting what happens when Send finds the
// internal buffer at capacity:
//
//   - OverflowReject (the default for NewOptions): Send returns
//     ErrSend(ErrBufferFull) immediately. The caller decides — retry,
//     shed, fallback — and saturation is loud rather than silent.
//   - OverflowBlock: Send blocks until a slot opens or the caller's
//     ctx expires; the historical behavior, useful when message loss
//     is unacceptable and the publisher can absorb backpressure.
//   - OverflowDropNewest / OverflowDropOldest: Send returns nil and the
//     ErrorHandler hook fires with ErrOverflow joined with ErrDropped;
//     useful for telemetry / metrics where eviction is acceptable.
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
	_ Channel[any] = (*null[any])(nil)

	_ ErrorHandler = DefaultErrorHandler
	_ ErrorHandler = SilentErrorHandler
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

// ErrorHandler is the function type for the per-handler error
// observability hook installed on a TopicChannel, QueueChannel, or
// NullChannel via WithErrorHandler.
//
// The hook fires once per failed handler invocation, after the
// dispatcher has recovered any panic. err carries the handler's
// returned error or, on panic, an error wrapping ErrHandlerPanic with
// the recovered value. msg is type-erased; cast it inside the hook
// when payload-specific behavior is needed. The hook is invoked from
// the worker goroutine and must not block — long observability work
// should be dispatched asynchronously by the implementer.
//
// The default hook (DefaultErrorHandler) logs every failure via
// common/log so a consumer that forgets to wire observability still
// gets a record of handler errors. Callers that genuinely want silence
// must opt out by installing SilentErrorHandler explicitly.
type ErrorHandler func(ctx context.Context, msg any, err error)

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
//
// Context propagation: see the package doc. Sync impls forward the
// Send ctx to each handler; async impls forward a ctx whose lifecycle
// follows the worker and whose Value lookups fall through to the Send
// ctx (so trace spans and slogctx values propagate but publisher
// cancellation does not abort in-flight handlers).
type Channel[T any] interface {
	// Send dispatches msg to all currently subscribed handlers. The
	// returned error reflects the dispatch outcome per implementation:
	// PipelineChannel propagates the first handler error; TopicChannel
	// returns ErrClosed if the worker is no longer accepting work, or
	// nil after successful enqueue. ctx gates the enqueue/dispatch
	// step; how it reaches the Handler depends on the implementation
	// (see package doc).
	Send(ctx context.Context, msg Message[T]) error
	// Subscribe registers handler and returns a Cancel that detaches
	// it. Cancel is idempotent. Subscribe returns an error if the
	// channel is closed or if handler is nil.
	Subscribe(handler Handler[T]) (Cancel, error)
}
