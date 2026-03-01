package passwords

import (
	"golang.org/x/crypto/bcrypt"
)

// Option is a functional option for configuring passwords Options.
type Option func(opts *Options)

// Options holds the configuration for a passwords Method.
type Options struct {
	encodeFn        EncodeFn
	verifyFn        VerifyFn
	upgradeNeededFn UpgradeNeededFn
	argon2Params    *argon2Config
	bcryptParams    *bcryptConfig
	pbkdf2Params    *pbkdf2Config
	scryptParams    *scryptConfig
}

// NewOptions creates Options with defaults.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		encodeFn:        encode,
		verifyFn:        verify,
		upgradeNeededFn: upgradeNeeded,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithEncodeFn sets a custom encode function.
func WithEncodeFn(fn EncodeFn) Option {
	return func(opts *Options) {
		if fn != nil {
			opts.encodeFn = fn
		}
	}
}

// WithVerifyFn sets a custom verify function.
func WithVerifyFn(fn VerifyFn) Option {
	return func(opts *Options) {
		if fn != nil {
			opts.verifyFn = fn
		}
	}
}

// WithUpgradeNeededFn sets a custom upgrade needed function.
func WithUpgradeNeededFn(fn UpgradeNeededFn) Option {
	return func(opts *Options) {
		if fn != nil {
			opts.upgradeNeededFn = fn
		}
	}
}

// Algorithm-specific parameter structs.
type argon2Config struct {
	iterations int
	memory     int
	threads    int
	saltLength int
	keyLength  int
}

type bcryptConfig struct {
	cost int
}

type pbkdf2Config struct {
	iterations int
	saltLength int
	keyLength  int
	hashFunc   HashFunc
}

type scryptConfig struct {
	n          int
	r          int
	p          int
	saltLength int
	keyLength  int
}

// Default algorithm parameters.
const (
	Argon2Iterations = 1
	Argon2Memory     = 64 * 1024
	Argon2Threads    = 2
	Argon2SaltLength = 16
	Argon2KeyLength  = 32

	BcryptDefaultCost = bcrypt.DefaultCost

	Pbkdf2Iterations = 600_000
	Pbkdf2SaltLength = 32
	Pbkdf2KeyLength  = 64

	ScryptN          = 32768
	ScryptR          = 8
	ScryptP          = 1
	ScryptSaltLength = 16
	ScryptKeyLength  = 32
)

// WithArgon2Params sets the argon2 algorithm parameters.
func WithArgon2Params(iterations, memory, threads, saltLength, keyLength int) Option {
	return func(opts *Options) {
		if iterations >= Argon2Iterations && memory >= Argon2Memory && threads >= Argon2Threads && saltLength >= Argon2SaltLength && keyLength >= Argon2KeyLength {
			opts.argon2Params = &argon2Config{
				iterations: iterations,
				memory:     memory,
				threads:    threads,
				saltLength: saltLength,
				keyLength:  keyLength,
			}
		}
	}
}

// WithBcryptParams sets the bcrypt algorithm parameters.
func WithBcryptParams(cost int) Option {
	return func(opts *Options) {
		if cost >= bcrypt.DefaultCost && cost <= bcrypt.MaxCost {
			opts.bcryptParams = &bcryptConfig{
				cost: cost,
			}
		}
	}
}

// WithPbkdf2Params sets the pbkdf2 algorithm parameters.
func WithPbkdf2Params(iterations, saltLength, keyLength int, hashFunc HashFunc) Option {
	return func(opts *Options) {
		if iterations >= Pbkdf2Iterations && saltLength >= Pbkdf2SaltLength && keyLength >= Pbkdf2KeyLength && hashFunc != nil {
			opts.pbkdf2Params = &pbkdf2Config{
				iterations: iterations,
				saltLength: saltLength,
				keyLength:  keyLength,
				hashFunc:   hashFunc,
			}
		}
	}
}

// WithScryptParams sets the scrypt algorithm parameters.
func WithScryptParams(n, r, p, saltLength, keyLength int) Option {
	return func(opts *Options) {
		if n >= ScryptN && r >= ScryptR && p >= ScryptP && saltLength >= ScryptSaltLength && keyLength >= ScryptKeyLength {
			opts.scryptParams = &scryptConfig{
				n:          n,
				r:          r,
				p:          p,
				saltLength: saltLength,
				keyLength:  keyLength,
			}
		}
	}
}
