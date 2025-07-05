package tokens

import (
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

type JwtGeneratorOptions struct {
	issuer        string
	timeout       time.Duration
	signingKey    any
	verifyingKey  any
	signingMethod jwt.SigningMethod
}

func NewJwtGeneratorOptions(opts ...JwtGeneratorOption) *JwtGeneratorOptions {
	options := &JwtGeneratorOptions{
		issuer:        "",
		timeout:       time.Hour * 24,
		signingKey:    "a-valid-string-secret-that-is-at-least-512-bits-long-which-is-very-long",
		verifyingKey:  "a-valid-string-secret-that-is-at-least-512-bits-long-which-is-very-long",
		signingMethod: jwt.SigningMethodHS512,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

type JwtGeneratorOption func(opts *JwtGeneratorOptions)

func WithJwtIssuer(issuer string) JwtGeneratorOption {
	return func(opts *JwtGeneratorOptions) {
		opts.issuer = issuer
	}
}

func WithJwtTimeout(timeout time.Duration) JwtGeneratorOption {
	return func(opts *JwtGeneratorOptions) {
		opts.timeout = timeout
	}
}

func WithJwtSigningKey(signingKey any) JwtGeneratorOption {
	return func(opts *JwtGeneratorOptions) {
		opts.signingKey = signingKey
	}
}

func WithJwtVerifyingKey(verifyingKey any) JwtGeneratorOption {
	return func(opts *JwtGeneratorOptions) {
		opts.verifyingKey = verifyingKey
	}
}

func WithJwtSigningMethod(signingMethod jwt.SigningMethod) JwtGeneratorOption {
	return func(opts *JwtGeneratorOptions) {
		opts.signingMethod = signingMethod
	}
}
