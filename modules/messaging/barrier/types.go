// Package barrier provides a Barrier pattern over
// messaging.Channel[T].
//
// A Barrier subscribes to a source Channel[T] and accumulates messages
// per Headers.CorrelationID. When the count of accumulated messages
// for a given correlation reaches the configured quorum, the Barrier
// releases ALL accumulated messages (the originals, unmodified) to
// the destination channel in arrival order, then forgets the group.
//
// # Barrier vs Aggregator
//
// Both patterns buffer messages by correlation and act on completion,
// but they differ in what they emit:
//
//   - Aggregator combines the buffered messages into one new payload
//     (sum, list, business object) and emits a SINGLE message.
//   - Barrier emits the buffered messages AS-IS — N messages in, N
//     messages out — preserving the original envelopes (payload,
//     headers, sequence info).
//
// Use Barrier when downstream stages need the original messages but
// must not start processing them until a quorum has arrived (for
// example: wait for all "saga participant ready" events before
// fanning out the saga payload).
//
// # Memory bounding (REQUIRED via WithGroupTimeout)
//
// In-memory accumulation per correlation MUST be bounded. The
// constructor REQUIRES a positive WithGroupTimeout so groups that
// never reach quorum are eventually dropped by the sweeper goroutine
// (the timeout sweeper drops every accumulated message in the group
// via WithDropHandler). The optional WithMaxGroups caps the number of
// distinct correlations the Barrier tracks at any one time; new
// correlations beyond the cap are dropped with WithDropHandler.
//
// # Lifecycle
//
// Barrier implements common/lifecycle.Component (worker-style): Start
// registers the subscription on the source channel and spawns the
// timeout sweeper goroutine; Stop cancels the subscription, stops the
// sweeper, drops every still-incomplete group through WithDropHandler,
// and closes Done. Wire it via lifecycle.Build for the standard
// daemon CloseFn pattern.
//
// # Error handling
//
// The handler installed on the source channel always returns nil.
// Forward Send failures (during release) flow through WithErrorHandler;
// intentional drops (missing correlation, quorum timeout, MaxGroups
// cap, Stop drain) flow through WithDropHandler. Nothing propagates
// to the source channel's Send caller error path, consistent with the
// package-wide policy in modules/messaging/CODING_STANDARDS.md.
package barrier

import (
	"context"
	"time"

	"github.com/guidomantilla/yarumo/core/common/lifecycle"
)

// DefaultMaxGroups is the default cap on the number of distinct
// CorrelationIDs the Barrier tracks at any one time. New correlations
// beyond the cap are dropped via WithDropHandler. Override with
// WithMaxGroups.
const DefaultMaxGroups = 1000

// DefaultSweepInterval is the default cadence at which the timeout
// sweeper inspects in-flight groups to evict ones that have exceeded
// WithGroupTimeout. The interval is chosen so the sweeper does not
// dominate CPU while still bounding the worst-case time-to-eviction.
const DefaultSweepInterval = 250 * time.Millisecond

var (
	_ Barrier[any] = (*barrier[any])(nil)

	_ ErrBarrierFn = ErrBarrier
)

// Barrier is the public interface for a Barrier endpoint. It embeds
// lifecycle.Component so callers wire it up with lifecycle.Build. The
// interface exists (rather than returning lifecycle.Component
// directly) so the consumer's API surface preserves "this is a
// Barrier" semantics and the type stays open to future barrier-
// specific methods without breaking callers.
type Barrier[T any] interface {
	lifecycle.Component
}

// DropHandler is the optional observability hook invoked once per
// intentional drop:
//
//   - source message had empty CorrelationID (cannot be grouped);
//   - source message arrived for a correlation already at the
//     WithMaxGroups cap;
//   - a group failed to reach quorum within WithGroupTimeout (fires
//     once per accumulated message in that group);
//   - Stop was called with still-incomplete groups (fires once per
//     accumulated message).
//
// msg is type-erased; cast it inside the hook when payload-specific
// behavior is needed. The hook is invoked from the source channel's
// dispatcher or the sweeper goroutine and must not block — long
// observability work should be dispatched asynchronously by the
// implementer.
type DropHandler func(ctx context.Context, msg any)

// ErrBarrierFn is the function type for ErrBarrier.
type ErrBarrierFn func(causes ...error) error
