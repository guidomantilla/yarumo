package diagnostics

import (
	"time"
)

// blockProfileRateDefault is the default sampling rate for
// BlockProfiling — sample every blocking event. Production callers
// should usually pass a coarser rate via WithBlockProfileRate.
const blockProfileRateDefault = 1

// Option is a functional option for configuring diagnostics Options.
type Option func(opts *Options)

// Options holds the configuration for diagnostic components.
type Options struct {
	minAge           time.Duration
	maxBytes         uint64
	blockProfileRate int
}

// NewOptions creates a new Options with sensible defaults and applies the given options.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		minAge:           10 * time.Second,
		maxBytes:         10 << 20,
		blockProfileRate: blockProfileRateDefault,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithMinAge sets the minimum age of trace data the flight recorder
// retains in its buffer. Zero or negative values are ignored.
func WithMinAge(minAge time.Duration) Option {
	return func(opts *Options) {
		if minAge > 0 {
			opts.minAge = minAge
		}
	}
}

// WithMaxBytes sets the maximum number of bytes the flight recorder
// may use for its buffer. Zero values are ignored.
func WithMaxBytes(maxBytes uint64) Option {
	return func(opts *Options) {
		if maxBytes > 0 {
			opts.maxBytes = maxBytes
		}
	}
}

// WithBlockProfileRate sets the BlockProfiling sampling rate (passed
// to runtime.SetBlockProfileRate when Start is called). A rate of 1
// samples every blocking event; higher values sample more sparsely.
// Zero or negative values are ignored.
func WithBlockProfileRate(rate int) Option {
	return func(opts *Options) {
		if rate > 0 {
			opts.blockProfileRate = rate
		}
	}
}
