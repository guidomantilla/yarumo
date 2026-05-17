package validation

import (
	cexpressions "github.com/guidomantilla/yarumo/common/expressions"
)

// Options holds the engine configuration.
type Options struct {
	registry  *Registry
	evaluator cexpressions.Evaluator
}

// Option is a functional option for configuring engine Options.
type Option func(*Options)

// NewOptions creates Options from the given functional options. Sensible
// defaults are installed for any unset field.
func NewOptions(opts ...Option) *Options {
	o := &Options{
		registry:  DefaultRegistry(),
		evaluator: cexpressions.NewEvaluator(),
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}

// WithRegistry sets the registry used to resolve leaf rule names. Nil values
// are silently ignored, preserving the default registry.
func WithRegistry(r *Registry) Option {
	return func(o *Options) {
		if r != nil {
			o.registry = r
		}
	}
}

// WithEvaluator sets the expression evaluator used for "when" predicates.
// Nil values are silently ignored, preserving the default evaluator.
func WithEvaluator(e cexpressions.Evaluator) Option {
	return func(o *Options) {
		if e != nil {
			o.evaluator = e
		}
	}
}
