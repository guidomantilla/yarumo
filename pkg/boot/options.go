package boot

type Options struct {
	Logger            BeanFn
	Config            BeanFn
	Validator         BeanFn
	PasswordEncoder   BeanFn
	PasswordGenerator BeanFn
	TokenGenerator    BeanFn
}

func NewOptions(opts ...Option) *Options {
	options := &Options{
		Logger:            Logger,
		Config:            Config,
		Validator:         Validator,
		PasswordEncoder:   PasswordEncoder,
		PasswordGenerator: PasswordGenerator,
		TokenGenerator:    TokenGenerator,
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

// WithPasswordEncoder allows setting a custom password encoder function into the WireContext (wctx *boot.WireContext).
//
// wctx.PasswordEncoder = <password encoder object>
func WithPasswordEncoder(passwordEncoderFn BeanFn) Option {
	return func(opts *Options) {
		opts.PasswordEncoder = passwordEncoderFn
	}
}

// WithPasswordGenerator allows setting a custom password generator function into the WireContext (wctx *boot.WireContext).
//
// wctx.PasswordGenerator = <password generator object>
func WithPasswordGenerator(passwordGeneratorFn BeanFn) Option {
	return func(opts *Options) {
		opts.PasswordGenerator = passwordGeneratorFn
	}
}

// WithTokenGenerator allows setting a custom token generator function into the WireContext (wctx *boot.WireContext).
//
// wctx.TokenGenerator = <token generator object>
func WithTokenGenerator(tokenGeneratorFn BeanFn) Option {
	return func(opts *Options) {
		opts.TokenGenerator = tokenGeneratorFn
	}
}
