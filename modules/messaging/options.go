package messaging

import (
	"time"
)

// Default buffer and drain bounds for QueueChannel.
const (
	defaultBufferSize   = 64
	defaultDrainTimeout = 5 * time.Second
)

// Option is a functional option for configuring messaging Options.
type Option func(opts *Options)

// Options holds the configuration for a QueueChannel.
type Options struct {
	bufferSize   int
	drainTimeout time.Duration
}

// NewOptions creates a new Options with sensible defaults and applies
// the given options.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		bufferSize:   defaultBufferSize,
		drainTimeout: defaultDrainTimeout,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithBufferSize sets the capacity of the in-memory queue used by the
// QueueChannel worker. Non-positive values are ignored.
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
