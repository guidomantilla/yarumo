package rest

import (
	chttp "github.com/guidomantilla/yarumo/common/http"
)

// Option is a functional option for configuring rest Options.
type Option func(opts *Options)

const defaultMaxResponseSize int64 = 10 << 20 // 10 MB

// Options holds the configuration for a REST call.
type Options struct {
	doFn            chttp.DoFn
	maxResponseSize int64
}

// NewOptions creates Options with defaults and applies the provided options.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		doFn:            chttp.Do,
		maxResponseSize: defaultMaxResponseSize,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithDoFn sets the HTTP execution function.
func WithDoFn(doFn chttp.DoFn) Option {
	return func(opts *Options) {
		if doFn != nil {
			opts.doFn = doFn
		}
	}
}

// WithMaxResponseSize sets the maximum allowed response body size in bytes.
func WithMaxResponseSize(n int64) Option {
	return func(opts *Options) {
		if n > 0 {
			opts.maxResponseSize = n
		}
	}
}
