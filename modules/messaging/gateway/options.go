package gateway

import (
	"time"

	cuids "github.com/guidomantilla/yarumo/core/common/uids"
	"github.com/guidomantilla/yarumo/messaging"
)

// Option is a functional option for configuring gateway Options. Gateway
// has no T-typed options, so Option is non-generic — matching bridge and
// filter and diverging from router/Option[T] is intentional.
type Option func(opts *Options)

// Options holds the configuration for a Gateway.
type Options struct {
	errorHandler   messaging.ErrorHandler
	uid            cuids.UID
	requestTimeout time.Duration
}

// NewOptions creates a new Options with sensible defaults and applies
// the given options. The default ErrorHandler is
// messaging.DefaultErrorHandler (logs via common/log); the default
// requestTimeout is DefaultRequestTimeout (5s); the default uid is nil
// — callers MUST supply one via WithUIDGenerator or every Request will
// fail with ErrCorrelationIDFailed.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		errorHandler:   messaging.DefaultErrorHandler,
		requestTimeout: DefaultRequestTimeout,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithErrorHandler installs an observability hook fired once per real
// gateway failure (request-channel Send fail, reply with unknown
// CorrelationID). The default (when WithErrorHandler is not passed) is
// messaging.DefaultErrorHandler, which logs each failure via common/log
// so consumers that forget to wire observability still see failures.
// Pass messaging.SilentErrorHandler to opt out, or any custom hook to
// redirect. Nil values are ignored (the previously installed handler is
// preserved).
func WithErrorHandler(handler messaging.ErrorHandler) Option {
	return func(opts *Options) {
		if handler != nil {
			opts.errorHandler = handler
		}
	}
}

// WithUIDGenerator installs the cuids.UID generator used to mint
// CorrelationIDs for outgoing requests. The default is nil; callers
// MUST supply a generator or every Request will fail with
// ErrCorrelationIDFailed. The typical choice is a UUIDv4 or UUIDv7
// generator from modules/extension/common/uids/. Nil values are
// ignored (the previously installed generator is preserved).
func WithUIDGenerator(uid cuids.UID) Option {
	return func(opts *Options) {
		if uid != nil {
			opts.uid = uid
		}
	}
}

// WithRequestTimeout overrides the per-Request timeout used when the
// caller's ctx has no deadline. When the caller's ctx has a deadline
// tighter than this value, the ctx deadline wins. The default is
// DefaultRequestTimeout (5s). Non-positive values are ignored (the
// previously installed timeout is preserved).
func WithRequestTimeout(d time.Duration) Option {
	return func(opts *Options) {
		if d > 0 {
			opts.requestTimeout = d
		}
	}
}
