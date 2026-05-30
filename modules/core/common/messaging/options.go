package messaging

import (
	"time"
)

// Default buffer, drain bounds, and worker pool size for the
// async channels.
const (
	defaultBufferSize   = 64
	defaultDrainTimeout = 5 * time.Second
	defaultWorkerCount  = 1
)

// Option is a functional option for configuring messaging Options.
type Option func(opts *Options)

// Options holds the configuration for async channels (TopicChannel,
// QueueChannel).
type Options struct {
	bufferSize   int
	drainTimeout time.Duration
	workerCount  int
	errorHandler ErrorHandler
}

// NewOptions creates a new Options with sensible defaults and applies
// the given options. The default ErrorHandler logs handler failures
// via common/log; pass WithErrorHandler(SilentErrorHandler) to opt
// out, or any custom hook to redirect.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		bufferSize:   defaultBufferSize,
		drainTimeout: defaultDrainTimeout,
		workerCount:  defaultWorkerCount,
		errorHandler: DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithBufferSize sets the capacity of the in-memory queue used by the
// TopicChannel worker. Non-positive values are ignored.
func WithBufferSize(size int) Option {
	return func(opts *Options) {
		if size > 0 {
			opts.bufferSize = size
		}
	}
}

// WithDrainTimeout sets the maximum time Stop waits for pending
// messages to drain before returning. Non-positive values are ignored.
func WithDrainTimeout(timeout time.Duration) Option {
	return func(opts *Options) {
		if timeout > 0 {
			opts.drainTimeout = timeout
		}
	}
}

// WithErrorHandler installs an observability hook fired once per
// handler invocation that returns an error or panics. The default
// (when WithErrorHandler is not passed) is DefaultErrorHandler, which
// logs each failure via common/log so consumers that forget to wire
// observability still see handler failures. Pass SilentErrorHandler
// to opt out, or any custom hook to redirect. Nil values are ignored
// (the previously installed handler is preserved).
func WithErrorHandler(handler ErrorHandler) Option {
	return func(opts *Options) {
		if handler != nil {
			opts.errorHandler = handler
		}
	}
}

// WithWorkerCount sets the number of worker goroutines a
// QueueChannel spawns to consume from the inbound buffer. Workers
// compete for messages — each message goes to exactly one worker
// (and from there to exactly one subscriber via round-robin). Non-
// positive values are ignored.
func WithWorkerCount(n int) Option {
	return func(opts *Options) {
		if n > 0 {
			opts.workerCount = n
		}
	}
}
