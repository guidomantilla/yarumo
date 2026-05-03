package ed25519

// Option is a functional option for configuring ed25519 Options.
type Option func(opts *Options)

// Options holds the configuration for an Ed25519 Method.
type Options struct {
	keyFn    KeyFn
	signFn   SignFn
	verifyFn VerifyFn
}

// NewOptions creates a new Options with defaults and applies the given options.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		keyFn:    key,
		signFn:   sign,
		verifyFn: verify,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithKeyFn sets the key generation function.
func WithKeyFn(keyFn KeyFn) Option {
	return func(opts *Options) {
		if keyFn != nil {
			opts.keyFn = keyFn
		}
	}
}

// WithSignFn sets the signing function.
func WithSignFn(signFn SignFn) Option {
	return func(opts *Options) {
		if signFn != nil {
			opts.signFn = signFn
		}
	}
}

// WithVerifyFn sets the verification function.
func WithVerifyFn(verifyFn VerifyFn) Option {
	return func(opts *Options) {
		if verifyFn != nil {
			opts.verifyFn = verifyFn
		}
	}
}
