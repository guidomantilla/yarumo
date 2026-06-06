// Package gateway provides a Messaging Gateway pattern over
// messaging.Channel[Req] / messaging.Channel[Res] pairs.
//
// A Gateway exposes a synchronous request-reply API
// (Request(ctx, req) (res, error)) on top of asynchronous messaging.
// The caller invokes Request; the Gateway publishes the request to
// requestChan with an auto-generated CorrelationID and Headers.ReplyTo
// = gateway name, then waits for a Message[Res] on replyChan whose
// Headers.CorrelationID matches. The first matching reply is returned
// to the caller; a timeout or ctx cancellation aborts the wait.
//
// # Downstream contract
//
// The downstream consumer (whoever subscribes to requestChan to do the
// real work) MUST respect Headers.ReplyTo and Headers.CorrelationID and
// publish its response to the named reply channel with the same
// CorrelationID echoed back. Without this contract the Gateway cannot
// correlate responses and every Request will time out.
//
// Replies arriving for an unknown CorrelationID are silently dropped
// (they came too late, after the originating Request returned). This
// keeps the gateway resilient to slow downstreams that occasionally
// reply after the caller has given up.
//
// # Correlation
//
// CorrelationIDs are generated via the cuids.UID generator passed to
// NewGateway (typically uuids.V4). The internal pending-request map is
// guarded by a mutex; concurrent Requests are safe and each gets its
// own waiter channel.
//
// # Lifecycle
//
// Gateway implements common/lifecycle.Component (worker-style): Start
// subscribes to replyChan and returns immediately; Stop cancels the
// subscription, fails every pending Request with
// ErrGatewayShuttingDown, and closes Done.
//
// # Error handling
//
// The handler installed on the reply channel always returns nil.
// Real failures (request-channel Send fail) are surfaced via the
// Gateway's own ErrorHandler (installed with WithErrorHandler, defaulting
// to messaging.DefaultErrorHandler which logs via common/log).
package gateway

import (
	"context"
	"time"

	"github.com/guidomantilla/yarumo/core/common/lifecycle"
)

// DefaultRequestTimeout is the default per-Request timeout applied when
// WithRequestTimeout is not configured and the caller's ctx has no
// deadline. The per-request ctx deadline always takes precedence when
// it is tighter than this value.
const DefaultRequestTimeout = 5 * time.Second

var (
	_ Gateway[any, any] = (*gateway[any, any])(nil)

	_ ErrGatewayFn = ErrGateway
)

// Gateway is the public interface for a Messaging Gateway. It embeds
// lifecycle.Component so callers wire it up with lifecycle.Build, and
// exposes Request as the synchronous request-reply entry point.
type Gateway[Req, Res any] interface {
	lifecycle.Component
	// Request publishes req to the configured request channel with an
	// auto-generated CorrelationID and Headers.ReplyTo = the Gateway's
	// name, then waits for a Message[Res] on the reply channel whose
	// Headers.CorrelationID matches. The first matching reply is
	// returned. Request honours the caller's ctx (cancellation /
	// deadline) and the configured per-request timeout (whichever
	// fires first). A timeout returns the zero Res and ErrRequestTimeout;
	// ctx cancellation returns the zero Res and ctx.Err() wrapped in
	// ErrRequestCancelled; gateway shutdown returns the zero Res and
	// ErrGatewayShuttingDown.
	Request(ctx context.Context, req Req) (Res, error)
}

// ErrGatewayFn is the function type for ErrGateway.
type ErrGatewayFn func(causes ...error) error
