package expressions

// Options holds configuration for expression evaluation.
type Options struct {
	funcs map[string]Func
}

// Option is a functional option for configuring expression Options.
type Option func(*Options)

// NewOptions creates Options from the given functional options.
func NewOptions(opts ...Option) *Options {
	o := &Options{
		funcs: DefaultFuncs(),
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}

// WithFunc registers a custom function for use in expressions.
func WithFunc(name string, fn Func) Option {
	return func(o *Options) {
		if fn != nil {
			o.funcs[name] = fn
		}
	}
}
