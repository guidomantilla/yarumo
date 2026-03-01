package hmacs

// Option is a functional option for configuring hmacs Options.
type Option func(opts *Options)

// Options holds the configuration for an HMAC Method.
type Options struct {
	keyFn      KeyFn
	digestFn   DigestFn
	validateFn ValidateFn
}

// NewOptions creates a new Options with defaults and applies the given options.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		keyFn:      key,
		digestFn:   digest,
		validateFn: validate,
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

// WithDigestFn sets the digest computation function.
func WithDigestFn(digestFn DigestFn) Option {
	return func(opts *Options) {
		if digestFn != nil {
			opts.digestFn = digestFn
		}
	}
}

// WithValidateFn sets the validation function.
func WithValidateFn(validateFn ValidateFn) Option {
	return func(opts *Options) {
		if validateFn != nil {
			opts.validateFn = validateFn
		}
	}
}
