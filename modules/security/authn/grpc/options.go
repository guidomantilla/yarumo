package grpc

import "strings"

// Default settings for the gRPC interceptors. gRPC metadata keys are
// lowercase per HTTP/2 conventions; the scheme follows the same
// "Bearer" convention as the HTTP middleware.
const (
	defaultMetadataKey = "authorization"
	defaultScheme      = "Bearer"
)

// Option is a functional option for configuring gRPC interceptor
// Options.
type Option func(opts *Options)

// Options holds the configuration for the unary and stream
// interceptors.
type Options struct {
	metadataKey string
	scheme      string
}

// NewOptions creates a new Options with sensible defaults and applies
// the given options. The metadata key is normalized to lower-case since
// gRPC stores metadata keys exclusively in lower-case.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		metadataKey: defaultMetadataKey,
		scheme:      defaultScheme,
	}

	for _, opt := range opts {
		opt(options)
	}

	options.metadataKey = strings.ToLower(options.metadataKey)

	return options
}

// WithMetadataKey overrides the gRPC metadata key read by the
// interceptors. Empty values are ignored (the default "authorization"
// is preserved). The key is stored lower-cased.
func WithMetadataKey(key string) Option {
	return func(opts *Options) {
		if key != "" {
			opts.metadataKey = key
		}
	}
}

// WithScheme overrides the credential scheme expected as the first
// whitespace-delimited token of the metadata value. Empty values are
// ignored (the default "Bearer" is preserved). Comparisons are
// case-insensitive.
func WithScheme(scheme string) Option {
	return func(opts *Options) {
		if scheme != "" {
			opts.scheme = scheme
		}
	}
}
