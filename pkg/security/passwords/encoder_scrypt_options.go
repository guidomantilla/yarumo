package passwords

const (
	ScryptN          = 32768
	ScryptR          = 8
	ScryptP          = 1
	ScryptSaltLength = 16
	ScryptKeyLength  = 32
)

type ScryptEncoderOptions struct {
	N          int
	r          int
	p          int
	saltLength int
	keyLength  int
}

func NewScryptEncoderOptions(opts ...ScryptEncoderOption) *ScryptEncoderOptions {
	options := &ScryptEncoderOptions{
		N:          ScryptN,
		r:          ScryptR,
		p:          ScryptP,
		saltLength: ScryptSaltLength,
		keyLength:  ScryptKeyLength,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

type ScryptEncoderOption func(opts *ScryptEncoderOptions)

func WithScryptN(N int) ScryptEncoderOption {
	return func(opts *ScryptEncoderOptions) {
		if N > ScryptN {
			opts.N = N
		}
	}
}

func WithScryptR(r int) ScryptEncoderOption {
	return func(opts *ScryptEncoderOptions) {
		if r > ScryptR {
			opts.r = r
		}
	}
}

func WithScryptP(p int) ScryptEncoderOption {
	return func(opts *ScryptEncoderOptions) {
		if p > ScryptP {
			opts.p = p
		}
	}
}

func WithScryptSaltLength(saltLength int) ScryptEncoderOption {
	return func(opts *ScryptEncoderOptions) {
		if saltLength > ScryptSaltLength {
			opts.saltLength = saltLength
		}
	}
}

func WithScryptKeyLength(keyLength int) ScryptEncoderOption {
	return func(opts *ScryptEncoderOptions) {
		if keyLength > ScryptKeyLength {
			opts.keyLength = keyLength
		}
	}
}
