package variable

const defaultResolution = 100

// Options holds configuration for variable creation.
type Options struct {
	resolution int
}

// Option is a functional option for configuring variable Options.
type Option func(*Options)

// NewOptions creates Options from the given functional options.
func NewOptions(opts ...Option) Options {
	o := Options{
		resolution: defaultResolution,
	}

	for _, opt := range opts {
		opt(&o)
	}

	return o
}

// WithResolution sets the number of sampling points for defuzzification.
func WithResolution(n int) Option {
	return func(o *Options) {
		if n > 0 {
			o.resolution = n
		}
	}
}
