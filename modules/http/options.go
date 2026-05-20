package http

import (
	"crypto/tls"
	"time"
)

// Option is a functional option for configuring HTTP Server Options.
type Option func(opts *Options)

// Options holds the configuration for HTTP Server construction.
type Options struct {
	readHeaderTimeout time.Duration
	readTimeout       time.Duration
	writeTimeout      time.Duration
	idleTimeout       time.Duration
	maxHeaderBytes    int
	tlsConfig         *tls.Config
}

// NewOptions creates Options with secure defaults and applies the given functional options.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		readHeaderTimeout: 5 * time.Second,
		readTimeout:       15 * time.Second,
		writeTimeout:      15 * time.Second,
		idleTimeout:       60 * time.Second,
		maxHeaderBytes:    1 << 20, // 1 MiB
		tlsConfig:         nil,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithReadHeaderTimeout sets the maximum duration for reading request headers.
func WithReadHeaderTimeout(readHeaderTimeout time.Duration) Option {
	return func(opts *Options) {
		if readHeaderTimeout > 0 {
			opts.readHeaderTimeout = readHeaderTimeout
		}
	}
}

// WithReadTimeout sets the maximum duration for reading the entire request.
func WithReadTimeout(readTimeout time.Duration) Option {
	return func(opts *Options) {
		if readTimeout > 0 {
			opts.readTimeout = readTimeout
		}
	}
}

// WithWriteTimeout sets the maximum duration for writing the response.
func WithWriteTimeout(writeTimeout time.Duration) Option {
	return func(opts *Options) {
		if writeTimeout > 0 {
			opts.writeTimeout = writeTimeout
		}
	}
}

// WithIdleTimeout sets the maximum duration to wait for the next request on a keep-alive connection.
func WithIdleTimeout(idleTimeout time.Duration) Option {
	return func(opts *Options) {
		if idleTimeout > 0 {
			opts.idleTimeout = idleTimeout
		}
	}
}

// WithMaxHeaderBytes sets the maximum number of bytes the server reads parsing request headers.
func WithMaxHeaderBytes(maxHeaderBytes int) Option {
	return func(opts *Options) {
		if maxHeaderBytes > 0 {
			opts.maxHeaderBytes = maxHeaderBytes
		}
	}
}

// WithTLSConfig sets the TLS configuration for the server.
func WithTLSConfig(tlsConfig *tls.Config) Option {
	return func(opts *Options) {
		if tlsConfig != nil {
			opts.tlsConfig = tlsConfig
		}
	}
}
