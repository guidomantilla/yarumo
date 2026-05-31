// Package headerfilter provides a Header Filter pattern over
// messaging.Channel[T].
//
// A HeaderFilter subscribes to a source Channel[T] and forwards every
// received Message[T] to a single destination Channel[T] with
// configured Headers fields cleared (zeroed for known struct fields,
// removed from the Custom map for arbitrary keys). The pattern is the
// metadata analogue of a Content Filter: the payload is untouched, but
// metadata fields the consumer should not see (or which leaked from
// upstream) are removed before forwarding.
//
// Typical uses:
//
//   - Strip a sensitive ReplyTo / Source from a public outbound topic.
//   - Drop a redundant CorrelationID before re-injecting into a
//     downstream system that re-generates its own.
//   - Remove debug-only Custom entries before crossing an internal/
//     external boundary.
//
// # Known fields vs Custom map
//
// WithClearHeader accepts both struct field names (zeroed in place) and
// arbitrary names (deleted from the Custom map). Recognised struct
// field names: "MessageID", "CorrelationID", "CausationID", "ReplyTo",
// "Type", "Source", "ContentType", "Priority", "ExpirationTime",
// "SequenceNumber", "SequenceSize", "Timestamp". Any other name is
// treated as a Custom key and removed via delete(Headers.Custom, name).
//
// # Immutability
//
// The forwarded Message is a shallow copy of the original with a
// rebuilt Headers value; the source Message[T] is never mutated. When
// only Custom keys are removed the original Custom map is also left
// intact (a defensive copy is made before deletion).
//
// # Lifecycle
//
// HeaderFilter implements common/lifecycle.Component (worker-style):
// Start registers the subscription on the source channel and returns
// immediately; Stop cancels the subscription and closes Done. The
// HeaderFilter does not spawn goroutines of its own — dispatch
// concurrency is inherited from the source channel implementation.
//
// # Error handling
//
// The handler installed on the source channel always returns nil.
// Forward Send failures are surfaced via the HeaderFilter's own
// ErrorHandler (installed with WithErrorHandler, defaulting to
// messaging.DefaultErrorHandler which logs via common/log).
package headerfilter

import (
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
)

var (
	_ HeaderFilter[any] = (*headerFilter[any])(nil)

	_ ErrHeaderFilterFn = ErrHeaderFilter
)

// HeaderFilter is the public interface for a Header Filter. It embeds
// lifecycle.Component so callers wire it up with lifecycle.Build. The
// interface exists (rather than returning lifecycle.Component directly)
// so the consumer's API surface preserves "this is a HeaderFilter"
// semantics and the type stays open to future header-filter-specific
// methods without breaking callers.
type HeaderFilter[T any] interface {
	lifecycle.Component
}

// ErrHeaderFilterFn is the function type for ErrHeaderFilter.
type ErrHeaderFilterFn func(causes ...error) error
