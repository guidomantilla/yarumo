package parser

// options holds configuration for the parser.
type options struct{}

// Option is a functional option for configuring the parser.
type Option func(*options)

func newOptions(opts ...Option) options {
	o := options{}

	for _, opt := range opts {
		opt(&o)
	}

	return o
}
