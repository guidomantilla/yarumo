// Package resequencer provides a Resequencer pattern over
// messaging.Channel[T].
//
// A Resequencer subscribes to a source Channel[T] that delivers
// out-of-order messages produced upstream (typically by a Splitter
// pattern) and forwards them to a destination Channel[T] in
// SequenceNumber order. Per-correlation buffering bridges the gap
// between async fan-out (where order is not guaranteed) and a
// downstream stage that needs the original order.
//
// Each input message must carry:
//
//   - Headers.CorrelationID — non-empty; groups messages of the same
//     logical sequence.
//   - Headers.SequenceNumber — 0-based position within the sequence.
//   - Headers.SequenceSize — total number of messages in the
//     sequence (set by the upstream Splitter). Must be positive.
//
// Messages without a valid correlation/size triplet are dropped via
// WithDropHandler — they are not part of any sequence and the
// resequencer would otherwise allocate an open-ended buffer for them.
//
// # Memory bounding (REQUIRED via WithGroupTimeout)
//
// Per-correlation buffering MUST be bounded. The constructor REQUIRES
// a positive WithGroupTimeout so groups whose missing position never
// arrives are eventually dropped by the sweeper goroutine — every
// still-buffered (unforwarded) message in the group fires the drop
// hook. WithMaxGroups caps the number of distinct correlations the
// Resequencer tracks at any one time; new correlations beyond the cap
// are dropped via WithDropHandler.
//
// # Emit semantics
//
// The Resequencer maintains a nextEmit cursor per correlation. On
// every arrival, it stores the message at msgs[seqNumber] and then
// drains as many consecutive positions as it can — emitting them to
// dst in order. When the cursor reaches expected size, the group is
// removed from the map (sequence complete).
//
// # Lifecycle
//
// Resequencer implements common/lifecycle.Component (worker-style):
// Start registers the subscription on the source channel and spawns
// the timeout sweeper goroutine; Stop cancels the subscription, stops
// the sweeper, drops every still-buffered message in incomplete
// groups via WithDropHandler, and closes Done.
//
// # Error handling
//
// The handler installed on the source channel always returns nil.
// Forward Send failures during emit flow through WithErrorHandler;
// intentional drops (missing/invalid sequence metadata, duplicate
// sequence number, MaxGroups cap, quorum timeout, Stop drain) flow
// through WithDropHandler. Nothing propagates to the source channel's
// Send caller error path, consistent with the package-wide policy in
// modules/messaging/CODING_STANDARDS.md.
package resequencer

import (
	"context"
	"time"

	"github.com/guidomantilla/yarumo/core/common/lifecycle"
)

// DefaultMaxGroups is the default cap on the number of distinct
// CorrelationIDs the Resequencer tracks at any one time. New
// correlations beyond the cap are dropped via WithDropHandler.
// Override with WithMaxGroups.
const DefaultMaxGroups = 1000

// DefaultSweepInterval is the default cadence at which the timeout
// sweeper inspects in-flight groups to evict ones that have exceeded
// WithGroupTimeout. The interval is chosen so the sweeper does not
// dominate CPU while still bounding the worst-case time-to-eviction.
const DefaultSweepInterval = 250 * time.Millisecond

var (
	_ Resequencer[any] = (*resequencer[any])(nil)

	_ ErrResequencerFn = ErrResequencer
)

// Resequencer is the public interface for a Resequencer endpoint. It
// embeds lifecycle.Component so callers wire it up with
// lifecycle.Build. The interface exists (rather than returning
// lifecycle.Component directly) so the consumer's API surface
// preserves "this is a Resequencer" semantics and the type stays open
// to future resequencer-specific methods without breaking callers.
type Resequencer[T any] interface {
	lifecycle.Component
}

// DropHandler is the optional observability hook invoked once per
// intentional drop:
//
//   - source message had empty CorrelationID or non-positive
//     SequenceSize (not part of any sequence);
//   - source message arrived for a new correlation while the
//     in-flight group count was already at WithMaxGroups;
//   - source message had SequenceNumber out of [0, SequenceSize) or
//     duplicated a position already buffered;
//   - a group's missing position never arrived within
//     WithGroupTimeout (fires once per still-buffered message);
//   - Stop was called with still-incomplete groups (fires once per
//     still-buffered message).
//
// msg is type-erased; cast it inside the hook when payload-specific
// behavior is needed. The hook is invoked from the source channel's
// dispatcher or the sweeper goroutine and must not block — long
// observability work should be dispatched asynchronously by the
// implementer.
type DropHandler func(ctx context.Context, msg any)

// ErrResequencerFn is the function type for ErrResequencer.
type ErrResequencerFn func(causes ...error) error
