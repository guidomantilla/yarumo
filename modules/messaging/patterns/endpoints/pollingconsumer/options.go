package pollingconsumer

import (
	"time"

	"github.com/guidomantilla/yarumo/messaging"
)

// defaultMaxConcurrency is the worker pool size when WithMaxConcurrency
// is not configured: a strictly sequential consumer.
const defaultMaxConcurrency = 1

// Option is a functional option for configuring pollingconsumer
// Options. Polling consumer has no T-typed options, so Option is
// non-generic — matching bridge and filter.
type Option func(opts *Options)

// Options holds the configuration for a PollingConsumer.
type Options struct {
	pollInterval   time.Duration
	maxConcurrency int
	errorHandler   messaging.ErrorHandler
}

// NewOptions creates a new Options with sensible defaults and applies
// the given options. Defaults: pollInterval 0 (poll immediately each
// iteration, relying on PollableChannel.Receive for backpressure);
// maxConcurrency 1 (sequential consumer); ErrorHandler logs via
// common/log.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		maxConcurrency: defaultMaxConcurrency,
		errorHandler:   messaging.DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithPollInterval inserts a pause between Receive calls. The default
// 0 polls immediately each iteration, relying on the PollableChannel's
// blocking Receive for natural backpressure; non-zero values rate-limit
// the worker. Non-positive values are ignored (the default is
// preserved).
func WithPollInterval(d time.Duration) Option {
	return func(opts *Options) {
		if d > 0 {
			opts.pollInterval = d
		}
	}
}

// WithMaxConcurrency sets the size of the worker pool. With N > 1, N
// workers poll the same PollableChannel concurrently and the Handler
// must be safe for concurrent invocation. The default 1 is a strictly
// sequential consumer. Non-positive values are ignored.
func WithMaxConcurrency(n int) Option {
	return func(opts *Options) {
		if n > 0 {
			opts.maxConcurrency = n
		}
	}
}

// WithErrorHandler installs an observability hook fired once per real
// consumer failure (Handler error, Handler panic, unexpected Receive
// error). The default (when WithErrorHandler is not passed) is
// messaging.DefaultErrorHandler, which logs each failure via
// common/log so consumers that forget to wire observability still see
// failures. Pass messaging.SilentErrorHandler to opt out, or any
// custom hook to redirect. Nil values are ignored (the previously
// installed handler is preserved).
func WithErrorHandler(handler messaging.ErrorHandler) Option {
	return func(opts *Options) {
		if handler != nil {
			opts.errorHandler = handler
		}
	}
}
