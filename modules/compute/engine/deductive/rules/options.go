package rules

// Options holds configuration for rule creation.
type Options struct {
	priority int
}

// Option is a functional option for configuring rule Options.
type Option func(*Options)

// NewOptions creates Options from the given functional options.
func NewOptions(opts ...Option) Options {
	o := Options{}

	for _, opt := range opts {
		opt(&o)
	}

	return o
}

// WithPriority sets the rule priority (lower = higher priority).
func WithPriority(priority int) Option {
	return func(o *Options) {
		if priority >= 0 {
			o.priority = priority
		}
	}
}
