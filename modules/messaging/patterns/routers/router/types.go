// Package router provides a Content-Based Router pattern over
// messaging.Channel[T].
//
// A Router subscribes to a source Channel[T] and forwards each received
// Message[T] to one of N destination channels chosen by a user-supplied
// RouteFn. The pattern decouples publishers from destination logic:
// publishers send to a single input channel, while routing rules live
// declaratively in a (key → channel) map.
//
// # Lifecycle
//
// Router implements common/lifecycle.Component (worker-style): Start
// registers the subscription on the source channel and returns
// immediately; Stop cancels the subscription and closes Done. The
// Router does not spawn goroutines of its own — dispatch concurrency
// is inherited from the source channel implementation. Wire it via
// lifecycle.Build for the standard daemon CloseFn pattern.
//
// # Error handling
//
// The handler installed on the source channel always returns nil.
// Routing failures (RouteFn error or panic, no matching key, forward
// Send failure) are surfaced via the Router's own ErrorHandler
// (installed with WithErrorHandler, defaulting to
// messaging.DefaultErrorHandler which logs via common/log). This keeps
// routing concerns out of the source channel's caller error path.
//
// # NoRoute policy
//
// When RouteFn returns a key that is absent from the routes map and no
// WithDefaultChannel option was passed, the Router forwards
// ErrRoute(ErrNoRoute, ...) to the ErrorHandler and drops the message.
// When WithDefaultChannel is set, the message is forwarded to that
// channel instead and the ErrorHandler is not invoked unless the
// default channel's Send itself fails.
//
// # Variants on the EIP catalog
//
// The package implements the canonical Content-Based Router. Related
// patterns sit elsewhere:
//
//   - Filter — predicate gate (planned, future sub-package).
//   - Recipient List — fan-out to multiple destinations (planned).
//   - Dynamic Router — runtime-mutable route table (planned).
package router

import (
	"context"

	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/messaging"
)

var (
	_ Router[any] = (*router[any])(nil)

	_ ErrRouteFn = ErrRoute
)

// Router is the public interface for a Content-Based Router. It
// embeds lifecycle.Component so callers wire it up with
// lifecycle.Build. The interface exists (rather than returning
// lifecycle.Component directly) so the consumer's API surface
// preserves "this is a Router" semantics and the type stays open to
// future router-specific methods without breaking callers.
type Router[T any] interface {
	lifecycle.Component
}

// RouteFn returns the destination key for msg, or an error. The key is
// looked up in the routes map provided to NewRouter; a missing key
// triggers the NoRoute path (WithDefaultChannel or the ErrorHandler).
// An error returned by RouteFn is wrapped in ErrRoute(ErrRouteFnFailed,
// err) and forwarded to the ErrorHandler. A panic in RouteFn is
// recovered and wrapped in ErrRoute(ErrRoutePanic, ...) so it cannot
// kill the source channel's dispatcher.
type RouteFn[T any] func(ctx context.Context, msg messaging.Message[T]) (string, error)

// ErrRouteFn is the function type for ErrRoute.
type ErrRouteFn func(causes ...error) error
