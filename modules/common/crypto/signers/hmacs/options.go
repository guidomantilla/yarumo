package hmacs

import (
	"github.com/guidomantilla/yarumo/common/utils"
)

type Option func(opts *Options)

type Options struct {
	keyFn      KeyFn
	digestFn   DigestFn
	validateFn ValidateFn
}

func NewOptions(opts ...Option) *Options {
	options := &Options{
		keyFn:      key,
		digestFn:   digest,
		validateFn: validate,
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

func WithDigestFn(digestFn DigestFn) Option {
	return func(opts *Options) {
		if utils.NotNil(digestFn) {
			opts.digestFn = digestFn
		}
	}
}

func WithValidateFn(validateFn ValidateFn) Option {
	return func(opts *Options) {
		if utils.NotNil(validateFn) {
			opts.validateFn = validateFn
		}
	}
}
