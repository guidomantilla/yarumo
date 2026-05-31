package controlbus

import (
	"context"

	"github.com/guidomantilla/yarumo/messaging"
)

// Option is a functional option for configuring controlbus Options.
// ControlBus has no T-typed options, so Option is non-generic — matching
// bridge and filter and diverging from router/Option[T] is intentional.
type Option func(opts *Options)

// Options holds the configuration for a ControlBus.
type Options struct {
	errorHandler         messaging.ErrorHandler
	unknownVerbHandler   Handler
}

// NewOptions creates a new Options with sensible defaults and applies
// the given options. The default ErrorHandler is
// messaging.DefaultErrorHandler (logs via common/log); the default
// UnknownVerbHandler returns Result{Success: false, Message:
// "unknown verb"}.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		errorHandler:       messaging.DefaultErrorHandler,
		unknownVerbHandler: defaultUnknownVerbHandler,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithErrorHandler installs an observability hook fired once per real
// control-bus failure (handler panic, reply-channel Send failure). The
// default (when WithErrorHandler is not passed) is
// messaging.DefaultErrorHandler, which logs each failure via common/log
// so consumers that forget to wire observability still see failures.
// Pass messaging.SilentErrorHandler to opt out, or any custom hook to
// redirect. Nil values are ignored (the previously installed handler
// is preserved).
func WithErrorHandler(handler messaging.ErrorHandler) Option {
	return func(opts *Options) {
		if handler != nil {
			opts.errorHandler = handler
		}
	}
}

// WithUnknownVerbHandler installs a fallback Handler that runs when the
// dispatched Command's Verb is not present in the handler registry. The
// default fallback returns Result{Success: false, Message: "unknown
// verb"}. Use this option to customise the unknown-verb response (e.g.
// return a help string listing the registered verbs). Nil values are
// ignored (the previously installed handler is preserved).
func WithUnknownVerbHandler(handler Handler) Option {
	return func(opts *Options) {
		if handler != nil {
			opts.unknownVerbHandler = handler
		}
	}
}

// defaultUnknownVerbHandler is the package-default fallback for unknown
// verbs. It returns Result{Success: false, Message: "unknown verb"}
// with the originating Command echoed back.
func defaultUnknownVerbHandler(_ context.Context, cmd Command) Result {
	return Result{
		Command: cmd,
		Success: false,
		Message: "unknown verb",
	}
}
