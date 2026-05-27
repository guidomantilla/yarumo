package messaging

import (
	"context"
	"time"

	clog "github.com/guidomantilla/yarumo/core/common/log"
)

// Default buffer, drain bounds, and worker pool size for the
// async channels.
const (
	defaultBufferSize   = 64
	defaultDrainTimeout = 5 * time.Second
	defaultWorkerCount  = 1
)

// ErrorHandler is the function type for the per-handler error
// observability hook installed on a TopicChannel or QueueChannel via
// WithErrorHandler.
//
// The hook fires once per failed handler invocation, after the
// dispatcher has recovered any panic. err carries the handler's
// returned error or, on panic, an error wrapping ErrHandlerPanic
// with the recovered value. msg is type-erased; cast it inside the
// hook when payload-specific behavior is needed. The hook is invoked
// from the worker goroutine and must not block — long observability
// work should be dispatched asynchronously by the implementer.
//
// The default hook logs every failure via common/log so a consumer
// that forgets to wire observability still gets a record of handler
// errors. Callers that genuinely want silence must opt out by
// installing a no-op hook explicitly (see DefaultErrorHandler and
// SilentErrorHandler below).
type ErrorHandler func(ctx context.Context, msg any, err error)

// DefaultErrorHandler is the hook installed by NewOptions when the
// caller does not pass WithErrorHandler. It logs every failure via
// common/log at Error level so handler bugs surface in standard
// telemetry without explicit caller wiring.
func DefaultErrorHandler(ctx context.Context, _ any, err error) {
	clog.Error(ctx, "messaging handler failed",
		"error", err.Error(),
	)
}

// SilentErrorHandler is a no-op ErrorHandler. Use it when the caller
// genuinely wants to suppress error logging — for example, in tests
// that intentionally drive failure paths.
func SilentErrorHandler(_ context.Context, _ any, _ error) {}

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
