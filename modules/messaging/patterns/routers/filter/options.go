package filter

import (
	"github.com/guidomantilla/yarumo/messaging"
)

// Option is a functional option for configuring filter Options. Filter
// has no T-typed options, so Option is non-generic — matching bridge
// and diverging from router/Option[T] is intentional.
type Option func(opts *Options)

// Options holds the configuration for a Filter.
type Options struct {
	errorHandler messaging.ErrorHandler
	dropHandler  DropHandler
}

// NewOptions creates a new Options with sensible defaults and applies
// the given options. The default ErrorHandler is
// messaging.DefaultErrorHandler (logs via common/log); the default
// DropHandler is nil (intentional drops are silent unless wired).
func NewOptions(opts ...Option) *Options {
	options := &Options{
		errorHandler: messaging.DefaultErrorHandler,
		dropHandler:  nil,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithErrorHandler installs an observability hook fired once per real
// filter failure (predicate returned error, predicate panicked,
// forward Send failed). The default (when WithErrorHandler is not
// passed) is messaging.DefaultErrorHandler, which logs each failure
// via common/log so consumers that forget to wire observability still
// see real failures. Pass messaging.SilentErrorHandler to opt out, or
// any custom hook to redirect. Nil values are ignored (the previously
// installed handler is preserved).
func WithErrorHandler(handler messaging.ErrorHandler) Option {
	return func(opts *Options) {
		if handler != nil {
			opts.errorHandler = handler
		}
	}
}

// WithDropHandler installs an observability hook fired once per
// intentional drop (predicate returned false). The default (when
// WithDropHandler is not passed) is nil — intentional drops are
// silent. Wire this to ship throughput metrics ("passed vs dropped")
// or audit trails. Nil arguments are ignored (the previously installed
// handler is preserved); to explicitly disable observability after a
// non-nil hook was installed, install messaging.SilentErrorHandler-
// equivalent: a `func(_ context.Context, _ any) {}` no-op.
func WithDropHandler(handler DropHandler) Option {
	return func(opts *Options) {
		if handler != nil {
			opts.dropHandler = handler
		}
	}
}
