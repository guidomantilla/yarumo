package boot

type Options struct {
	Logger    BeanFn
	Config    BeanFn
	Validator BeanFn
}

func NewOptions(opts ...Option) *Options {
	options := &Options{
		Logger:    Logger,
		Config:    Config,
		Validator: Validator,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

type Option func(opts *Options)

// WithLogger allows setting a custom logger function into the WireContext (wctx *boot.WireContext).
//
// wctx.Logger = <logger object>
func WithLogger(loggerFn BeanFn) Option {
	return func(opts *Options) {
		opts.Logger = loggerFn
	}
}

// WithConfig allows setting a custom config function into the WireContext (wctx *boot.WireContext).
//
// wctx.Config = <config object>
func WithConfig(configFn BeanFn) Option {
	return func(opts *Options) {
		opts.Config = configFn
	}
}

// WithValidator allows setting a custom validator function into the WireContext (wctx *boot.WireContext).
//
// wctx.Validator = <validator object>
func WithValidator(validatorFn BeanFn) Option {
	return func(opts *Options) {
		opts.Validator = validatorFn
	}
}

//

type OptionsChain struct {
	chain []Option
}

func Chain() *OptionsChain {
	return &OptionsChain{
		chain: make([]Option, 0),
	}
}

func (chain *OptionsChain) Build() Option {
	return func(opts *Options) {
		for _, option := range chain.chain {
			option(opts)
		}
	}
}

// WithLogger allows setting a custom logger function into the WireContext (wctx *boot.WireContext).
//
// wctx.Logger = <config object>
func (chain *OptionsChain) WithLogger(loggerFn BeanFn) *OptionsChain {
	chain.chain = append(chain.chain, WithLogger(loggerFn))
	return chain
}

// WithConfig allows setting a custom config function into the WireContext (wctx *boot.WireContext).
//
// wctx.Config = <config object>
func (chain *OptionsChain) WithConfig(configFn BeanFn) *OptionsChain {
	chain.chain = append(chain.chain, WithConfig(configFn))
	return chain
}

// WithValidator allows setting a custom validator function into the WireContext (wctx *boot.WireContext).
//
// wctx.Validator = <validator object>
func (chain *OptionsChain) WithValidator(validatorFn BeanFn) *OptionsChain {
	chain.chain = append(chain.chain, WithValidator(validatorFn))
	return chain
}
