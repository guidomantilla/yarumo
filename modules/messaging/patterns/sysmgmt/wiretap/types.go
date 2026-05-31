// Package wiretap provides a Wire Tap pattern over messaging.Channel[T].
//
// A Wiretap subscribes to a source Channel[T] and forwards each
// received Message[T] to BOTH a primary destination channel and a
// side-channel "tap" used for observability — logging, metrics, audit
// trails, or debug capture — without altering the primary flow.
//
// The tap is intentionally treated as second-class: failures to publish
// to the tap are silenced from the primary flow (the tap MUST NEVER
// affect the production path). Tap failures are still observable
// through the optional WithErrorHandler hook so that operators can
// monitor "the observability of the observability".
//
// Typical use cases:
//
//   - **Audit logging**: tap every order event into a write-only audit
//     channel while the primary flow continues to billing.
//   - **Metrics scraping**: tap a side channel that batches counts and
//     latencies without coupling the primary handler to a metrics SDK.
//   - **Debug capture**: in dev/staging, tap a circular buffer
//     subscriber to inspect the last N messages without modifying the
//     primary subscribers.
//
// # Lifecycle
//
// Wiretap implements common/lifecycle.Component (worker-style): Start
// registers the subscription on the source channel and returns
// immediately; Stop cancels the subscription and closes Done. The
// wiretap does not spawn goroutines of its own — dispatch concurrency
// is inherited from the source channel implementation.
//
// # Error handling
//
// The handler installed on the source channel always returns nil.
// Primary dst Send failures are surfaced via the Wiretap's own
// ErrorHandler with ErrWiretap(ErrForwardFailed, err); tap Send
// failures are also surfaced (separately, with ErrTapSendFailed) so
// they never go truly silent — but they never alter the primary flow.
// Consistent with the package-wide policy in
// modules/messaging/CODING_STANDARDS.md, nothing propagates to the
// source channel's Send caller.
//
// The order of operations is documented: the primary dst Send happens
// first; the tap Send happens second regardless of whether the primary
// succeeded. This guarantees that a tap subscriber sees the same
// messages the primary path attempts to forward, even when the primary
// path fails (useful for "what did we try to do" audit).
package wiretap

import (
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
)

var (
	_ Wiretap[any] = (*wiretap[any])(nil)

	_ ErrWiretapFn = ErrWiretap
)

// Wiretap is the public interface for a Wire Tap. It embeds
// lifecycle.Component so callers wire it up with lifecycle.Build. The
// interface exists (rather than returning lifecycle.Component
// directly) so the consumer's API surface preserves "this is a
// Wiretap" semantics and the type stays open to future
// wiretap-specific methods without breaking callers.
type Wiretap[T any] interface {
	lifecycle.Component
}

// ErrWiretapFn is the function type for ErrWiretap.
type ErrWiretapFn func(causes ...error) error
