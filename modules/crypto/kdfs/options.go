package kdfs

// Default parameter values aligned with current industry guidance:
//   - PBKDF2 iterations: OWASP 2024 password-storage cheat sheet recommends
//     600,000 for HMAC-SHA-256 and 210,000 for HMAC-SHA-512. We default to
//     600,000 across HMAC variants for consistency with the passwords/ package.
//   - Scrypt parameters: OWASP 2024 recommends N=2^17 (131072), r=8, p=1.
const (
	Pbkdf2DefaultIterations = 600_000

	ScryptDefaultN = 131072
	ScryptDefaultR = 8
	ScryptDefaultP = 1
)

// Option is a functional option for configuring kdfs Options.
type Option func(opts *Options)

// pbkdf2Config holds PBKDF2-specific parameters.
type pbkdf2Config struct {
	iterations int
}

// scryptConfig holds Scrypt-specific parameters.
type scryptConfig struct {
	n int
	r int
	p int
}

// Options holds the configuration for a kdfs Method.
type Options struct {
	deriveFn     DeriveFn
	pbkdf2Params *pbkdf2Config
	scryptParams *scryptConfig
}

// NewOptions creates a new Options with defaults and applies the given options.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		deriveFn: hkdfDerive,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithDeriveFn sets the derivation function used by the Method.
func WithDeriveFn(deriveFn DeriveFn) Option {
	return func(opts *Options) {
		if deriveFn != nil {
			opts.deriveFn = deriveFn
		}
	}
}

// WithPbkdf2Iterations sets the PBKDF2 iteration count and installs the
// PBKDF2 derive function as the active derive function.
func WithPbkdf2Iterations(iterations int) Option {
	return func(opts *Options) {
		if iterations >= 1 {
			opts.pbkdf2Params = &pbkdf2Config{iterations: iterations}
			opts.deriveFn = pbkdf2Derive
		}
	}
}

// WithScryptParams sets the Scrypt cost parameters and installs the Scrypt
// derive function as the active derive function.
//
// n is the CPU/memory cost (must be a power of 2 greater than 1), r is the
// block size, p is the parallelization parameter.
func WithScryptParams(n, r, p int) Option {
	return func(opts *Options) {
		if n >= 2 && r >= 1 && p >= 1 {
			opts.scryptParams = &scryptConfig{n: n, r: r, p: p}
			opts.deriveFn = scryptDerive
		}
	}
}
