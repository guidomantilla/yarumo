package boot

type Options struct {
	Hasher            BeanFn
	UIDGen            BeanFn
	Logger            BeanFn
	Config            BeanFn
	Validator         BeanFn
	PasswordEncoder   BeanFn
	PasswordGenerator BeanFn
	TokenGenerator    BeanFn
	Cipher            BeanFn
}

func NewOptions(opts ...Option) *Options {
	options := &Options{
		Hasher:            Hasher,
		UIDGen:            UIDGen,
		Logger:            Logger,
		Config:            Config,
		Validator:         Validator,
		PasswordEncoder:   PasswordEncoder,
		PasswordGenerator: PasswordGenerator,
		TokenGenerator:    TokenGenerator,
		Cipher:            Cipher,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

type Option func(opts *Options)

// WithHasher allows setting a custom hasher function into the WireContext (wctx *boot.WireContext).
//
// wctx.Hasher = <hasher object>
func WithHasher(hasherFn BeanFn) Option {
	return func(opts *Options) {
		opts.Hasher = hasherFn
	}
}

// WithUIDGen allows setting a custom UID generator function into the WireContext (wctx *boot.WireContext).
//
// wctx.UIDGen = <uid generator object>
func WithUIDGen(uidGenFn BeanFn) Option {
	return func(opts *Options) {
		opts.UIDGen = uidGenFn
	}
}

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

// WithCipher allows setting a custom cipher function into the WireContext (wctx *boot.WireContext).
//
// wctx.Cipher = <cipher object>
func WithCipher(cipherFn BeanFn) Option {
	return func(opts *Options) {
		opts.Cipher = cipherFn
	}
}
