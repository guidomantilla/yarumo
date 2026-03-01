package engine

const defaultMaxIterations = 1000

// Options holds configuration for engine execution.
type Options struct {
	maxIterations int
	strategy      Strategy
}

// Option is a functional option for configuring engine Options.
type Option func(*Options)

// NewOptions creates Options from the given functional options.
func NewOptions(opts ...Option) Options {
	o := Options{
		maxIterations: defaultMaxIterations,
	}

	for _, opt := range opts {
		opt(&o)
	}

	return o
}

// WithMaxIterations sets the maximum number of forward chaining iterations.
func WithMaxIterations(n int) Option {
	return func(o *Options) {
		if n > 0 {
			o.maxIterations = n
		}
	}
}

// WithStrategy sets the conflict resolution strategy.
func WithStrategy(s Strategy) Option {
	return func(o *Options) {
		if s >= PriorityOrder && s <= FirstMatch {
			o.strategy = s
		}
	}
}
