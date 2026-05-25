package messaging

import (
	"context"
	"time"
)

// Default buffer, drain bounds, and worker pool size for the
// async channels.
const (
	defaultBufferSize   = 64
	defaultDrainTimeout = 5 * time.Second
	defaultWorkerCount  = 1
)

// ErrorHandler is the function type for the per-handler error
// observability hook installed on a TopicChannel via WithErrorHandler.
//
// The hook fires once per failed handler invocation, after the
// dispatcher has recovered any panic. err carries the handler's
// returned error or, on panic, an error wrapping ErrHandlerPanic
// with the recovered value. msg is type-erased; cast it inside the
// hook when payload-specific behavior is needed. The hook is invoked
// from the worker goroutine and must not block — long observability
// work should be dispatched asynchronously by the implementer.
type ErrorHandler func(ctx context.Context, msg any, err error)

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
// the given options.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		bufferSize:   defaultBufferSize,
		drainTimeout: defaultDrainTimeout,
		workerCount:  defaultWorkerCount,
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
// handler invocation that returns an error or panics. The default is
// a no-op — handler errors are silently dropped — so installing this
// is strongly recommended for any production wiring. Nil values are
// ignored.
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
