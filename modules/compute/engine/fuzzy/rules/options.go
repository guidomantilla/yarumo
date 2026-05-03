package rules

// Options holds configuration for rule creation.
type Options struct {
	operator Operator
	weight   float64
}

// Option is a functional option for configuring rule Options.
type Option func(*Options)

// NewOptions creates Options from the given functional options.
func NewOptions(opts ...Option) Options {
	o := Options{
		weight: 1.0,
	}

	for _, opt := range opts {
		opt(&o)
	}

	return o
}

// WithOperator sets the condition combination operator.
func WithOperator(op Operator) Option {
	return func(o *Options) {
		if op >= And && op <= Or {
			o.operator = op
		}
	}
}

// WithWeight sets the rule weight in [0,1].
func WithWeight(w float64) Option {
	return func(o *Options) {
		if w >= 0 && w <= 1 {
			o.weight = w
		}
	}
}
