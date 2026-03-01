package parser

// Options holds configuration for the parser.
type Options struct {
	strict bool
}

// Option is a functional option for configuring parser Options.
type Option func(*Options)

// NewOptions creates Options with the given functional options applied.
func NewOptions(opts ...Option) Options {
	o := Options{}

	for _, opt := range opts {
		opt(&o)
	}

	return o
}

// WithStrict enables strict mode, which rejects Unicode operator synonyms
// and keyword synonyms, accepting only the canonical operators.
func WithStrict(strict bool) Option {
	return func(o *Options) {
		if strict {
			o.strict = strict
		}
	}
}
