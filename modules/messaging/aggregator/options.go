package aggregator

import (
	"time"

	"github.com/guidomantilla/yarumo/messaging"
)

// defaultMaxGroups bounds the in-flight group map when the caller does
// not pass WithMaxGroups. Sized to be comfortably above realistic batch
// fan-in workloads (scatter/gather of dozens, request/reply batches of
// hundreds) while still catching runaway producers within a reasonable
// memory budget.
const defaultMaxGroups = 1000

// Option is a functional option for configuring aggregator Options. It
// is generic over T so options that carry T-typed values (CorrelationFn,
// CompletionFn) stay type-safe. U is not threaded through Option because
// none of the options carry U-typed values; AggregateFn[T, U] is a
// mandatory positional argument to NewAggregator rather than an option.
type Option[T any] func(opts *Options[T])

// Options holds the configuration for an Aggregator.
type Options[T any] struct {
	correlation    CorrelationFn[T]
	completion     CompletionFn[T]
	completionSize int
	groupTimeout   time.Duration
	maxGroups      int
	errorHandler   messaging.ErrorHandler
	dropHandler    DropHandler
}

// NewOptions creates a new Options[T] with sensible defaults and
// applies the given options. The default CorrelationFn reads
// msg.Headers.CorrelationID; the default ErrorHandler is
// messaging.DefaultErrorHandler; the default MaxGroups cap is
// defaultMaxGroups; the default DropHandler is nil (silent intentional
// drops); no completion strategy is configured by default — at least
// one of WithCompletionFn, WithCompletionSize or WithGroupTimeout MUST
// be passed by the caller.
func NewOptions[T any](opts ...Option[T]) *Options[T] {
	options := &Options[T]{
		correlation:  defaultCorrelation[T],
		maxGroups:    defaultMaxGroups,
		errorHandler: messaging.DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// defaultCorrelation extracts the correlation key from the standard
// Headers.CorrelationID field. Empty string disables aggregation for
// the message; see CorrelationFn doc and WithDropHandler.
func defaultCorrelation[T any](msg messaging.Message[T]) string {
	return msg.Headers.CorrelationID
}

// WithCorrelationFn overrides the default correlation extractor. The
// default reads msg.Headers.CorrelationID. Use a custom CorrelationFn
// when correlation lives in the payload, in a custom Headers.Custom
// entry, or is derived (e.g. tenant + day). Nil values are ignored
// (the previously installed extractor is preserved).
func WithCorrelationFn[T any](fn CorrelationFn[T]) Option[T] {
	return func(opts *Options[T]) {
		if fn != nil {
			opts.correlation = fn
		}
	}
}

// WithCompletionFn installs a predicate-based completion strategy.
// fn(group) is evaluated after every message added to a group; true
// releases the group. Combine with WithCompletionSize and/or
// WithGroupTimeout — the first enabled strategy that fires wins. Nil
// values are ignored.
func WithCompletionFn[T any](fn CompletionFn[T]) Option[T] {
	return func(opts *Options[T]) {
		if fn != nil {
			opts.completion = fn
		}
	}
}

// WithCompletionSize installs a size-based completion strategy. A group
// is released the moment its message count reaches n. Non-positive
// values are ignored.
func WithCompletionSize[T any](n int) Option[T] {
	return func(opts *Options[T]) {
		if n > 0 {
			opts.completionSize = n
		}
	}
}

// WithGroupTimeout installs a timeout-based completion strategy. A
// background sweeper goroutine (spawned in Start) releases any group
// that has sat idle for d since its last message arrival. Required to
// bound memory when partial groups may never reach size or predicate
// completion. Non-positive values are ignored.
func WithGroupTimeout[T any](d time.Duration) Option[T] {
	return func(opts *Options[T]) {
		if d > 0 {
			opts.groupTimeout = d
		}
	}
}

// WithMaxGroups caps the number of concurrently tracked groups. A
// message that would create the n+1-th group is dropped and the
// ErrorHandler is invoked with ErrMaxGroupsExceeded. The default is
// defaultMaxGroups (1000). Non-positive values are ignored.
func WithMaxGroups[T any](n int) Option[T] {
	return func(opts *Options[T]) {
		if n > 0 {
			opts.maxGroups = n
		}
	}
}

// WithErrorHandler installs an observability hook fired once per real
// aggregator failure (AggregateFn returned error or panicked, forward
// Send failed, MaxGroups exceeded). The default (when WithErrorHandler
// is not passed) is messaging.DefaultErrorHandler, which logs each
// failure via common/log so consumers that forget to wire observability
// still see failures. Pass messaging.SilentErrorHandler to opt out, or
// any custom hook to redirect. Nil values are ignored.
func WithErrorHandler[T any](handler messaging.ErrorHandler) Option[T] {
	return func(opts *Options[T]) {
		if handler != nil {
			opts.errorHandler = handler
		}
	}
}

// WithDropHandler installs an observability hook fired once per
// intentional drop (empty correlation key, group expired by sweeper
// without messages to aggregate, partial groups released during Stop).
// The default (when WithDropHandler is not passed) is nil — intentional
// drops are silent. Wire this for throughput metrics or audit trails.
// Nil values are ignored.
func WithDropHandler[T any](handler DropHandler) Option[T] {
	return func(opts *Options[T]) {
		if handler != nil {
			opts.dropHandler = handler
		}
	}
}
