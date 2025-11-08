package comm

import (
	"net/http"
	"time"

	"github.com/guidomantilla/yarumo/modules/common/utils"
	resilience "github.com/guidomantilla/yarumo/sandbox/resilience"
)

type HttpClientOptions struct {
	timeout   time.Duration
	transport http.RoundTripper
}

func NewHttpClientOptions(opts ...HttpClientOption) *HttpClientOptions {
	options := &HttpClientOptions{
		timeout:   0,
		transport: NewHttpTransport(resilience.NewRateLimiterRegistry(), resilience.NewCircuitBreakerRegistry()),
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

type HttpClientOption func(opts *HttpClientOptions)

func WithHttpClientTimeout(timeout time.Duration) HttpClientOption {
	return func(opts *HttpClientOptions) {
		opts.timeout = timeout
	}
}

func WithHttpClientTransport(transport *HttpTransport) HttpClientOption {
	return func(opts *HttpClientOptions) {
		if utils.NotNil(transport) {
			opts.transport = transport
		}
	}
}
