package bridge

import (
	"github.com/guidomantilla/yarumo/messaging"
)

// Option is a functional option for configuring bridge Options. Bridge
// has no T-typed options, so Option is non-generic — diverging from
// router/Option[T] is intentional and keeps the callsite simpler.
type Option func(opts *Options)

// Options holds the configuration for a Bridge.
type Options struct {
	errorHandler messaging.ErrorHandler
}

// NewOptions creates a new Options with sensible defaults and applies
// the given options. The default ErrorHandler is
// messaging.DefaultErrorHandler, which logs via common/log; pass
// WithErrorHandler(messaging.SilentErrorHandler) to opt out, or any
// custom hook to redirect.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		errorHandler: messaging.DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithErrorHandler installs an observability hook fired once per
// forward Send failure. The default (when WithErrorHandler is not
// passed) is messaging.DefaultErrorHandler, which logs each failure
// via common/log so consumers that forget to wire observability still
// see forward failures. Pass messaging.SilentErrorHandler to opt out,
// or any custom hook to redirect. Nil values are ignored (the
// previously installed handler is preserved).
func WithErrorHandler(handler messaging.ErrorHandler) Option {
	return func(opts *Options) {
		if handler != nil {
			opts.errorHandler = handler
		}
	}
}
