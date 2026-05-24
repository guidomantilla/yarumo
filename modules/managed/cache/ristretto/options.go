package ristretto

import (
	"time"
)

// Option is a functional option for configuring Options.
type Option func(opts *Options)

// Options holds the configuration applied at cache construction time.
type Options struct {
	ttl       time.Duration
	keyPrefix string

	numCtrs  int64
	maxCost  int64
	bufItems int64
}

// NewOptions creates Options with safe defaults and applies the given functional options.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		ttl:       5 * time.Minute,
		keyPrefix: "",

		numCtrs:  1_000_000,
		maxCost:  100 << 20,
		bufItems: 64,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithTTL sets the default time-to-live applied to entries when Set is called
// with a non-positive ttl. Values less than or equal to zero are ignored.
func WithTTL(ttl time.Duration) Option {
	return func(opts *Options) {
		if ttl > 0 {
			opts.ttl = ttl
		}
	}
}

// WithKeyPrefix overrides the key prefix used to namespace cache keys.
// Effective prefix is "<name>:" when this option is not provided. Empty
// values are ignored, preserving the default.
func WithKeyPrefix(prefix string) Option {
	return func(opts *Options) {
		if prefix != "" {
			opts.keyPrefix = prefix
		}
	}
}

// WithCapacity overrides the ristretto counter count, max cost and buffer
// item size. Non-positive values are ignored per-parameter.
func WithCapacity(numCounters int64, maxCost int64, bufferItems int64) Option {
	return func(opts *Options) {
		if numCounters > 0 {
			opts.numCtrs = numCounters
		}
		if maxCost > 0 {
			opts.maxCost = maxCost
		}
		if bufferItems > 0 {
			opts.bufItems = bufferItems
		}
	}
}
