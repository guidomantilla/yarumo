package rest

import (
	"github.com/guidomantilla/yarumo/common/http"
	"github.com/guidomantilla/yarumo/common/utils"
)

type Option func(opts *Options)

type Options struct {
	DoFn http.DoFn
}

func NewOptions(opts ...Option) *Options {
	options := &Options{
		DoFn: http.Do,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

func WithDoFn(doFn http.DoFn) Option {
	return func(opts *Options) {
		if utils.NotNil(doFn) {
			opts.DoFn = doFn
		}
	}
}
