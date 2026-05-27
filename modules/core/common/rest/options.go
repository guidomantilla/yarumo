package rest

import (
	chttp "github.com/guidomantilla/yarumo/core/common/http"
)

// Option is a functional option for configuring rest Options.
type Option func(opts *Options)

const defaultMaxResponseSize int64 = 10 << 20 // 10 MB

// Options holds the configuration for a REST call.
type Options struct {
	client          chttp.Client
	maxResponseSize int64
}

// NewOptions creates Options with defaults and applies the provided options.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		client:          chttp.NewClient(),
		maxResponseSize: defaultMaxResponseSize,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithClient sets the HTTP client used to execute REST requests.
func WithClient(client chttp.Client) Option {
	return func(opts *Options) {
		if client != nil {
			opts.client = client
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
