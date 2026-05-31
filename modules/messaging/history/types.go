// Package history provides a Message History pattern over
// messaging.Channel[T].
//
// A History endpoint subscribes to a source Channel[T] and forwards each
// received Message[T] to a destination Channel[T] after appending its
// own configured name to a per-message slice stored under
// Headers.Custom[HistoryKey] (default "History"). The result is a
// chronological trail of endpoint names — useful for observability,
// audit and debugging when a message flows through a chain of EIP
// stages.
//
// History is pure header manipulation: no group state, no timeouts, no
// memory bounding. The pattern is intentionally trivial; its value is
// composition — drop one between any two channels and the downstream
// payload now carries provenance metadata.
//
// # Header semantics
//
// On each pass-through:
//
//   - If Headers.Custom is nil, a new map is allocated and seeded with
//     HistoryKey → []string{name}.
//   - If Headers.Custom[HistoryKey] is missing or carries a value of a
//     different shape (anything not []string), it is overwritten with a
//     fresh []string{name}. The pattern is forgiving — a misuse of the
//     same key by other code does not break the forward path.
//   - Otherwise the configured name is appended to the existing
//     []string and stored back.
//
// The forwarded message carries a COPY of Headers and Custom so
// downstream mutations do not leak back into the source.
//
// # Lifecycle
//
// History implements common/lifecycle.Component (worker-style): Start
// registers the subscription on the source channel and returns
// immediately; Stop cancels the subscription and closes Done. History
// does not spawn goroutines of its own — dispatch concurrency is
// inherited from the source channel implementation.
//
// # Error handling
//
// The handler installed on the source channel always returns nil.
// Forward Send failures are surfaced via the History's own ErrorHandler
// (installed with WithErrorHandler, defaulting to
// messaging.DefaultErrorHandler which logs via common/log). This keeps
// observability concerns out of the source channel's caller error path
// — consistent with the package-wide policy in
// modules/messaging/CODING_STANDARDS.md.
package history

import (
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
)

var (
	_ History[any] = (*history[any])(nil)

	_ ErrHistoryFn = ErrHistory
)

// DefaultHistoryKey is the default Headers.Custom map key used to
// store the history trail. Override with WithHistoryKey when the
// default would collide with a caller-defined custom field.
const DefaultHistoryKey = "History"

// History is the public interface for a Message History endpoint. It
// embeds lifecycle.Component so callers wire it up with
// lifecycle.Build. The interface exists (rather than returning
// lifecycle.Component directly) so the consumer's API surface
// preserves "this is a History" semantics and the type stays open to
// future history-specific methods without breaking callers.
type History[T any] interface {
	lifecycle.Component
}

// ErrHistoryFn is the function type for ErrHistory.
type ErrHistoryFn func(causes ...error) error
