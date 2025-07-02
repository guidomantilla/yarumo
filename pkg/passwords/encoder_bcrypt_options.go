package passwords

import "golang.org/x/crypto/bcrypt"

type BcryptEncoderOptions struct {
	cost int
}

func NewBcryptEncoderOptions(opts ...BcryptEncoderOption) *BcryptEncoderOptions {
	options := &BcryptEncoderOptions{
		cost: bcrypt.DefaultCost,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

type BcryptEncoderOption func(opts *BcryptEncoderOptions)

func WithBcryptCost(cost int) BcryptEncoderOption {
	return func(opts *BcryptEncoderOptions) {
		opts.cost = cost
	}
}
