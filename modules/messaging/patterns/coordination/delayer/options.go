package delayer

import (
	"time"

	"github.com/guidomantilla/yarumo/messaging"
)

// defaultMaxPending bounds the number of messages in flight (scheduled
// but not yet delivered). New messages over the bound are dropped via
// WithDropHandler.
const defaultMaxPending = 1024

// Option is a functional option for configuring delayer Options[T]. It
// is generic over T so options can carry T-typed values (e.g. DelayFn[T])
// without losing type safety.
type Option[T any] func(opts *Options[T])

// Options holds the configuration for a Delayer[T].
type Options[T any] struct {
	fixedDelay   time.Duration
	delayFn      DelayFn[T]
	maxPending   int
	errorHandler messaging.ErrorHandler
	dropHandler  DropHandler
}

// NewOptions creates a new Options[T] with sensible defaults and
// applies the given options. The default ErrorHandler is
// messaging.DefaultErrorHandler (logs via common/log); the default
// DropHandler is nil (drops are silent unless wired); maxPending is
// defaultMaxPending; no fixedDelay and no DelayFn are set so the
// fallback path uses Headers.ExpirationTime.
func NewOptions[T any](opts ...Option[T]) *Options[T] {
	options := &Options[T]{
		maxPending:   defaultMaxPending,
		errorHandler: messaging.DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithFixedDelay configures the delayer to wait this constant duration
// per message before forwarding it. Takes precedence over WithDelayFn
// and the Headers.ExpirationTime fallback when set. Non-positive values
// are ignored (the default zero is preserved, meaning "no fixed delay
// configured" so the next strategy is tried).
func WithFixedDelay[T any](d time.Duration) Option[T] {
	return func(opts *Options[T]) {
		if d > 0 {
			opts.fixedDelay = d
		}
	}
}

// WithDelayFn installs a function that computes the per-message delay
// from the message. Used when WithFixedDelay is not set. Nil values are
// ignored (the previously installed DelayFn is preserved). When DelayFn
// returns a non-positive duration the message forwards immediately.
func WithDelayFn[T any](fn DelayFn[T]) Option[T] {
	return func(opts *Options[T]) {
		if fn != nil {
			opts.delayFn = fn
		}
	}
}

// WithMaxPending caps the number of in-flight (scheduled but
// undelivered) messages. New messages over the bound are dropped via
// WithDropHandler with ErrDelayer(ErrMaxPendingExceeded). The default
// (defaultMaxPending) keeps the delayer's memory footprint bounded;
// non-positive values are ignored.
func WithMaxPending[T any](n int) Option[T] {
	return func(opts *Options[T]) {
		if n > 0 {
			opts.maxPending = n
		}
	}
}

// WithErrorHandler installs an observability hook fired once per real
// delayer failure (schedule rejection, forward Send failure). The
// default (when WithErrorHandler is not passed) is
// messaging.DefaultErrorHandler, which logs each failure via common/log
// so consumers that forget to wire observability still see real
// failures. Pass messaging.SilentErrorHandler to opt out, or any custom
// hook to redirect. Nil values are ignored (the previously installed
// handler is preserved).
func WithErrorHandler[T any](handler messaging.ErrorHandler) Option[T] {
	return func(opts *Options[T]) {
		if handler != nil {
			opts.errorHandler = handler
		}
	}
}

// WithDropHandler installs an observability hook fired once per
// intentional drop (pending queue exceeded WithMaxPending). The default
// (when WithDropHandler is not passed) is nil — drops are silent. Wire
// this to ship throughput metrics ("accepted vs dropped due to
// backpressure"). Nil arguments are ignored (the previously installed
// handler is preserved).
func WithDropHandler[T any](handler DropHandler) Option[T] {
	return func(opts *Options[T]) {
		if handler != nil {
			opts.dropHandler = handler
		}
	}
}
