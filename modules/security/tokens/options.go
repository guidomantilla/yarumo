package tokens

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/guidomantilla/yarumo/common/random"
	"github.com/guidomantilla/yarumo/common/types"
	"github.com/guidomantilla/yarumo/common/utils"
)

type Option func(opts *Options)

type Options struct {
	issuer        string
	timeout       time.Duration
	cipherKey     types.Bytes
	signingKey    types.Bytes
	verifyingKey  types.Bytes
	signingMethod jwt.SigningMethod
}

func NewOptions(opts ...Option) *Options {
	signingKey := random.Key(64) // for HS512
	cipherKey := random.Key(32)  // for AES-256

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

func WithJwtKey(key types.Bytes) Option {
	return func(opts *Options) {
		if utils.NotEmpty(key) {
			opts.signingKey, opts.verifyingKey = key, key
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

func WithOpaqueKey(cipherKey types.Bytes) Option {
	return func(opts *Options) {
		if utils.NotEmpty(cipherKey) {
			opts.cipherKey = cipherKey
		}
	}
}
