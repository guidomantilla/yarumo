package rsaoaep

// Option is a functional option for configuring rsaoaep Options.
type Option func(opts *Options)

// Options holds the configuration for an RSA-OAEP Method.
type Options struct {
	keyFn     KeyFn
	encryptFn EncryptFn
	decryptFn DecryptFn
}

// NewOptions creates a new Options with defaults and applies the given options.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		keyFn:     key,
		encryptFn: encrypt,
		decryptFn: decrypt,
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

// WithEncryptFn sets the encryption function.
func WithEncryptFn(encryptFn EncryptFn) Option {
	return func(opts *Options) {
		if encryptFn != nil {
			opts.encryptFn = encryptFn
		}
	}
}

// WithDecryptFn sets the decryption function.
func WithDecryptFn(decryptFn DecryptFn) Option {
	return func(opts *Options) {
		if decryptFn != nil {
			opts.decryptFn = decryptFn
		}
	}
}
