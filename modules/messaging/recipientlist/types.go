// Package recipientlist provides a Recipient List pattern over
// messaging.Channel[T].
//
// A RecipientList subscribes to a source Channel[T] and forwards each
// received Message[T] to N destinations chosen by a user-supplied
// SelectorFn. Unlike a Content-Based Router (1→1 by key), the Recipient
// List is 1→N: SelectorFn returns a slice of keys and the message is
// sent to every channel resolved from routes[key]. The pattern decouples
// publishers from the dynamic set of interested consumers; routing rules
// live declaratively in a (key → channel) map.
//
// # Lifecycle
//
// RecipientList implements common/lifecycle.Component (worker-style):
// Start registers the subscription on the source channel and returns
// immediately; Stop cancels the subscription and closes Done. The
// RecipientList does not spawn goroutines of its own — dispatch
// concurrency is inherited from the source channel implementation. Wire
// it via lifecycle.Build for the standard daemon CloseFn pattern.
//
// # Error handling
//
// The handler installed on the source channel always returns nil.
// Routing failures (SelectorFn error or panic, missing key for a single
// recipient, per-recipient forward Send failure) are surfaced via the
// RecipientList's own ErrorHandler (installed with WithErrorHandler,
// defaulting to messaging.DefaultErrorHandler which logs via
// common/log). Per-recipient errors are reported individually so a
// single missing key or single forward failure does NOT abort delivery
// to the other recipients — partial success is the design.
//
// # Empty selection policy
//
// When SelectorFn returns an empty slice (no recipients matched), the
// message is intentionally dropped and forwarded to the configured
// DropHandler (nil by default — silent). This is the analogue of
// filter's "predicate returned false": a deliberate decision, not an
// error.
package recipientlist

import (
	"context"

	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/messaging"
)

var (
	_ RecipientList[any] = (*recipientList[any])(nil)

	_ ErrRecipientListFn = ErrRecipientList
)

// RecipientList is the public interface for a Recipient List. It embeds
// lifecycle.Component so callers wire it up with lifecycle.Build. The
// interface exists (rather than returning lifecycle.Component directly)
// so the consumer's API surface preserves "this is a RecipientList"
// semantics and the type stays open to future recipient-list-specific
// methods without breaking callers.
type RecipientList[T any] interface {
	lifecycle.Component
}

// SelectorFn returns the ordered list of destination keys for msg. Each
// key is looked up independently in the routes map provided to
// NewRecipientList; missing keys trigger per-recipient ErrNoRoute
// reports through the ErrorHandler without aborting the other
// deliveries. An empty slice means "no recipients" and routes the
// message to the DropHandler. An error returned by SelectorFn is wrapped
// in ErrRecipientList(ErrSelectorFnFailed, err) and forwarded to the
// ErrorHandler. A panic in SelectorFn is recovered and wrapped in
// ErrRecipientList(ErrSelectorPanic, ...) so it cannot kill the source
// channel's dispatcher.
type SelectorFn[T any] func(ctx context.Context, msg messaging.Message[T]) ([]string, error)

// DropHandler is the optional observability hook invoked once per
// intentional drop (SelectorFn returned an empty slice). msg is
// type-erased; cast it inside the hook when payload-specific behavior
// is needed. The hook is invoked synchronously from the source
// channel's dispatcher and must not block — long observability work
// should be dispatched asynchronously by the implementer.
//
// DropHandler is NOT invoked when SelectorFn errors or panics (those
// are routed to the ErrorHandler instead) — DropHandler fires only on
// successful, deliberate "no recipients" decisions.
type DropHandler func(ctx context.Context, msg any)

// ErrRecipientListFn is the function type for ErrRecipientList.
type ErrRecipientListFn func(causes ...error) error
