package splitter

import (
	"github.com/guidomantilla/yarumo/messaging"
)

// Option is a functional option for configuring splitter Options. It is
// generic over U so options can carry U-typed values (e.g. future hooks
// that observe the produced Message[U] children) without losing type
// safety. None of the current options use U; the generic shape mirrors
// router/Option[T] and reserves the door for future extensions.
type Option[U any] func(opts *Options[U])

// Options holds the configuration for a Splitter.
type Options[U any] struct {
	errorHandler messaging.ErrorHandler
	dropHandler  DropHandler
}

// NewOptions creates a new Options[U] with sensible defaults and
// applies the given options. The default ErrorHandler is
// messaging.DefaultErrorHandler (logs via common/log); the default
// DropHandler is nil (empty-slice drops are silent unless wired).
func NewOptions[U any](opts ...Option[U]) *Options[U] {
	options := &Options[U]{
		errorHandler: messaging.DefaultErrorHandler,
		dropHandler:  nil,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithErrorHandler installs an observability hook fired once per real
// splitter failure (SplitFn returned error, SplitFn panicked, forward
// Send failed on any child). The default (when WithErrorHandler is not
// passed) is messaging.DefaultErrorHandler, which logs each failure
// via common/log so consumers that forget to wire observability still
// see real failures. Pass messaging.SilentErrorHandler to opt out, or
// any custom hook to redirect. Nil values are ignored (the previously
// installed handler is preserved).
func WithErrorHandler[U any](handler messaging.ErrorHandler) Option[U] {
	return func(opts *Options[U]) {
		if handler != nil {
			opts.errorHandler = handler
		}
	}
}

// WithDropHandler installs an observability hook fired once per
// intentional drop (SplitFn returned an empty slice). The default
// (when WithDropHandler is not passed) is nil — empty-slice drops are
// silent. Wire this to ship throughput metrics ("emitted vs dropped")
// or audit trails. Nil arguments are ignored (the previously installed
// handler is preserved).
func WithDropHandler[U any](handler DropHandler) Option[U] {
	return func(opts *Options[U]) {
		if handler != nil {
			opts.dropHandler = handler
		}
	}
}
