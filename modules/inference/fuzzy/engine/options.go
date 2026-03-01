package engine

import fuzzym "github.com/guidomantilla/yarumo/maths/fuzzy"

const defaultResolution = 100

// Options holds configuration for fuzzy engine execution.
type Options struct {
	method        Method
	tnorm         fuzzym.TNormFn
	tconorm       fuzzym.TConormFn
	defuzzify     fuzzym.DefuzzifyFn
	resolution    int
	sugenoOutputs map[string]float64
}

// Option is a functional option for configuring engine Options.
type Option func(*Options)

// NewOptions creates Options from the given functional options.
func NewOptions(opts ...Option) Options {
	o := Options{
		tnorm:      fuzzym.Min,
		tconorm:    fuzzym.Max,
		defuzzify:  fuzzym.Centroid,
		resolution: defaultResolution,
	}

	for _, opt := range opts {
		opt(&o)
	}

	return o
}

// WithMethod sets the inference method (Mamdani or Sugeno).
func WithMethod(m Method) Option {
	return func(o *Options) {
		if m >= Mamdani && m <= Sugeno {
			o.method = m
		}
	}
}

// WithTNorm sets the t-norm function for fuzzy AND.
func WithTNorm(fn fuzzym.TNormFn) Option {
	return func(o *Options) {
		if fn != nil {
			o.tnorm = fn
		}
	}
}

// WithTConorm sets the t-conorm function for fuzzy OR.
func WithTConorm(fn fuzzym.TConormFn) Option {
	return func(o *Options) {
		if fn != nil {
			o.tconorm = fn
		}
	}
}

// WithDefuzzify sets the defuzzification function.
func WithDefuzzify(fn fuzzym.DefuzzifyFn) Option {
	return func(o *Options) {
		if fn != nil {
			o.defuzzify = fn
		}
	}
}

// WithResolution sets the number of sampling points for defuzzification.
func WithResolution(n int) Option {
	return func(o *Options) {
		if n > 0 {
			o.resolution = n
		}
	}
}

// WithSugenoOutputs sets singleton output values for Sugeno method.
// The map keys use "variable/term" format.
func WithSugenoOutputs(outputs map[string]float64) Option {
	return func(o *Options) {
		if len(outputs) > 0 {
			o.sugenoOutputs = outputs
		}
	}
}
