package validation

import (
	cexpressions "github.com/guidomantilla/yarumo/core/common/expressions"
)

// Options holds the engine configuration.
type Options struct {
	registry      *Registry
	evaluator     cexpressions.Evaluator
	hook          Hook
	lintOnLoad    bool
	strictVersion bool
}

// Option is a functional option for configuring engine Options.
type Option func(*Options)

// NewOptions creates Options from the given functional options. Sensible
// defaults are installed for any unset field.
func NewOptions(opts ...Option) *Options {
	o := &Options{
		registry:  DefaultRegistry(),
		evaluator: cexpressions.NewEvaluator(),
		hook:      NoopHook{},
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

// WithLintOnLoad enables fail-at-boot lint inside BuildEngine: the
// constructor runs Validate(ruleset) before returning the Engine and
// surfaces every structural issue as an error. The default constructor
// NewEngine ignores this flag.
func WithLintOnLoad() Option {
	return func(o *Options) {
		o.lintOnLoad = true
	}
}

// WithStrictVersion enables strict schema-version checking: a ruleset
// without a Version field is rejected, and any non-matching Version fails
// the lint. Without the flag, an empty Version is accepted unconditionally
// and only non-empty mismatches are reported.
func WithStrictVersion() Option {
	return func(o *Options) {
		o.strictVersion = true
	}
}

// WithHook installs an observability hook fired before and after every
// leaf evaluation. Pass MultiHook to compose several.
func WithHook(h Hook) Option {
	return func(o *Options) {
		if h != nil {
			o.hook = h
		}
	}
}
