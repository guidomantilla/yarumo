package router

import (
	"github.com/guidomantilla/yarumo/core/common/messaging"
)

// Option is a functional option for configuring router Options. It is
// generic over T so options can carry T-typed values (e.g. the default
// destination Channel[T]) without losing type safety.
type Option[T any] func(opts *Options[T])

// Options holds the configuration for a Router.
type Options[T any] struct {
	defaultChannel messaging.Channel[T]
	errorHandler   messaging.ErrorHandler
}

// NewOptions creates a new Options[T] with sensible defaults and
// applies the given options. The default ErrorHandler is
// messaging.DefaultErrorHandler, which logs via common/log; pass
// WithErrorHandler(messaging.SilentErrorHandler) to opt out, or any
// custom hook to redirect.
func NewOptions[T any](opts ...Option[T]) *Options[T] {
	options := &Options[T]{
		errorHandler: messaging.DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithDefaultChannel installs a fallback destination Channel[T] for
// messages whose RouteFn key is not present in the routes map. When
// unset, NoRoute messages are dropped and forwarded to the
// ErrorHandler with ErrRoute(ErrNoRoute, ...). Nil values are ignored
// (the previously installed default is preserved).
func WithDefaultChannel[T any](ch messaging.Channel[T]) Option[T] {
	return func(opts *Options[T]) {
		if ch != nil {
			opts.defaultChannel = ch
		}
	}
}

// WithErrorHandler installs an observability hook fired once per
// routing failure (RouteFn error, RouteFn panic, NoRoute without
// default, forward Send failure). The default (when WithErrorHandler
// is not passed) is messaging.DefaultErrorHandler, which logs each
// failure via common/log so consumers that forget to wire
// observability still see routing failures. Pass
// messaging.SilentErrorHandler to opt out, or any custom hook to
// redirect. Nil values are ignored (the previously installed handler
// is preserved).
func WithErrorHandler[T any](handler messaging.ErrorHandler) Option[T] {
	return func(opts *Options[T]) {
		if handler != nil {
			opts.errorHandler = handler
		}
	}
}
