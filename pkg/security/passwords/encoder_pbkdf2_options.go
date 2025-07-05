package passwords

import "crypto/sha512"

type Pbkdf2EncoderOptions struct {
	iterations int
	saltLength int
	keyLength  int
	hashFunc   HashFunc
}

func NewPbkdf2EncoderOptions(opts ...Pbkdf2EncoderOption) *Pbkdf2EncoderOptions {
	options := &Pbkdf2EncoderOptions{
		iterations: 600_000,
		saltLength: 32,
		keyLength:  64,
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
		opts.iterations = iterations
	}
}

func WithPbkdf2SaltLength(saltLength int) Pbkdf2EncoderOption {
	return func(opts *Pbkdf2EncoderOptions) {
		opts.saltLength = saltLength
	}
}

func WithPbkdf2KeyLength(keyLength int) Pbkdf2EncoderOption {
	return func(opts *Pbkdf2EncoderOptions) {
		opts.keyLength = keyLength
	}
}

func WithHashFunc(hashFunc HashFunc) Pbkdf2EncoderOption {
	return func(opts *Pbkdf2EncoderOptions) {
		opts.hashFunc = hashFunc
	}
}
