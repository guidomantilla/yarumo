package tokens

import (
	"time"

	crandom "github.com/guidomantilla/yarumo/core/crypto/random"
	cpointer "github.com/guidomantilla/yarumo/core/common/pointer"
)

// Option is a functional option for configuring tokens Options.
type Option func(opts *Options)

// Options holds the configuration for a tokens Method.
//
// signingKey and verifyingKey are typed as any so a single Options value
// can carry the byte-slice secret used by HMAC variants alongside the
// *rsa.PrivateKey / *ecdsa.PrivateKey / ed25519.PrivateKey values (and
// matching public keys) required by the asymmetric variants.
type Options struct {
	issuer       string
	timeout      time.Duration
	signingKey   any
	verifyingKey any
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
//
// key is accepted as any so the same option works for HMAC ([]byte) and
// for asymmetric algorithms where a single value carries both roles
// (none of the predefined asymmetric algorithms do — pass WithSigningKey
// and WithVerifyingKey separately in that case). Nil or empty byte
// slices are ignored.
func WithKey(key any) Option {
	return func(opts *Options) {
		if isUsableKey(key) {
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
//
// If crypto/rand fails to deliver the requested bytes, the keys are left at
// their previous value (typically nil). Method.Generate and Method.Validate
// will then surface ErrSigningKeyNil / ErrVerifyingKeyNil at call-time —
// consistent with the no-key path documented on NewOptions.
func WithGeneratedKey() Option {
	return func(opts *Options) {
		// Convert to a plain []byte so golang-jwt/v5 HMAC type-assertion
		// (key.([]byte)) succeeds — types.Bytes is a named type and would
		// fail that assertion.
		raw, err := crandom.Bytes(64)
		if err != nil {
			return
		}

		key := []byte(raw)
		opts.signingKey = key
		opts.verifyingKey = key
	}
}

// WithSigningKey sets the signing key independently.
//
// key is accepted as any so the same option works for HMAC ([]byte
// secrets) and for the asymmetric algorithms whose signing keys are
// *rsa.PrivateKey, *ecdsa.PrivateKey, or ed25519.PrivateKey. Nil or
// empty byte slices are ignored.
func WithSigningKey(key any) Option {
	return func(opts *Options) {
		if isUsableKey(key) {
			opts.signingKey = key
		}
	}
}

// WithVerifyingKey sets the verifying key independently.
//
// key is accepted as any so the same option works for HMAC ([]byte
// secrets) and for the asymmetric algorithms whose verifying keys are
// *rsa.PublicKey, *ecdsa.PublicKey, or ed25519.PublicKey. Nil or empty
// byte slices are ignored.
func WithVerifyingKey(key any) Option {
	return func(opts *Options) {
		if isUsableKey(key) {
			opts.verifyingKey = key
		}
	}
}

// isUsableKey reports whether the value is non-nil and, in the case of
// a []byte, also non-empty. Asymmetric key values (e.g. *rsa.PrivateKey)
// pass through whenever they are non-nil.
func isUsableKey(key any) bool {
	if cpointer.IsNil(key) {
		return false
	}
	b, ok := key.([]byte)
	if ok {
		return len(b) > 0
	}
	return true
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
