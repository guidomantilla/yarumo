package tokens

import (
	"encoding/base64"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"

	"github.com/guidomantilla/yarumo/pkg/security/cryptos"
)

type JwtGeneratorOptions struct {
	issuer        string
	timeout       time.Duration
	signingKey    []byte
	verifyingKey  []byte
	signingMethod jwt.SigningMethod
}

func NewJwtGeneratorOptions(opts ...JwtGeneratorOption) *JwtGeneratorOptions {
	key := func() []byte {
		key, _ := cryptos.Key(64)
		b, _ := base64.StdEncoding.DecodeString(*key)
		return b
	}()
	options := &JwtGeneratorOptions{
		issuer:        "",
		timeout:       time.Hour * 24,
		signingKey:    key,
		verifyingKey:  key,
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

func WithJwtSigningKey(signingKey []byte) JwtGeneratorOption {
	return func(opts *JwtGeneratorOptions) {
		opts.signingKey = signingKey
	}
}

func WithJwtVerifyingKey(verifyingKey []byte) JwtGeneratorOption {
	return func(opts *JwtGeneratorOptions) {
		opts.verifyingKey = verifyingKey
	}
}

func WithJwtSigningMethod(signingMethod jwt.SigningMethod) JwtGeneratorOption {
	return func(opts *JwtGeneratorOptions) {
		opts.signingMethod = signingMethod
	}
}
