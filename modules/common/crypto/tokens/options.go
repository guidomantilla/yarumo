package tokens

import (
	"time"

	jwt "github.com/golang-jwt/jwt/v5"

	crandom "github.com/guidomantilla/yarumo/common/random"
)

// Option is a functional option for configuring tokens Options.
type Option func(opts *Options)

// Options holds the configuration for a tokens Method.
type Options struct {
	issuer       string
	timeout      time.Duration
	signingKey   []byte
	verifyingKey []byte
	generateFn   GenerateFn
	validateFn   ValidateFn
}

// NewOptions creates Options with defaults.
func NewOptions(opts ...Option) *Options {
	key := crandom.Bytes(64)
	options := &Options{
		timeout:      24 * time.Hour,
		signingKey:   key,
		verifyingKey: key,
		generateFn:   generate,
		validateFn:   validate,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithIssuer sets the token issuer claim.
func WithIssuer(issuer string) Option {
	return func(opts *Options) {
		if issuer != "" {
			opts.issuer = issuer
		}
	}
}

// WithTimeout sets the token expiration duration.
func WithTimeout(timeout time.Duration) Option {
	return func(opts *Options) {
		if timeout > 0 {
			opts.timeout = timeout
		}
	}
}

// WithKey sets both signing and verifying keys to the same value.
func WithKey(key []byte) Option {
	return func(opts *Options) {
		if len(key) > 0 {
			opts.signingKey = key
			opts.verifyingKey = key
		}
	}
}

// WithSigningKey sets the signing key independently.
func WithSigningKey(key []byte) Option {
	return func(opts *Options) {
		if len(key) > 0 {
			opts.signingKey = key
		}
	}
}

// WithVerifyingKey sets the verifying key independently.
func WithVerifyingKey(key []byte) Option {
	return func(opts *Options) {
		if len(key) > 0 {
			opts.verifyingKey = key
		}
	}
}

// WithGenerateFn sets a custom generate function.
func WithGenerateFn(fn GenerateFn) Option {
	return func(opts *Options) {
		if fn != nil {
			opts.generateFn = fn
		}
	}
}

// WithValidateFn sets a custom validate function.
func WithValidateFn(fn ValidateFn) Option {
	return func(opts *Options) {
		if fn != nil {
			opts.validateFn = fn
		}
	}
}

// Predefined signing methods for convenience.
var (
	SigningMethodHS256 = jwt.SigningMethodHS256
	SigningMethodHS384 = jwt.SigningMethodHS384
	SigningMethodHS512 = jwt.SigningMethodHS512
)
