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
// can use WithSecureDefaults (see below).
//
// # WithSecureDefaults — recommended high-security profile
//
// WithSecureDefaults applies yarumo's recommended high-security profile for
// the password algorithm associated with a Method. It is the ceiling
// counterpart to the package-level constants, which act as the floor.
//
// Profile values, reviewed as of 2026-05 (may shift with future yarumo
// releases — pin a specific version if you require deterministic parameters
// across deploys):
//
//   - Argon2id / Argon2i: iterations=3, memory=64 MiB (65536 KiB), threads=2.
//     Salt length 16 bytes, key length 32 bytes (inherits from
//     Argon2SaltLength / Argon2KeyLength).
//   - Bcrypt: cost=12 (matches OWASP 2024 recommendation).
//   - PBKDF2: iterations=1,200,000 with SHA-512. Salt length 32 bytes, key
//     length 64 bytes (inherits from Pbkdf2SaltLength / Pbkdf2KeyLength).
//   - Scrypt: N=2^17 (131072), r=8, p=1. Salt length 16 bytes, key length 32
//     bytes (inherits from ScryptSaltLength / ScryptKeyLength).
//
// WithSecureDefaults dispatches on the Method prefix (set by NewMethod before
// user options run); using it inside NewMethod is the supported entry point.
// Calling it on raw Options via NewOptions is a no-op because the prefix is
// empty in that path. Unknown prefixes are also no-ops — combine with a
// WithXxxParams call for full control of custom Methods.
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
//
// # Config-driven algorithm selection
//
// *Method implements encoding.TextMarshaler / encoding.TextUnmarshaler.
// MarshalText emits the registered algorithm name; UnmarshalText resolves a
// name against the package registry (via Get) and overwrites the receiver.
// This makes Method directly compatible with libraries that honor the
// encoding interfaces — including encoding/json, viper, kong, and koanf —
// so deployments can load encoder choice from YAML/JSON/TOML config.
//
// Caveat: UnmarshalText resolves against whatever the registry contains at
// the time of the call. Custom methods registered via Register after config
// load will not resolve here; callers that need late-bound lookup should
// call Get(name) directly.
//
// # Migration: Argon2 → Argon2id
//
// The predefined Argon2 method was always backed by argon2.IDKey — i.e. the
// argon2id variant, the OWASP-recommended option for password storage. The
// generic "Argon2" name was ambiguous because the Argon2 family has three
// variants (argon2i, argon2d, argon2id). YA-0030 renames the predefined to
// Argon2id with the matching {argon2id} prefix and adds a second predefined,
// Argon2i, that uses argon2.Key (the side-channel-resistant variant) under
// the {argon2i} prefix. argon2d is intentionally NOT added — it is not
// appropriate for password storage.
//
// Backward compatibility:
//
//   - Stored hashes encoded under the legacy {argon2} prefix continue to
//     verify: ByPrefix matches both {argon2} and {argon2id} and routes them
//     to the Argon2id method. Newly encoded passwords use {argon2id}.
//   - The Go-level identifier passwords.Argon2 remains as a deprecated alias
//     of passwords.Argon2id for one release; direct consumers will see a
//     staticcheck SA1019 warning prompting migration. The deprecated alias
//     is not separately registered in the algorithm map, so Get("Argon2")
//     returns ErrAlgorithmNotSupported — callers using registry lookup must
//     migrate to Get("Argon2id"). Likewise the Argon2PrefixKey constant is
//     deprecated in favor of Argon2idPrefixKey.
//
// The {argon2i} prefix has no legacy alias (the variant did not exist before
// this change). To pick the i-variant explicitly, use passwords.Argon2i or
// Get("Argon2i").
package passwords

import (
	"crypto/sha512"
	"hash"
)

// securePbkdf2HashFunc is the hash constructor applied by WithSecureDefaults
// for pbkdf2. SHA-512 matches the OWASP 2024 recommendation and the package
// default; isolated here so the secure profile is self-contained.
var securePbkdf2HashFunc HashFunc = sha512.New

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
	// useArgon2i selects the side-channel-resistant argon2.Key (argon2i)
	// implementation instead of the default argon2.IDKey (argon2id). It is
	// set by WithArgon2iParams and read inside argon2Encode / argon2Verify
	// to pick the correct KDF call. False means argon2id (the default).
	useArgon2i bool
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

// Secure-profile algorithm parameters applied by WithSecureDefaults. These
// represent yarumo's recommended high-security profile (reviewed 2026-05) and
// may shift with future releases. See WithSecureDefaults for usage and the
// package doc for the authoritative profile listing.
const (
	// SecureArgon2Iterations is the t parameter (time cost) used by
	// WithSecureDefaults for argon2id and argon2i.
	SecureArgon2Iterations = 3
	// SecureArgon2Memory is the m parameter (memory cost in KiB) used by
	// WithSecureDefaults for argon2id and argon2i: 64 MiB.
	SecureArgon2Memory = 64 * 1024
	// SecureArgon2Threads is the p parameter (parallelism) used by
	// WithSecureDefaults for argon2id and argon2i.
	SecureArgon2Threads = 2

	// SecureBcryptCost is the bcrypt cost applied by WithSecureDefaults.
	SecureBcryptCost = 12

	// SecurePbkdf2Iterations is the iteration count applied by
	// WithSecureDefaults for pbkdf2 (SHA-512).
	SecurePbkdf2Iterations = 1_200_000

	// SecureScryptN is the CPU/memory cost parameter applied by
	// WithSecureDefaults for scrypt: 2^17 = 131072.
	SecureScryptN = 131072
	// SecureScryptR is the block size parameter applied by WithSecureDefaults
	// for scrypt.
	SecureScryptR = 8
	// SecureScryptP is the parallelism parameter applied by WithSecureDefaults
	// for scrypt.
	SecureScryptP = 1
)
