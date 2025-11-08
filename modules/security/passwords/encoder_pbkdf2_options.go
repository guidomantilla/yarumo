package passwords

import (
	"crypto/sha512"

	"github.com/guidomantilla/yarumo/common/utils"
)

const (
	Pbkdf2Iterations = 600_000
	Pbkdf2SaltLength = 32
	Pbkdf2KeyLength  = 64
)

type Pbkdf2EncoderOptions struct {
	iterations int
	saltLength int
	keyLength  int
	hashFunc   HashFunc
}

func NewPbkdf2EncoderOptions(opts ...Pbkdf2EncoderOption) *Pbkdf2EncoderOptions {
	options := &Pbkdf2EncoderOptions{
		iterations: Pbkdf2Iterations,
		saltLength: Pbkdf2SaltLength,
		keyLength:  Pbkdf2KeyLength,
		hashFunc:   sha512.New,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

type Pbkdf2EncoderOption func(opts *Pbkdf2EncoderOptions)

func WithPbkdf2Iterations(iterations int) Pbkdf2EncoderOption {
	return func(opts *Pbkdf2EncoderOptions) {
		if iterations > Pbkdf2Iterations {
			opts.iterations = iterations
		}
	}
}

func WithPbkdf2SaltLength(saltLength int) Pbkdf2EncoderOption {
	return func(opts *Pbkdf2EncoderOptions) {
		if saltLength > Pbkdf2SaltLength {
			opts.saltLength = saltLength
		}
	}
}

func WithPbkdf2KeyLength(keyLength int) Pbkdf2EncoderOption {
	return func(opts *Pbkdf2EncoderOptions) {
		if keyLength > Pbkdf2KeyLength {
			opts.keyLength = keyLength
		}
	}
}

func WithHashFunc(hashFunc HashFunc) Pbkdf2EncoderOption {
	return func(opts *Pbkdf2EncoderOptions) {
		if utils.NotNil(hashFunc) {
			opts.hashFunc = hashFunc
		}
	}
}
