package idempotent

import (
	"time"

	"github.com/guidomantilla/yarumo/messaging"
)

// defaultTTL is the dedup window applied when no WithTTL option is
// supplied. 24 hours is the conventional ceiling for at-least-once
// broker redelivery — adjust per workload.
const defaultTTL = 24 * time.Hour

// Option is a functional option for configuring idempotent Options.
// Option is generic in T because WithKeyFn carries a T-typed extractor;
// keeping the Option type-parameterized lets WithKeyFn match the
// constructor's T without an unsafe-any cast.
type Option[T any] func(opts *Options[T])

// Options holds the configuration for an Idempotent receiver.
type Options[T any] struct {
	ttl          time.Duration
	keyFn        KeyFn[T]
	errorHandler messaging.ErrorHandler
	dropHandler  DropHandler
}

// NewOptions creates a new Options with sensible defaults and applies
// the given options. Defaults:
//
//   - TTL: 24h (conventional broker-redelivery ceiling).
//   - KeyFn: DefaultKeyFn (extracts Headers.MessageID).
//   - ErrorHandler: messaging.DefaultErrorHandler (logs via common/log).
//   - DropHandler: nil (intentional drops are silent unless wired).
func NewOptions[T any](opts ...Option[T]) *Options[T] {
	options := &Options[T]{
		ttl:          defaultTTL,
		keyFn:        DefaultKeyFn[T],
		errorHandler: messaging.DefaultErrorHandler,
		dropHandler:  nil,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithTTL configures the dedup window recorded via MetadataStore.Add
// for each newly observed key. Non-positive values are ignored (the
// previously installed, or default, TTL is preserved). The TTL is
// passed through to the underlying MetadataStore on every Add — the
// store decides how to honor it (in-memory map with sweeper, Redis
// SETEX, …).
func WithTTL[T any](ttl time.Duration) Option[T] {
	return func(opts *Options[T]) {
		if ttl > 0 {
			opts.ttl = ttl
		}
	}
}

// WithKeyFn installs the dedup key extractor used per Message[T]. The
// default (DefaultKeyFn) returns Headers.MessageID. Wire this when a
// different field carries the dedup identity (CorrelationID for saga-
// idempotency, a payload field via a closure, …). Nil values are
// ignored (the previously installed extractor is preserved).
func WithKeyFn[T any](fn KeyFn[T]) Option[T] {
	return func(opts *Options[T]) {
		if fn != nil {
			opts.keyFn = fn
		}
	}
}

// WithErrorHandler installs an observability hook fired once per real
// idempotent failure (store Has failed, store Add failed, forward
// Send failed). The default (when WithErrorHandler is not passed) is
// messaging.DefaultErrorHandler, which logs each failure via
// common/log so consumers that forget to wire observability still see
// real failures. Pass messaging.SilentErrorHandler to opt out, or any
// custom hook to redirect. Nil values are ignored (the previously
// installed handler is preserved).
func WithErrorHandler[T any](handler messaging.ErrorHandler) Option[T] {
	return func(opts *Options[T]) {
		if handler != nil {
			opts.errorHandler = handler
		}
	}
}

// WithDropHandler installs an observability hook fired once per
// intentional drop (duplicate observed or no key available). The
// default (when WithDropHandler is not passed) is nil — intentional
// drops are silent. Wire this to ship throughput metrics ("forwarded
// vs deduped") or audit trails. Nil arguments are ignored (the
// previously installed handler is preserved); to explicitly disable
// observability after a non-nil hook was installed, install a no-op
// `func(_ context.Context, _ any, _ DropReason) {}`.
func WithDropHandler[T any](handler DropHandler) Option[T] {
	return func(opts *Options[T]) {
		if handler != nil {
			opts.dropHandler = handler
		}
	}
}

// DefaultKeyFn extracts Headers.MessageID as the dedup key. This is
// the canonical envelope identifier per the messaging package and the
// natural identity for at-least-once redelivery semantics. Empty
// MessageID is returned as an empty string — the Idempotent receiver
// interprets that as "no dedup key, drop" and routes the message
// through DropHandler with DropReasonNoKey.
func DefaultKeyFn[T any](msg messaging.Message[T]) string {
	return msg.Headers.MessageID
}
