package comm

import (
	"context"
	"crypto/tls"
	"net"
	"time"
)

const (
	HttpTransportMaxRetries            = 3
	HttpTransportMaxIdleConns          = 100
	HttpTransportMaxIdleConnsPerHost   = 10
	HttpTransportIdleConnTimeout       = 90 * time.Second
	HttpTransportDialTimeout           = 5 * time.Second
	HttpTransportKeepAlive             = 30 * time.Second
	HttpTransportTLSHandshakeTimeout   = 5 * time.Second
	HttpTransportResponseHeaderTimeout = 10 * time.Second
	HttpTransportExpectContinueTimeout = 1 * time.Second
)

type HttpTransportOptions struct {
	maxRetries            uint
	tlsClientConfig       *tls.Config
	maxIdleConns          int
	maxIdleConnsPerHost   int
	idleConnTimeout       time.Duration
	dialContext           func(ctx context.Context, network string, addr string) (net.Conn, error)
	tlsHandshakeTimeout   time.Duration
	responseHeaderTimeout time.Duration
	expectContinueTimeout time.Duration
}

func NewHttpTransportOptions(opts ...HttpTransportOption) *HttpTransportOptions {
	options := &HttpTransportOptions{
		maxRetries:          HttpTransportMaxRetries,
		tlsClientConfig:     nil,
		maxIdleConns:        HttpTransportMaxIdleConns,
		maxIdleConnsPerHost: HttpTransportMaxIdleConnsPerHost,
		idleConnTimeout:     HttpTransportIdleConnTimeout,
		dialContext: (&net.Dialer{
			Timeout:   HttpTransportDialTimeout,
			KeepAlive: HttpTransportKeepAlive,
		}).DialContext,
		tlsHandshakeTimeout:   HttpTransportTLSHandshakeTimeout,
		responseHeaderTimeout: HttpTransportResponseHeaderTimeout,
		expectContinueTimeout: HttpTransportExpectContinueTimeout,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

type HttpTransportOption func(opts *HttpTransportOptions)

func WithHttpTransportMaxRetries(maxRetries uint) HttpTransportOption {
	return func(opts *HttpTransportOptions) {
		if maxRetries > 0 {
			opts.maxRetries = maxRetries
		}
	}
}

func WithHttpTransportTLSClientConfig(tlsConfig *tls.Config) HttpTransportOption {
	return func(opts *HttpTransportOptions) {
		opts.tlsClientConfig = tlsConfig
	}
}

func WithHttpTransportMaxIdleConns(maxIdleConns int) HttpTransportOption {
	return func(opts *HttpTransportOptions) {
		opts.maxIdleConns = maxIdleConns
	}
}

func WithHttpTransportMaxIdleConnsPerHost(maxIdleConnsPerHost int) HttpTransportOption {
	return func(opts *HttpTransportOptions) {
		opts.maxIdleConnsPerHost = maxIdleConnsPerHost
	}
}

func WithHttpTransportIdleConnTimeout(idleConnTimeout time.Duration) HttpTransportOption {
	return func(opts *HttpTransportOptions) {
		opts.idleConnTimeout = idleConnTimeout
	}
}

func WithHttpTransportDialContext(timeout time.Duration, keepAlive time.Duration) HttpTransportOption {
	return func(opts *HttpTransportOptions) {
		opts.dialContext = (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: keepAlive,
		}).DialContext
	}
}

func WithHttpTransportTLSHandshakeTimeout(tlsHandshakeTimeout time.Duration) HttpTransportOption {
	return func(opts *HttpTransportOptions) {
		opts.tlsHandshakeTimeout = tlsHandshakeTimeout
	}
}

func WithHttpTransportResponseHeaderTimeout(responseHeaderTimeout time.Duration) HttpTransportOption {
	return func(opts *HttpTransportOptions) {
		opts.responseHeaderTimeout = responseHeaderTimeout
	}
}

func WithHttpTransportExpectContinueTimeout(expectContinueTimeout time.Duration) HttpTransportOption {
	return func(opts *HttpTransportOptions) {
		opts.expectContinueTimeout = expectContinueTimeout
	}
}
