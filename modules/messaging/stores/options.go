package stores

import (
	"time"
)

// defaultSweepInterval is the cadence at which the in-memory metadata
// store's sweeper goroutine evicts expired entries when no
// WithSweepInterval option is supplied.
const defaultSweepInterval = time.Minute

// Option is a functional option for configuring store Options. Only
// the in-memory MetadataStore currently consumes Options; the in-
// memory MessageStore takes no configuration.
type Option func(opts *Options)

// Options holds the configuration for the in-memory backends in this
// package.
type Options struct {
	sweepInterval time.Duration
}

// NewOptions creates a new Options with sensible defaults and applies
// the given options. The default sweep interval is one minute — a
// compromise between memory pressure on a busy dedup store and
// goroutine wakeups on an idle one. Consumers with short TTLs (sub-
// second dedup windows) should pass WithSweepInterval to match.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		sweepInterval: defaultSweepInterval,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithSweepInterval configures how often the in-memory MetadataStore
// sweeper goroutine evicts expired entries. Values must be positive;
// non-positive values are ignored and the previously installed (or
// default) interval is preserved.
//
// The sweep interval is a hint, not a guarantee — entries may live
// past their TTL by up to one sweep interval before being evicted.
// Has correctly reports false for an expired-but-not-yet-swept entry
// (the TTL check happens on every Has call); the sweeper only frees
// the underlying map slot.
func WithSweepInterval(interval time.Duration) Option {
	return func(opts *Options) {
		if interval > 0 {
			opts.sweepInterval = interval
		}
	}
}
