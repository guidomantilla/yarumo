package cryptos

import "encoding/base64"

type AesCipherOptions struct {
	key []byte
}

func NewAesCipherOptions(opts ...AesCipherOption) *AesCipherOptions {
	options := &AesCipherOptions{
		key: func() []byte {
			key, _ := Key(32)
			b, _ := base64.StdEncoding.DecodeString(*key)
			return b
		}(),
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

type AesCipherOption func(opts *AesCipherOptions)

func WithAesCipherKeySize32() AesCipherOption {
	return func(opts *AesCipherOptions) {
		key, _ := Key(32)
		opts.key = []byte(*key)
	}
}

func WithAesCipherKey(key string) AesCipherOption {
	return func(opts *AesCipherOptions) {
		b, _ := base64.StdEncoding.DecodeString(key)
		opts.key = b
	}
}
