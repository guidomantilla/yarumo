// Package filter provides a Message Filter pattern over
// messaging.Channel[T].
//
// A Filter subscribes to a source Channel[T] and forwards each received
// Message[T] to a single destination Channel[T] only when a
// user-supplied PredicateFn returns true. Messages where the predicate
// returns false are dropped from the flow.
//
// # Two observability hooks (shape B)
//
// Filters distinguish two outcomes that look identical from the
// caller's perspective but mean very different things operationally:
//
//   - **Intentional drops** (`predicate(msg) == false`): the filter
//     did its job. These are routed to the optional DropHandler hook
//     (installed via WithDropHandler). Default is nil — silent drop —
//     because a noisy filter producing "I dropped a message" logs for
//     every gated message is rarely useful. Wire WithDropHandler when
//     you need throughput metrics ("how many passed / how many were
//     gated") or audit trails.
//   - **Real failures** (predicate returned error, predicate panicked,
//     forward Send failed): the filter could not do its job. These
//     are routed to the ErrorHandler (installed via WithErrorHandler,
//     defaulting to messaging.DefaultErrorHandler which logs via
//     common/log). The default ensures failures never disappear
//     silently.
//
// Keeping the two separate lets you ship metrics counters for drops
// without polluting error dashboards, and lets ops alert on real
// failures without false positives from normal gating.
//
// # Lifecycle
//
// Filter implements common/lifecycle.Component (worker-style): Start
// registers the subscription on the source channel and returns
// immediately; Stop cancels the subscription and closes Done. Filter
// does not spawn goroutines of its own — dispatch concurrency is
// inherited from the source channel implementation.
//
// # Error handling
//
// The handler installed on the source channel always returns nil.
// Filter concerns (predicate error/panic, forward Send failure) flow
// through WithErrorHandler; intentional drops flow through
// WithDropHandler. Nothing propagates to the source channel's Send
// caller error path, consistent with the package-wide policy in
// modules/integration/CODING_STANDARDS.md.
package filter

import (
	"context"

	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/core/common/messaging"
)

var (
	_ lifecycle.Component = (*filter[any])(nil)

	_ ErrFilterFn = ErrFilter
)

// PredicateFn returns true to forward msg to the destination channel
// and false to drop it. An error returned by PredicateFn is wrapped in
// ErrFilter(ErrPredicateFailed, err) and forwarded to the ErrorHandler;
// the message is treated as dropped. A panic in PredicateFn is
// recovered and wrapped in ErrFilter(ErrPredicatePanic, ...) so it
// cannot kill the source channel's dispatcher.
type PredicateFn[T any] func(ctx context.Context, msg messaging.Message[T]) (bool, error)

// DropHandler is the optional observability hook invoked once per
// intentional drop (predicate returned false). msg is type-erased; cast
// it inside the hook when payload-specific behavior is needed. The
// hook is invoked synchronously from the source channel's dispatcher
// and must not block — long observability work should be dispatched
// asynchronously by the implementer.
//
// DropHandler is NOT invoked when the predicate errors or panics
// (those are routed to the ErrorHandler instead) — DropHandler fires
// only on successful, deliberate drops.
type DropHandler func(ctx context.Context, msg any)

// ErrFilterFn is the function type for ErrFilter.
type ErrFilterFn func(causes ...error) error
