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
	// prefix is the method prefix used to identify the algorithm family.
	// NewMethod pre-populates this before applying user options, so
	// prefix-driven options like WithSecureDefaults can dispatch to the
	// correct WithXxxParams setter. Empty when Options is constructed
	// directly via NewOptions; algorithm-specific options that rely on
	// prefix become no-ops in that case.
	prefix string
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

// newOptionsForMethod creates Options with the given prefix pre-populated and
// then applies user options. It is the internal entry point used by NewMethod
// so prefix-driven options (e.g. WithSecureDefaults) can dispatch on the
// algorithm family.
func newOptionsForMethod(prefix string, opts ...Option) *Options {
	options := &Options{
		encodeFn:        encode,
		verifyFn:        verify,
		upgradeNeededFn: upgradeNeeded,
		prefix:          prefix,
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

// WithArgon2Params sets the argon2id algorithm parameters. The resulting
// Method uses argon2.IDKey (the OWASP-recommended variant) at encode and
// verify time.
func WithArgon2Params(iterations, memory, threads, saltLength, keyLength int) Option {
	return func(opts *Options) {
		if iterations >= Argon2Iterations && memory >= Argon2Memory && threads >= Argon2Threads && saltLength >= Argon2SaltLength && keyLength >= Argon2KeyLength {
			opts.argon2Params = &argon2Config{
				iterations: iterations,
				memory:     memory,
				threads:    threads,
				saltLength: saltLength,
				keyLength:  keyLength,
				useArgon2i: false,
			}
		}
	}
}

// WithArgon2iParams sets the argon2i algorithm parameters. The resulting
// Method uses argon2.Key (the side-channel-resistant variant) at encode and
// verify time. For general-purpose password storage prefer WithArgon2Params
// (argon2id), the OWASP-recommended option.
func WithArgon2iParams(iterations, memory, threads, saltLength, keyLength int) Option {
	return func(opts *Options) {
		if iterations >= Argon2Iterations && memory >= Argon2Memory && threads >= Argon2Threads && saltLength >= Argon2SaltLength && keyLength >= Argon2KeyLength {
			opts.argon2Params = &argon2Config{
				iterations: iterations,
				memory:     memory,
				threads:    threads,
				saltLength: saltLength,
				keyLength:  keyLength,
				useArgon2i: true,
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

// WithSecureDefaults applies yarumo's recommended high-security profile for
// the password algorithm associated with this Method. The profile is dispatched
// off the Method prefix (set by NewMethod before user options run), so
// WithSecureDefaults must be used inside NewMethod — calling it on raw Options
// via NewOptions is a no-op because the prefix is empty.
//
// Profile values (as of 2026-05, see package doc for the authoritative list):
//   - Argon2id / Argon2i: iterations=3, memory=64 MiB (65536), threads=2.
//   - Bcrypt: cost=12.
//   - PBKDF2: iterations=1,200,000 with SHA-512.
//   - Scrypt: N=2^17 (131072), r=8, p=1.
//
// Salt and key lengths fall back to the package constants (Argon2SaltLength,
// Argon2KeyLength, Pbkdf2SaltLength, Pbkdf2KeyLength, ScryptSaltLength,
// ScryptKeyLength). These are floors, not ceilings — callers that want larger
// salts or keys can stack a WithXxxParams call after WithSecureDefaults.
//
// Unknown prefixes are no-ops: WithSecureDefaults never panics on a custom
// Method that does not match one of the four known families. Combine with a
// WithXxxParams call for full control in that case.
//
// Profile values may shift with future yarumo releases as upstream guidance
// evolves; pin a specific yarumo version if you require deterministic
// parameters across deploys.
func WithSecureDefaults() Option {
	return func(opts *Options) {
		switch opts.prefix {
		case Argon2idPrefixKey, Argon2PrefixKey:
			opts.argon2Params = &argon2Config{
				iterations: SecureArgon2Iterations,
				memory:     SecureArgon2Memory,
				threads:    SecureArgon2Threads,
				saltLength: Argon2SaltLength,
				keyLength:  Argon2KeyLength,
				useArgon2i: false,
			}
		case Argon2iPrefixKey:
			opts.argon2Params = &argon2Config{
				iterations: SecureArgon2Iterations,
				memory:     SecureArgon2Memory,
				threads:    SecureArgon2Threads,
				saltLength: Argon2SaltLength,
				keyLength:  Argon2KeyLength,
				useArgon2i: true,
			}
		case BcryptPrefixKey:
			opts.bcryptParams = &bcryptConfig{
				cost: SecureBcryptCost,
			}
		case Pbkdf2PrefixKey:
			opts.pbkdf2Params = &pbkdf2Config{
				iterations: SecurePbkdf2Iterations,
				saltLength: Pbkdf2SaltLength,
				keyLength:  Pbkdf2KeyLength,
				hashFunc:   securePbkdf2HashFunc,
			}
		case ScryptPrefixKey:
			opts.scryptParams = &scryptConfig{
				n:          SecureScryptN,
				r:          SecureScryptR,
				p:          SecureScryptP,
				saltLength: ScryptSaltLength,
				keyLength:  ScryptKeyLength,
			}
		}
	}
}
