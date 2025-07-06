package comm

import (
	"time"

	"github.com/guidomantilla/yarumo/pkg/common/utils"
)

type HttpClientOptions struct {
	Timeout   time.Duration
	Transport *HttpTransport
}

func NewHttpClientOptions(opts ...HttpClientOption) *HttpClientOptions {
	options := &HttpClientOptions{
		Timeout:   0,
		Transport: NewHttpTransport(),
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

type HttpClientOption func(opts *HttpClientOptions)

func WithHttpClientTimeout(timeout time.Duration) HttpClientOption {
	return func(opts *HttpClientOptions) {
		opts.Timeout = timeout
	}
}

func WithHttpClientTransport(transport *HttpTransport) HttpClientOption {
	return func(opts *HttpClientOptions) {
		if utils.NotNil(transport) {
			opts.Transport = transport
		}
	}
}
