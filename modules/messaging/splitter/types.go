// Package splitter provides a Splitter pattern over
// messaging.Channel.
//
// A Splitter subscribes to a source Channel[T] and, for each received
// Message[T], invokes a user-supplied SplitFn to obtain a slice of U
// values. It then emits one Message[U] per slice item on the
// destination Channel[U], populating Headers.SequenceNumber (0-based)
// and Headers.SequenceSize so that a downstream Aggregator pattern can
// reconstruct the original message.
//
// Header lineage of the split children:
//
//   - **CorrelationID**: preserved from the source message. This is
//     the canonical Aggregator key — all children of one source
//     message share one CorrelationID.
//   - **MessageID**: re-stamped per child as
//     `<original-id>-<index>` so each child carries a unique ID
//     suitable for dedup downstream while still being traceable back
//     to the source. When the source has no MessageID, the child IDs
//     are just `<index>` (rare — most channels stamp MessageID at
//     publish time).
//   - **CausationID**: set to the source MessageID (the source caused
//     the splits).
//   - **SequenceNumber**: 0-based index into the slice.
//   - **SequenceSize**: total count = len(slice).
//   - All other Headers fields (Type, Priority, ContentType,
//     Source, Custom, ...) are preserved from the source message
//     verbatim. Callers that need per-child transformations should
//     chain a Transformer downstream rather than overloading the
//     SplitFn return slice with custom envelopes.
//
// Edge cases:
//
//   - **Empty slice**: SplitFn returned a 0-length slice. This is
//     treated as an "intentional drop" — analogous to filter's drop
//     semantics — and routes to the optional DropHandler (silent by
//     default). No message is forwarded to dst.
//   - **Single-item slice**: forwarded as a single Message[U] with
//     SequenceNumber=0, SequenceSize=1. The receiving Aggregator
//     therefore handles "splits of one" uniformly with multi-item
//     splits.
//
// # Lifecycle
//
// Splitter implements common/lifecycle.Component (worker-style): Start
// registers the subscription on the source channel and returns
// immediately; Stop cancels the subscription and closes Done. The
// splitter does not spawn goroutines of its own — dispatch concurrency
// is inherited from the source channel implementation.
//
// # Error handling
//
// The handler installed on the source channel always returns nil.
// Splitter failures (SplitFn error, SplitFn panic, forward Send
// failure on any child) are surfaced via the Splitter's own
// ErrorHandler. When a child Send fails mid-batch, the splitter
// reports the failure and continues to emit the remaining children —
// downstream consumers that need atomic batches must wrap the
// destination in a transactional Channel impl.
package splitter

import (
	"context"

	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/messaging"
)

var (
	_ Splitter[any, any] = (*splitter[any, any])(nil)

	_ ErrSplitterFn = ErrSplitter
)

// Splitter is the public interface for a Splitter. It embeds
// lifecycle.Component so callers wire it up with lifecycle.Build. The
// interface exists (rather than returning lifecycle.Component
// directly) so the consumer's API surface preserves "this is a
// Splitter" semantics and the type stays open to future
// splitter-specific methods without breaking callers.
type Splitter[T, U any] interface {
	lifecycle.Component
}

// SplitFn returns the slice of U values to emit for the incoming
// Message[T]. Returning a nil or empty slice is treated as an
// intentional drop and routed to the DropHandler (silent by default).
// An error returned by SplitFn is wrapped in
// ErrSplitter(ErrSplitFailed, err) and forwarded to the ErrorHandler;
// the message is dropped. A panic in SplitFn is recovered and wrapped
// in ErrSplitter(ErrSplitterPanic, ...) so it cannot kill the source
// channel's dispatcher.
type SplitFn[T, U any] func(ctx context.Context, msg messaging.Message[T]) ([]U, error)

// DropHandler is the optional observability hook invoked once per
// intentional drop (SplitFn returned an empty slice). msg is
// type-erased; cast it inside the hook when payload-specific behavior
// is needed. The hook is invoked synchronously from the source
// channel's dispatcher and must not block — long observability work
// should be dispatched asynchronously by the implementer.
//
// DropHandler is NOT invoked when the SplitFn errors or panics (those
// are routed to the ErrorHandler instead) — DropHandler fires only on
// deliberate empty-slice drops.
type DropHandler func(ctx context.Context, msg any)

// ErrSplitterFn is the function type for ErrSplitter.
type ErrSplitterFn func(causes ...error) error
