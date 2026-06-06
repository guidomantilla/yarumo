// Package delayer provides a Delayer pattern over messaging.Channel[T].
//
// A Delayer subscribes to a source Channel[T] and forwards each received
// Message[T] to a single destination Channel[T] after a configurable
// delay. Three strategies select how long each message waits before
// being forwarded; the first one matched wins:
//
//  1. **Fixed delay** (WithFixedDelay) — every message waits the same
//     constant duration. Useful for rate smoothing or coarse "give the
//     downstream a moment" hand-offs.
//  2. **Per-message delay** (WithDelayFn) — caller supplies a function
//     that computes the delay from the message (payload, headers, time
//     of day). Useful when the delay is data-driven, for example
//     deferring a retry by exponential backoff or honouring a
//     scheduled-at header from an upstream system.
//  3. **Headers.ExpirationTime fallback** — when neither of the above
//     is configured, the delayer uses Headers.ExpirationTime as the
//     deliver-at deadline. An ExpirationTime in the past forwards
//     immediately; an unset (zero) ExpirationTime also forwards
//     immediately.
//
// A computed delay of zero or less always forwards immediately on the
// same dispatcher goroutine as the source subscription — the deferred
// path is skipped entirely.
//
// # Composition over an internal ScheduledChannel
//
// Under the hood the delayer composes a messaging.ScheduledChannel[T]:
// the source-channel handler computes deliverAt, calls SendAfter on the
// internal scheduled channel, and an internal subscriber on the
// scheduled channel forwards each due message to the real destination.
// The min-heap, timer arming and worker goroutine all come from the
// scheduled primitive; the delayer adds the policy layer (strategy
// selection + max-pending bound + observability hooks) and the
// lifecycle wrapper around it.
//
// # Bounded pending queue
//
// WithMaxPending caps the number of messages currently in flight (i.e.
// scheduled but not yet delivered). When the bound is reached, new
// messages from the source are dropped — they do NOT block the source
// channel's dispatcher and they do NOT push back on the publisher. The
// drop is reported through WithDropHandler (silent default) with
// ErrMaxPendingExceeded so observability can count it. The default
// bound is defaultMaxPending; pass WithMaxPending(0) or a negative value
// and the default is preserved.
//
// # Lifecycle
//
// Delayer implements common/lifecycle.Component (worker-style): Start
// boots the internal ScheduledChannel, registers an internal subscriber
// on it (forwarding to the real destination), and registers the
// source-channel subscription. Stop cancels the source subscription,
// stops the internal ScheduledChannel and closes Done. Pending
// undelivered messages are dropped on Stop — schedule semantics are
// best-effort, matching the underlying ScheduledChannel contract.
//
// # Error handling
//
// The handler installed on the source channel always returns nil.
// Delayer concerns (schedule failure, forward Send failure) flow
// through WithErrorHandler; intentional drops on WithMaxPending flow
// through WithDropHandler. Nothing propagates to the source channel's
// Send caller error path, consistent with the package-wide policy in
// modules/messaging/CODING_STANDARDS.md.
package delayer

import (
	"context"
	"time"

	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/messaging"
)

var (
	_ Delayer[any] = (*delayer[any])(nil)

	_ ErrDelayerFn = ErrDelayer
)

// Delayer is the public interface for a Delayer pattern. It embeds
// lifecycle.Component so callers wire it up with lifecycle.Build. The
// interface exists (rather than returning lifecycle.Component directly)
// so the consumer's API surface preserves "this is a Delayer"
// semantics and the type stays open to future delayer-specific methods
// without breaking callers.
type Delayer[T any] interface {
	lifecycle.Component
}

// DelayFn computes the delay applied to msg before forwarding it to the
// destination. The function is invoked synchronously from the source
// channel's dispatcher; long computations should be avoided. A return
// value of zero or negative duration is treated as "forward
// immediately" and bypasses the internal scheduler entirely.
type DelayFn[T any] func(ctx context.Context, msg messaging.Message[T]) time.Duration

// DropHandler is the optional observability hook invoked once per
// message dropped because the pending queue exceeded WithMaxPending.
// msg is type-erased; cast it inside the hook when payload-specific
// behavior is needed. err is always ErrDelayer(ErrMaxPendingExceeded).
// The hook is invoked synchronously from the source channel's
// dispatcher and must not block — long observability work should be
// dispatched asynchronously by the implementer.
type DropHandler func(ctx context.Context, msg any, err error)

// ErrDelayerFn is the function type for ErrDelayer.
type ErrDelayerFn func(causes ...error) error
