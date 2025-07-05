package passwords

type Argon2EncoderOptions struct {
	iterations int
	memory     int
	threads    int
	saltLength int
	keyLength  int
}

func NewArgon2EncoderOptions(opts ...Argon2EncoderOption) *Argon2EncoderOptions {
	options := &Argon2EncoderOptions{
		iterations: 1,
		memory:     64 * 1024,
		threads:    2,
		saltLength: 16,
		keyLength:  32,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

type Argon2EncoderOption func(opts *Argon2EncoderOptions)

func WithArgon2Iterations(iterations int) Argon2EncoderOption {
	return func(opts *Argon2EncoderOptions) {
		opts.iterations = iterations
	}
}

func WithArgon2Memory(memory int) Argon2EncoderOption {
	return func(opts *Argon2EncoderOptions) {
		opts.memory = memory
	}
}

func WithArgon2Threads(threads int) Argon2EncoderOption {
	return func(opts *Argon2EncoderOptions) {
		opts.threads = threads
	}
}

func WithArgon2SaltLength(saltLength int) Argon2EncoderOption {
	return func(opts *Argon2EncoderOptions) {
		opts.saltLength = saltLength
	}
}

func WithArgon2KeyLength(keyLength int) Argon2EncoderOption {
	return func(opts *Argon2EncoderOptions) {
		opts.keyLength = keyLength
	}
}
