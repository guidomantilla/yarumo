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
//
// The signing and verifying keys default to nil. Callers must either supply a
// key explicitly via WithKey / WithSigningKey / WithVerifyingKey, or opt in to
// random-key generation by passing WithGeneratedKey. If neither path is taken,
// Method.Generate and Method.Validate will return ErrSigningKeyNil /
// ErrVerifyingKeyNil at call-time. This avoids burning entropy at construction
// and prevents init-time panics when crypto/rand is constrained.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		timeout:    24 * time.Hour,
		generateFn: generate,
		validateFn: validate,
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

// WithGeneratedKey draws 64 bytes from crypto/rand at apply-time and assigns
// them to both the signing and verifying keys. Use this when the caller does
// not have a key of its own and wants the package to mint a fresh symmetric
// secret. The entropy draw happens when this option is applied (inside
// NewOptions / NewMethod), not at package init, so unused constructors do not
// consume entropy.
func WithGeneratedKey() Option {
	return func(opts *Options) {
		key := crandom.Bytes(64)
		opts.signingKey = key
		opts.verifyingKey = key
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
