package rsapss

import (
	"github.com/guidomantilla/yarumo/common/utils"
)

type Option func(opts *Options)

type Options struct {
	keyFn    KeyFn
	signFn   SignFn
	verifyFn VerifyFn
}

func NewOptions(opts ...Option) *Options {
	options := &Options{
		keyFn:    key,
		signFn:   sign,
		verifyFn: verify,
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

func WithSignFn(signFn SignFn) Option {
	return func(opts *Options) {
		if utils.NotNil(signFn) {
			opts.signFn = signFn
		}
	}
}

func WithVerifyFn(verifyFn VerifyFn) Option {
	return func(opts *Options) {
		if utils.NotNil(verifyFn) {
			opts.verifyFn = verifyFn
		}
	}
}
