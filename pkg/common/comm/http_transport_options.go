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
	MaxRetries            uint
	TLSClientConfig       *tls.Config
	MaxIdleConns          int
	MaxIdleConnsPerHost   int
	IdleConnTimeout       time.Duration
	DialContext           func(ctx context.Context, network string, addr string) (net.Conn, error)
	TLSHandshakeTimeout   time.Duration
	ResponseHeaderTimeout time.Duration
	ExpectContinueTimeout time.Duration
}

func NewHttpTransportOptions(opts ...HttpTransportOption) *HttpTransportOptions {
	options := &HttpTransportOptions{
		MaxRetries:          HttpTransportMaxRetries,
		TLSClientConfig:     nil,
		MaxIdleConns:        HttpTransportMaxIdleConns,
		MaxIdleConnsPerHost: HttpTransportMaxIdleConnsPerHost,
		IdleConnTimeout:     HttpTransportIdleConnTimeout,
		DialContext: (&net.Dialer{
			Timeout:   HttpTransportDialTimeout,
			KeepAlive: HttpTransportKeepAlive,
		}).DialContext,
		TLSHandshakeTimeout:   HttpTransportTLSHandshakeTimeout,
		ResponseHeaderTimeout: HttpTransportResponseHeaderTimeout,
		ExpectContinueTimeout: HttpTransportExpectContinueTimeout,
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
			opts.MaxRetries = maxRetries
		}
	}
}

func WithHttpTransportTLSClientConfig(tlsConfig *tls.Config) HttpTransportOption {
	return func(opts *HttpTransportOptions) {
		opts.TLSClientConfig = tlsConfig
	}
}

func WithHttpTransportMaxIdleConns(maxIdleConns int) HttpTransportOption {
	return func(opts *HttpTransportOptions) {
		opts.MaxIdleConns = maxIdleConns
	}
}

func WithHttpTransportMaxIdleConnsPerHost(maxIdleConnsPerHost int) HttpTransportOption {
	return func(opts *HttpTransportOptions) {
		opts.MaxIdleConnsPerHost = maxIdleConnsPerHost
	}
}

func WithHttpTransportIdleConnTimeout(idleConnTimeout time.Duration) HttpTransportOption {
	return func(opts *HttpTransportOptions) {
		opts.IdleConnTimeout = idleConnTimeout
	}
}

func WithHttpTransportDialContext(timeout time.Duration, keepAlive time.Duration) HttpTransportOption {
	return func(opts *HttpTransportOptions) {
		opts.DialContext = (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: keepAlive,
		}).DialContext
	}
}

func WithHttpTransportTLSHandshakeTimeout(tlsHandshakeTimeout time.Duration) HttpTransportOption {
	return func(opts *HttpTransportOptions) {
		opts.TLSHandshakeTimeout = tlsHandshakeTimeout
	}
}

func WithHttpTransportResponseHeaderTimeout(responseHeaderTimeout time.Duration) HttpTransportOption {
	return func(opts *HttpTransportOptions) {
		opts.ResponseHeaderTimeout = responseHeaderTimeout
	}
}

func WithHttpTransportExpectContinueTimeout(expectContinueTimeout time.Duration) HttpTransportOption {
	return func(opts *HttpTransportOptions) {
		opts.ExpectContinueTimeout = expectContinueTimeout
	}
}
