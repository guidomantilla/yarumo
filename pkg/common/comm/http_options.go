package comm

import (
	"crypto/tls"
	"time"
)

type HttpOptions struct {
	Timeout         time.Duration
	MaxRetries      uint
	TLSClientConfig *tls.Config
}

func NewHttpOptions(opts ...HttpOption) *HttpOptions {
	options := &HttpOptions{
		Timeout:         0,
		MaxRetries:      3,
		TLSClientConfig: nil,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

type HttpOption func(opts *HttpOptions)

func WithTimeout(timeout time.Duration) HttpOption {
	return func(opts *HttpOptions) {
		opts.Timeout = timeout
	}
}

func WithMaxRetries(maxRetries uint) HttpOption {
	return func(opts *HttpOptions) {
		opts.MaxRetries = maxRetries
	}
}

func WithTLSClientConfig(tlsConfig *tls.Config) HttpOption {
	return func(opts *HttpOptions) {
		opts.TLSClientConfig = tlsConfig
	}
}
