// Package idempotent provides an Idempotent Receiver pattern over
// messaging.Channel[T].
//
// An Idempotent subscribes to a source Channel[T] and forwards each
// received Message[T] to a single destination Channel[T] only the FIRST
// time a given dedup key is observed within the configured TTL. Repeat
// messages with the same key are dropped — the receiver is then safe to
// re-deliver from upstream (broker redelivery, outbox replay, at-least-
// once transports) without side effects.
//
// # Dedup state is external
//
// The dedup history is not kept in the idempotent struct itself; it is
// kept in a `store.MetadataStore` instance the caller injects. The
// store decides:
//
//   - Where the history lives (in-process map, Redis, Postgres, …).
//   - When entries expire (per-Add TTL).
//   - Concurrency semantics across replicas (Redis SETEX is the
//     canonical multi-replica pattern).
//
// Picking the right store is how you tune the dedup window and the
// blast radius of duplicates.
//
// # Key extraction
//
// The dedup key is extracted from each Message[T] via a KeyFn[T]. The
// default extractor pulls Headers.MessageID — the canonical envelope
// identifier per the messaging package — and treats an empty
// MessageID as "no key, cannot dedup, drop the message" (routed to
// WithDropHandler with the reason "no key"). Wire WithKeyFn when a
// different field carries the dedup identity (CorrelationID for saga-
// idempotency, a payload field via a custom extractor, …).
//
// # Two observability hooks (shape B, mirrors filter/)
//
// Idempotent distinguishes two outcomes that look identical from the
// caller's perspective but mean very different things operationally:
//
//   - **Intentional drops** (duplicate observed, or empty key): the
//     idempotent receiver did its job. These are routed to the optional
//     DropHandler hook (installed via WithDropHandler). Default is nil
//     — silent drop — because a noisy dedup receiver that logs every
//     duplicate is rarely useful. Wire WithDropHandler when you need
//     throughput metrics ("how many passed / how many were deduped")
//     or audit trails.
//   - **Real failures** (store check failed, store add failed, forward
//     Send failed): the receiver could not do its job. These are
//     routed to the ErrorHandler (installed via WithErrorHandler,
//     defaulting to messaging.DefaultErrorHandler which logs via
//     common/log). The default ensures failures never disappear
//     silently.
//
// # Fail-open on Add, fail-closed on Has
//
// When the underlying MetadataStore errors, the receiver picks
// different policies per call:
//
//   - `Has` error → fail-CLOSED: the message is NOT forwarded, the
//     error is routed through WithErrorHandler. Forwarding under an
//     unknown dedup state is worse than dropping a single message — a
//     duplicate then has consequences.
//   - `Add` error → fail-OPEN: the message IS forwarded, the error is
//     routed through WithErrorHandler. The check already returned false
//     (genuinely new message); failing to record the receipt makes a
//     future duplicate possible, but dropping a known-new message
//     guarantees lost work today. Prefer a known duplicate over a
//     known drop.
//
// # Lifecycle
//
// Idempotent implements common/lifecycle.Component (worker-style):
// Start registers the subscription on the source channel and returns
// immediately; Stop cancels the subscription and closes Done.
// Idempotent does not spawn goroutines of its own — dispatch
// concurrency is inherited from the source channel implementation, and
// the dedup store owns its own goroutines (if any).
//
// # Error handling
//
// The handler installed on the source channel always returns nil.
// Idempotent concerns (store error, forward Send failure) flow through
// WithErrorHandler; intentional drops flow through WithDropHandler.
// Nothing propagates to the source channel's Send caller error path,
// consistent with the package-wide policy in
// modules/messaging/CODING_STANDARDS.md.
package idempotent

import (
	"context"

	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/messaging"
)

var (
	_ Idempotent[any] = (*idempotent[any])(nil)

	_ ErrIdempotentFn = ErrIdempotent
)

// Idempotent is the public interface for an Idempotent Receiver. It
// embeds lifecycle.Component so callers wire it up with
// lifecycle.Build. The interface exists (rather than returning
// lifecycle.Component directly) so the consumer's API surface
// preserves "this is an Idempotent" semantics and the type stays open
// to future idempotent-specific methods without breaking callers.
type Idempotent[T any] interface {
	lifecycle.Component
}

// KeyFn extracts the dedup key from a Message[T]. Implementations must
// be deterministic — the same logical message must yield the same
// key. An empty string return means "no dedup key available, drop the
// message" (routed to the DropHandler with the reason "no key").
//
// Errors are not modeled on KeyFn intentionally: key extraction is
// expected to be a pure read of Headers/Payload fields with no I/O.
// If a future extractor needs I/O, escalate the dedup decision to a
// pre-stage filter or a custom message-store backend.
type KeyFn[T any] func(msg messaging.Message[T]) string

// DropReason classifies why a message was intentionally dropped by an
// Idempotent receiver. It is the second argument to DropHandler so
// observers can split metrics between "duplicate" (dedup did its job)
// and "no key" (message could not be deduped at all and was rejected
// rather than forwarded under unknown identity).
type DropReason string

const (
	// DropReasonDuplicate indicates the dedup key was already recorded
	// in the MetadataStore — the message is a duplicate and was not
	// forwarded.
	DropReasonDuplicate DropReason = "duplicate"
	// DropReasonNoKey indicates KeyFn returned an empty string for the
	// message — the receiver has no dedup identity to record and chose
	// not to forward.
	DropReasonNoKey DropReason = "no-key"
)

// DropHandler is the optional observability hook invoked once per
// intentional drop (duplicate or no-key). msg is type-erased; cast it
// inside the hook when payload-specific behavior is needed. reason
// classifies the drop. The hook is invoked synchronously from the
// source channel's dispatcher and must not block — long observability
// work should be dispatched asynchronously by the implementer.
//
// DropHandler is NOT invoked when the store errors or the forward
// fails (those are routed to the ErrorHandler instead) — DropHandler
// fires only on successful, deliberate drops.
type DropHandler func(ctx context.Context, msg any, reason DropReason)

// ErrIdempotentFn is the function type for ErrIdempotent.
type ErrIdempotentFn func(causes ...error) error
