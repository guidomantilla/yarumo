package hashes

// Option is a functional option for configuring hashes Options.
type Option func(opts *Options)

// Options holds the configuration for a hash Method.
type Options struct {
	hashFn HashFn
}

// NewOptions creates a new Options with defaults and applies the given options.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		hashFn: Hash,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithHashFn sets the hash function used by the Method.
func WithHashFn(hashFn HashFn) Option {
	return func(opts *Options) {
		if hashFn != nil {
			opts.hashFn = hashFn
		}
	}
}
