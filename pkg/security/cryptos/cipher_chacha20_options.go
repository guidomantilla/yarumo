package cryptos

import "encoding/base64"

type ChaCha20CipherOptions struct {
	key []byte
}

func NewChaCha20CipherOptions(opts ...ChaCha20CipherOption) *ChaCha20CipherOptions {
	options := &ChaCha20CipherOptions{
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

type ChaCha20CipherOption func(opts *ChaCha20CipherOptions)

func WithChaCha20CipherKeySize32() ChaCha20CipherOption {
	return func(opts *ChaCha20CipherOptions) {
		key, _ := Key(32)
		opts.key = []byte(*key)
	}
}

func WithChaCha20CipherKey(key string) ChaCha20CipherOption {
	return func(opts *ChaCha20CipherOptions) {
		b, _ := base64.StdEncoding.DecodeString(key)
		opts.key = b
	}
}
