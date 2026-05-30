// Package bridge provides a one-to-one channel forwarder over
// messaging.Channel[T].
//
// A Bridge subscribes to a source Channel[T] and republishes every
// received Message[T] to a single destination Channel[T] without
// alteration. The pattern is intentionally the identity transform — its
// value is structural, not logical:
//
//   - **Sync ↔ async decoupling**: read from a BroadcastChannel
//     (synchronous, caller waits) and forward to a TopicChannel
//     (asynchronous, fire-and-forget) so producers do not pay the
//     latency of downstream consumers.
//   - **Wiring graph labels**: surface a named lifecycle.Component for
//     each src→dst hop so dashboards, logs and architecture diagrams
//     can talk about "the orders bridge" instead of an anonymous
//     Subscribe closure.
//   - **Cross-cutting hook point**: the Bridge owns the error
//     observability (WithErrorHandler) for the hop. When forward Send
//     fails, the bridge surfaces it through a single configured hook
//     instead of leaving every Subscribe-caller to re-implement
//     observability.
//
// If a hop needs to mutate the payload, drop messages by predicate, or
// route to multiple destinations, use the dedicated patterns
// (transformer, filter, router) — not Bridge with extra options.
//
// # Lifecycle
//
// Bridge implements common/lifecycle.Component (worker-style): Start
// registers the subscription on the source channel and returns
// immediately; Stop cancels the subscription and closes Done. Bridge
// does not spawn goroutines of its own — dispatch concurrency is
// inherited from the source channel implementation. Wire it via
// lifecycle.Build for the standard daemon CloseFn pattern.
//
// # Error handling
//
// The handler installed on the source channel always returns nil.
// Forward Send failures are surfaced via the Bridge's own ErrorHandler
// (installed with WithErrorHandler, defaulting to
// messaging.DefaultErrorHandler which logs via common/log). This keeps
// bridge concerns out of the source channel's caller error path —
// consistent with the package-wide policy in modules/integration/
// CODING_STANDARDS.md.
package bridge

import (
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
)

var (
	_ lifecycle.Component = (*bridge[any])(nil)

	_ ErrBridgeFn = ErrBridge
)

// ErrBridgeFn is the function type for ErrBridge.
type ErrBridgeFn func(causes ...error) error
