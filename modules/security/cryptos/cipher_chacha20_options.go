package cryptos

import (
	"encoding/base64"

	"github.com/guidomantilla/yarumo/common/random"
	"github.com/guidomantilla/yarumo/common/utils"
)

type ChaCha20CipherOptions struct {
	key []byte
}

func NewChaCha20CipherOptions(opts ...ChaCha20CipherOption) *ChaCha20CipherOptions {
	options := &ChaCha20CipherOptions{
		key: random.Key(32),
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

type ChaCha20CipherOption func(opts *ChaCha20CipherOptions)

func WithChaCha20CipherKeySize32() ChaCha20CipherOption {
	return func(opts *ChaCha20CipherOptions) {
		opts.key = random.Key(32)
	}
}

func WithChaCha20CipherKey(key string) ChaCha20CipherOption {
	return func(opts *ChaCha20CipherOptions) {
		if utils.NotEmpty(key) {
			b, _ := base64.RawStdEncoding.DecodeString(key)
			opts.key = b
		}
	}
}
