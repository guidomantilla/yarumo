package engine

import "github.com/guidomantilla/yarumo/maths/probability"

// Options holds configuration for Bayesian engine execution.
type Options struct {
	algorithm        Algorithm
	eliminationOrder []probability.Var
}

// Option is a functional option for configuring engine Options.
type Option func(*Options)

// NewOptions creates Options from the given functional options.
func NewOptions(opts ...Option) Options {
	o := Options{}

	for _, opt := range opts {
		opt(&o)
	}

	return o
}

// WithAlgorithm sets the inference algorithm.
func WithAlgorithm(a Algorithm) Option {
	return func(o *Options) {
		if a >= Enumeration && a <= VariableElimination {
			o.algorithm = a
		}
	}
}

// WithEliminationOrder sets the variable elimination order.
func WithEliminationOrder(order []probability.Var) Option {
	return func(o *Options) {
		if len(order) > 0 {
			o.eliminationOrder = order
		}
	}
}
