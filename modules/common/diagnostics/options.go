package diagnostics

import (
	"time"
)

// Option is a functional option for configuring diagnostics Options.
type Option func(opts *Options)

// Options holds the configuration for diagnostic components.
type Options struct {
	minAge   time.Duration
	maxBytes uint64
}

// NewOptions creates a new Options with sensible defaults and applies the given options.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		minAge:   10 * time.Second,
		maxBytes: 10 << 20,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithMinAge sets the minimum age of trace data to retain.
func WithMinAge(minAge time.Duration) Option {
	return func(opts *Options) {
		if minAge > 0 {
			opts.minAge = minAge
		}
	}
}

// WithMaxBytes sets the maximum number of bytes the flight recorder may use.
func WithMaxBytes(maxBytes uint64) Option {
	return func(opts *Options) {
		if maxBytes > 0 {
			opts.maxBytes = maxBytes
		}
	}
}
