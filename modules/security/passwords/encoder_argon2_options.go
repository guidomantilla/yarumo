package passwords

const (
	Argon2Iterations = 1
	Argon2Memory     = 64 * 1024
	Argon2Threads    = 2
	Argon2SaltLength = 16
	Argon2KeyLength  = 32
)

type Argon2EncoderOptions struct {
	iterations int
	memory     int
	threads    int
	saltLength int
	keyLength  int
}

func NewArgon2EncoderOptions(opts ...Argon2EncoderOption) *Argon2EncoderOptions {
	options := &Argon2EncoderOptions{
		iterations: Argon2Iterations,
		memory:     Argon2Memory,
		threads:    Argon2Threads,
		saltLength: Argon2SaltLength,
		keyLength:  Argon2KeyLength,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

type Argon2EncoderOption func(opts *Argon2EncoderOptions)

func WithArgon2Iterations(iterations int) Argon2EncoderOption {
	return func(opts *Argon2EncoderOptions) {
		if iterations > Argon2Iterations {
			opts.iterations = iterations
		}
	}
}

func WithArgon2Memory(memory int) Argon2EncoderOption {
	return func(opts *Argon2EncoderOptions) {
		if memory > Argon2Memory {
			opts.memory = memory
		}
	}
}

func WithArgon2Threads(threads int) Argon2EncoderOption {
	return func(opts *Argon2EncoderOptions) {
		if threads > Argon2Threads {
			opts.threads = threads
		}
	}
}

func WithArgon2SaltLength(saltLength int) Argon2EncoderOption {
	return func(opts *Argon2EncoderOptions) {
		if saltLength > Argon2SaltLength {
			opts.saltLength = saltLength
		}
	}
}

func WithArgon2KeyLength(keyLength int) Argon2EncoderOption {
	return func(opts *Argon2EncoderOptions) {
		if keyLength > Argon2KeyLength {
			opts.keyLength = keyLength
		}
	}
}
