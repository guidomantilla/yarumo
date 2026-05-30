package messaging

import (
	"time"
)

// Default buffer, drain bounds, worker pool size and overflow policy
// for the async channels.
const (
	defaultBufferSize     = 64
	defaultDrainTimeout   = 5 * time.Second
	defaultWorkerCount    = 1
	defaultOverflowPolicy = OverflowReject
)

// OverflowPolicy controls how an async Channel.Send reacts when the
// internal buffer is at capacity. The default for NewOptions is
// OverflowReject — Send returns ErrBufferFull immediately instead of
// blocking, forcing the caller to make an explicit decision about
// saturation. Pass WithOverflowPolicy(OverflowBlock) to opt into the
// historical blocking behavior.
type OverflowPolicy int

const (
	// OverflowBlock makes Send block until either a slot becomes
	// available or the caller's ctx expires. Guarantees no message is
	// dropped at the cost of propagating consumer slowness back to the
	// publisher.
	OverflowBlock OverflowPolicy = iota
	// OverflowDropNewest discards the message being sent when the
	// buffer is full. Send returns nil and the ErrorHandler hook fires
	// with ErrOverflow joined with ErrDropped. The buffer contents are
	// preserved; oldest messages keep priority.
	OverflowDropNewest
	// OverflowDropOldest evicts the oldest queued message and accepts
	// the new one when the buffer is full. Send returns nil and the
	// ErrorHandler hook fires with the evicted message and ErrOverflow
	// joined with ErrDropped. Newest messages keep priority.
	OverflowDropOldest
	// OverflowReject returns ErrBufferFull immediately when the buffer
	// is full instead of blocking, dropping, or evicting. The caller
	// is expected to implement retry / shedding / fallback logic.
	// This is the default for NewOptions.
	OverflowReject
)

// Option is a functional option for configuring messaging Options.
type Option func(opts *Options)

// Options holds the configuration for async channels (TopicChannel,
// QueueChannel).
type Options struct {
	bufferSize     int
	drainTimeout   time.Duration
	workerCount    int
	errorHandler   ErrorHandler
	overflowPolicy OverflowPolicy
}

// NewOptions creates a new Options with sensible defaults and applies
// the given options. Defaults: bufferSize 64, drainTimeout 5s,
// workerCount 1, ErrorHandler logs via common/log, overflowPolicy
// OverflowReject (Send returns ErrBufferFull when full instead of
// blocking). Pass WithOverflowPolicy(OverflowBlock) to opt into the
// historical blocking-Send behavior.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		bufferSize:     defaultBufferSize,
		drainTimeout:   defaultDrainTimeout,
		workerCount:    defaultWorkerCount,
		errorHandler:   DefaultErrorHandler,
		overflowPolicy: defaultOverflowPolicy,
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

// WithOverflowPolicy selects the strategy Send uses when the async
// channel's internal buffer is at capacity. See OverflowPolicy for
// the four available strategies. Values outside the defined range
// are ignored (the previously configured policy is preserved).
func WithOverflowPolicy(p OverflowPolicy) Option {
	return func(opts *Options) {
		if p >= OverflowBlock && p <= OverflowReject {
			opts.overflowPolicy = p
		}
	}
}
