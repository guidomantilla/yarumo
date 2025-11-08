package cryptos

import (
	"encoding/base64"

	"github.com/guidomantilla/yarumo/common/utils"
)

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
		if utils.NotEmpty(key) {
			b, _ := base64.StdEncoding.DecodeString(key)
			opts.key = b
		}
	}
}
