package hashes

import "github.com/guidomantilla/yarumo/common/utils"

type Option func(opts *Options)

type Options struct {
	hashFn HashFn
}

func NewOptions(opts ...Option) *Options {
	options := &Options{
		hashFn: Hash,
	}

	for _, opt := range opts {
		opt(options)
	}
	return options
}

func WithHashFn(hashFn HashFn) Option {
	return func(opts *Options) {
		if utils.NotNil(hashFn) {
			opts.hashFn = hashFn
		}
	}
}
