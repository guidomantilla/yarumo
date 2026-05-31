package recipientlist

import (
	"github.com/guidomantilla/yarumo/messaging"
)

// Option is a functional option for configuring recipientlist Options.
// Option is non-generic — recipientlist has no T-typed options.
type Option func(opts *Options)

// Options holds the configuration for a RecipientList.
type Options struct {
	errorHandler messaging.ErrorHandler
	dropHandler  DropHandler
}

// NewOptions creates a new Options with sensible defaults and applies
// the given options. The default ErrorHandler is
// messaging.DefaultErrorHandler (logs via common/log); the default
// DropHandler is nil (empty-selection drops are silent unless wired).
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

// WithErrorHandler installs an observability hook fired once per
// per-recipient routing failure (SelectorFn returned error, SelectorFn
// panicked, missing key for a recipient, forward Send failed for a
// recipient). The default (when WithErrorHandler is not passed) is
// messaging.DefaultErrorHandler, which logs each failure via common/log
// so consumers that forget to wire observability still see routing
// failures. Pass messaging.SilentErrorHandler to opt out, or any custom
// hook to redirect. Nil values are ignored (the previously installed
// handler is preserved).
func WithErrorHandler(handler messaging.ErrorHandler) Option {
	return func(opts *Options) {
		if handler != nil {
			opts.errorHandler = handler
		}
	}
}

// WithDropHandler installs an observability hook fired once per
// intentional drop (SelectorFn returned an empty slice). The default
// (when WithDropHandler is not passed) is nil — empty-selection drops
// are silent. Wire this to ship throughput metrics ("dispatched vs
// nobody-cared") or audit trails. Nil arguments are ignored (the
// previously installed handler is preserved).
func WithDropHandler(handler DropHandler) Option {
	return func(opts *Options) {
		if handler != nil {
			opts.dropHandler = handler
		}
	}
}
