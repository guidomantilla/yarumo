package health

import "runtime"

// Option is a functional option for configuring health Options.
type Option func(opts *Options)

// Options holds the configuration for the [Health] aggregator.
type Options struct {
	concurrency int
}

// NewOptions creates Options with sensible defaults and applies the given functional options.
//
// Defaults:
//   - concurrency: runtime.NumCPU() (always >= 1).
func NewOptions(opts ...Option) *Options {
	options := &Options{
		concurrency: defaultConcurrency(),
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithConcurrency sets the maximum number of probes that may run in parallel
// during a single call to [Health.Status]. Values <= 0 are ignored and the
// default (runtime.NumCPU()) is preserved.
func WithConcurrency(concurrency int) Option {
	return func(opts *Options) {
		if concurrency > 0 {
			opts.concurrency = concurrency
		}
	}
}

// defaultConcurrency returns the default parallelism for probes.
// runtime.NumCPU always returns >= 1, and WithConcurrency rejects values
// <= 0, so callers never observe a non-positive concurrency.
func defaultConcurrency() int {
	return runtime.NumCPU()
}
