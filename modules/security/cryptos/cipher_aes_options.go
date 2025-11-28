package cryptos

import (
	"encoding/base64"

	"github.com/guidomantilla/yarumo/common/utils"
	"github.com/guidomantilla/yarumo/security/keys"
)

type AesCipherOptions struct {
	key []byte
}

func NewAesCipherOptions(opts ...AesCipherOption) *AesCipherOptions {
	options := &AesCipherOptions{
		key: keys.Key(32),
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

type AesCipherOption func(opts *AesCipherOptions)

func WithAesCipherKeySize32() AesCipherOption {
	return func(opts *AesCipherOptions) {
		opts.key = keys.Key(32)
	}
}

func WithAesCipherKey(key string) AesCipherOption {
	return func(opts *AesCipherOptions) {
		if utils.NotEmpty(key) {
			b, _ := base64.RawStdEncoding.DecodeString(key)
			opts.key = b
		}
	}
}
