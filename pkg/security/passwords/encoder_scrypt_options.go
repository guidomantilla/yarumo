package passwords

type ScryptEncoderOptions struct {
	N          int
	r          int
	p          int
	saltLength int
	keyLength  int
}

func NewScryptEncoderOptions(opts ...ScryptEncoderOption) *ScryptEncoderOptions {
	options := &ScryptEncoderOptions{
		N:          32768,
		r:          8,
		p:          1,
		saltLength: 16,
		keyLength:  32,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

type ScryptEncoderOption func(opts *ScryptEncoderOptions)

func WithScryptN(N int) ScryptEncoderOption {
	return func(opts *ScryptEncoderOptions) {
		opts.N = N
	}
}

func WithScryptR(r int) ScryptEncoderOption {
	return func(opts *ScryptEncoderOptions) {
		opts.r = r
	}
}

func WithScryptP(p int) ScryptEncoderOption {
	return func(opts *ScryptEncoderOptions) {
		opts.p = p
	}
}

func WithScryptSaltLength(saltLength int) ScryptEncoderOption {
	return func(opts *ScryptEncoderOptions) {
		opts.saltLength = saltLength
	}
}

func WithScryptKeyLength(keyLength int) ScryptEncoderOption {
	return func(opts *ScryptEncoderOptions) {
		opts.keyLength = keyLength
	}
}
