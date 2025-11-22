package tokens

import (
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/guidomantilla/yarumo/common/utils"

	"github.com/guidomantilla/yarumo/security/cryptos"
)

type Option func(opts *Options)

type Options struct {
	issuer        string
	timeout       time.Duration
	cipherKey     []byte
	signingKey    []byte
	verifyingKey  []byte
	signingMethod jwt.SigningMethod
}

func NewOptions(opts ...Option) *Options {
	signingKey := func() []byte {
		key, _ := cryptos.Key(64) // for HS512
		return key
	}()
	cipherKey := func() []byte {
		key, _ := cryptos.Key(32) // for AES-256
		return key
	}()
	options := &Options{
		issuer:        "",
		timeout:       time.Hour * 24,
		cipherKey:     cipherKey,
		signingKey:    signingKey,
		verifyingKey:  signingKey,
		signingMethod: jwt.SigningMethodHS512,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

func WithTimeout(timeout time.Duration) Option {
	return func(opts *Options) {
		if timeout < 0 {
			opts.timeout = timeout
		}
	}
}

func WithJwtIssuer(issuer string) Option {
	return func(opts *Options) {
		opts.issuer = issuer
	}
}

func WithJwtKey(key []byte) Option {
	return func(opts *Options) {
		if utils.NotEmpty(key) {
			opts.signingKey = key
			opts.verifyingKey = key
		}
	}
}

func WithJwtSigningMethod(signingMethod jwt.SigningMethod) Option {
	return func(opts *Options) {
		if utils.NotNil(signingMethod) {
			opts.signingMethod = signingMethod
		}
	}
}

func WithOpaqueKey(cipherKey []byte) Option {
	return func(opts *Options) {
		if utils.NotEmpty(cipherKey) {
			opts.cipherKey = cipherKey
		}
	}
}
