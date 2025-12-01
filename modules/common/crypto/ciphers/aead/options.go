package aead

import (
	"github.com/guidomantilla/yarumo/common/utils"
)

type Option func(opts *Options)

type Options struct {
	keyFn     KeyFn
	encryptFn EncryptFn
	decryptFn DecryptFn
}

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

func WithKeyFn(keyFn KeyFn) Option {
	return func(opts *Options) {
		if utils.NotNil(keyFn) {
			opts.keyFn = keyFn
		}
	}
}

func WithEncryptFn(encryptFn EncryptFn) Option {
	return func(opts *Options) {
		if utils.NotNil(encryptFn) {
			opts.encryptFn = encryptFn
		}
	}
}

func WithDecryptFn(decryptFn DecryptFn) Option {
	return func(opts *Options) {
		if utils.NotNil(decryptFn) {
			opts.decryptFn = decryptFn
		}
	}
}
