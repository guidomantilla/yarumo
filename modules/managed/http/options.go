package http

import (
	"context"
	stdlog "log"
	"net"
	"time"
)

// Option is a functional option for configuring HTTP server Options.
type Option func(opts *Options)

// Options holds the configuration for creating an HTTP server.
type Options struct {
	readTimeout       time.Duration
	writeTimeout      time.Duration
	idleTimeout       time.Duration
	readHeaderTimeout time.Duration
	maxHeaderBytes    int
	errorLog          *stdlog.Logger
	baseContext       func(net.Listener) context.Context

	tlsEnabled  bool
	tlsCertFile string
	tlsKeyFile  string
}

// NewOptions creates a new Options applying all provided Option functions.
func NewOptions(opts ...Option) *Options {
	o := &Options{}

	for _, opt := range opts {
		opt(o)
	}

	return o
}

// WithReadTimeout returns an Option that sets the maximum duration for
// reading the entire request including the body.
func WithReadTimeout(d time.Duration) Option {
	return func(opts *Options) {
		opts.readTimeout = d
	}
}

// WithWriteTimeout returns an Option that sets the maximum duration before
// timing out writes of the response.
func WithWriteTimeout(d time.Duration) Option {
	return func(opts *Options) {
		opts.writeTimeout = d
	}
}

// WithIdleTimeout returns an Option that sets the maximum amount of time
// to wait for the next request when keep-alives are enabled.
func WithIdleTimeout(d time.Duration) Option {
	return func(opts *Options) {
		opts.idleTimeout = d
	}
}

// WithReadHeaderTimeout returns an Option that sets the amount of time
// allowed to read request headers.
func WithReadHeaderTimeout(d time.Duration) Option {
	return func(opts *Options) {
		opts.readHeaderTimeout = d
	}
}

// WithMaxHeaderBytes returns an Option that sets the maximum number of
// bytes the server reads when parsing request headers.
func WithMaxHeaderBytes(n int) Option {
	return func(opts *Options) {
		opts.maxHeaderBytes = n
	}
}

// WithErrorLog returns an Option that sets a logger for errors accepting
// connections, unexpected handler behavior, and underlying file system
// errors. If nil, logging is done via the http package's standard logger.
func WithErrorLog(l *stdlog.Logger) Option {
	return func(opts *Options) {
		opts.errorLog = l
	}
}

// WithBaseContext returns an Option that sets a function which returns the
// base context for incoming requests on the given listener.
func WithBaseContext(fn func(net.Listener) context.Context) Option {
	return func(opts *Options) {
		opts.baseContext = fn
	}
}

// WithTLS returns an Option that enables TLS using the certificate and key
// files. When set, Start will call ServeTLS instead of Serve.
func WithTLS(certFile string, keyFile string) Option {
	return func(opts *Options) {
		if certFile == "" || keyFile == "" {
			return
		}

		opts.tlsEnabled = true
		opts.tlsCertFile = certFile
		opts.tlsKeyFile = keyFile
	}
}
