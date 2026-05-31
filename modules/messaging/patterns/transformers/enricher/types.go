// Package enricher provides a Content/Header Enricher pattern over
// messaging.Channel[T].
//
// An Enricher subscribes to a source Channel[T] and forwards every
// received Message[T] to a single destination Channel[T] after applying
// a user-supplied EnrichFn that returns a new Message[T] with added or
// overridden Headers fields, Payload mutations, or both.
//
// # Header AND content enricher in one pattern
//
// The EIP catalog lists "Header Enricher" and "Content Enricher" as
// separate patterns. This package intentionally consolidates them into
// a single Enricher pattern: the EnrichFn callback receives the full
// Message[T] and returns the modified Message[T], so the caller decides
// whether to enrich the Headers, the Payload, or both in one place.
// Splitting them would multiply the wiring boilerplate for no semantic
// gain (both are "map(Message) → Message" steps subscribed to a source
// channel).
//
// Typical uses:
//
//   - Populate Headers.CorrelationID / CausationID before forwarding to
//     downstream sagas.
//   - Add an audit Source tag identifying the enriching gateway.
//   - Decorate the Payload with denormalised lookup fields fetched from
//     a side store (the EnrichFn may perform IO; ctx is propagated for
//     deadline and cancellation).
//
// # Immutability
//
// EnrichFn returns a NEW Message[T] value rather than mutating the
// input in place; the source Message[T] is forwarded unchanged when
// EnrichFn errors or panics (no partial enrichment leaks). Callers
// implementing EnrichFn are responsible for cloning maps/slices inside
// the input Message before mutating them — the Headers struct itself is
// a value, but Headers.Custom is a map reference shared with the
// publisher.
//
// # Lifecycle
//
// Enricher implements common/lifecycle.Component (worker-style): Start
// registers the subscription on the source channel and returns
// immediately; Stop cancels the subscription and closes Done. The
// Enricher does not spawn goroutines of its own — dispatch concurrency
// is inherited from the source channel implementation.
//
// # Error handling
//
// The handler installed on the source channel always returns nil.
// EnrichFn failures (returned error, panic) and forward Send failures
// are surfaced via the Enricher's own ErrorHandler (installed with
// WithErrorHandler, defaulting to messaging.DefaultErrorHandler which
// logs via common/log). When EnrichFn errors or panics the original
// message is NOT forwarded — enrichment is treated as all-or-nothing.
package enricher

import (
	"context"

	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/messaging"
)

var (
	_ Enricher[any] = (*enricher[any])(nil)

	_ ErrEnricherFn = ErrEnricher
)

// Enricher is the public interface for a Content/Header Enricher. It
// embeds lifecycle.Component so callers wire it up with
// lifecycle.Build. The interface exists (rather than returning
// lifecycle.Component directly) so the consumer's API surface
// preserves "this is an Enricher" semantics and the type stays open to
// future enricher-specific methods without breaking callers.
type Enricher[T any] interface {
	lifecycle.Component
}

// EnrichFn maps an input Message[T] to an enriched Message[T]. The
// caller may add / override Headers fields, mutate Payload (returning
// a new Message with the updated payload), or both. An error returned
// by EnrichFn is wrapped in ErrEnricher(ErrEnrichFnFailed, err) and
// forwarded to the ErrorHandler; the original message is not
// forwarded. A panic in EnrichFn is recovered and wrapped in
// ErrEnricher(ErrEnrichPanic, ...) so it cannot kill the source
// channel's dispatcher.
type EnrichFn[T any] func(ctx context.Context, msg messaging.Message[T]) (messaging.Message[T], error)

// ErrEnricherFn is the function type for ErrEnricher.
type ErrEnricherFn func(causes ...error) error
