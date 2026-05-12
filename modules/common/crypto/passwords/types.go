// Package passwords provides password encoding, verification, upgrade checking and generation
// using argon2, bcrypt, pbkdf2 and scrypt algorithms.
//
// # OWASP 2024 defaults
//
// Defaults track the OWASP Password Storage Cheat Sheet (2024 revision):
//   - Argon2id: t=1, m=64 MiB, p=2 (RFC 9106 second recommended profile).
//   - Bcrypt: cost 12 (above OWASP minimum of 10; matches OWASP recommendation).
//   - Pbkdf2: 600,000 iterations with SHA-512.
//   - Scrypt: N=2^17 (131072), r=8, p=1.
//
// # Backward compatibility of stored hashes
//
// Bumping the package-level defaults only affects newly encoded passwords.
// Pre-existing hashes encoded under weaker parameters (e.g. ScryptN=2^15 or
// BcryptDefaultCost=10) continue to verify after an upgrade: both bcrypt and
// scrypt embed their parameters into the encoded hash string, and Method.Verify
// reads the stored parameters when re-deriving the key — it never substitutes
// the current package defaults. Argon2 and Pbkdf2 likewise embed their full
// parameter set. Callers can detect stale parameters with Method.UpgradeNeeded
// and re-encode on next successful login.
//
// Callers wanting to pin a profile by name (e.g. "owasp-2024") rather than
// inherit whatever defaults the imported version of this package carries
// should track YA-0034 (WithSecureDefaults helper).
//
// # DelegatingEncoder and gradual algorithm migration
//
// DelegatingEncoder is a Spring-Security-style wrapper that encodes new
// passwords with a configurable primary Method but routes Verify and
// UpgradeNeeded calls via the package ByPrefix registry. This enables the
// canonical "login-time upgrade" pattern when an application needs to
// migrate stored hashes from a legacy algorithm to a new one (for example,
// bcrypt → argon2id): legacy hashes continue to verify, UpgradeNeeded
// returns true whenever the encoded prefix resolves to a method other than
// the primary, and the caller re-encodes the password with the primary on
// next successful login. See NewDelegatingEncoder for the constructor.
//
// # Random bytes for non-password use cases
//
// This package does not expose a public salt generator. Callers that need
// cryptographically-secure random bytes for adjacent purposes — a non-password
// KDF, a token nonce, a session id, etc. — should use
// [github.com/guidomantilla/yarumo/common/random.Bytes] directly. The
// passwords package itself sources salt entropy from that same primitive, so
// there is a single source of truth for random-bytes generation in the
// workspace.
package passwords

import (
	"hash"
)

var (
	_ EncodeFn        = encode
	_ VerifyFn        = verify
	_ UpgradeNeededFn = upgradeNeeded
)

// EncodeFn is the function type for encoding a raw password using a method.
type EncodeFn func(method *Method, rawPassword string) (string, error)

// VerifyFn is the function type for verifying a raw password against an encoded one.
type VerifyFn func(method *Method, encodedPassword string, rawPassword string) (bool, error)

// UpgradeNeededFn is the function type for checking if an encoded password needs re-encoding.
type UpgradeNeededFn func(method *Method, encodedPassword string) (bool, error)

// HashFunc is the type for hash functions used by pbkdf2.
type HashFunc func() hash.Hash

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

	// BcryptDefaultCost is the default bcrypt cost. OWASP 2024 minimum is 10 and
	// recommended is 12; we ship 12 as the default. Must be < bcrypt.MaxCost (31).
	BcryptDefaultCost = 12

	Pbkdf2Iterations = 600_000
	Pbkdf2SaltLength = 32
	Pbkdf2KeyLength  = 64

	// ScryptN is the default scrypt CPU/memory cost parameter (work factor).
	// OWASP 2024 recommends 2^17; we ship 2^17 = 131072 as the default.
	ScryptN          = 131072
	ScryptR          = 8
	ScryptP          = 1
	ScryptSaltLength = 16
	ScryptKeyLength  = 32
)
