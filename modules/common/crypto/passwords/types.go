// Package passwords provides password encoding, verification, upgrade checking and generation
// using argon2, bcrypt, pbkdf2 and scrypt algorithms.
package passwords

import "hash"

var (
	_ EncodeFn        = encode
	_ VerifyFn        = verify
	_ UpgradeNeededFn = upgradeNeeded
	_ GenerateSaltFn  = generateSalt
)

// EncodeFn is the function type for encoding a raw password using a method.
type EncodeFn func(method *Method, rawPassword string) (string, error)

// VerifyFn is the function type for verifying a raw password against an encoded one.
type VerifyFn func(method *Method, encodedPassword string, rawPassword string) (bool, error)

// UpgradeNeededFn is the function type for checking if an encoded password needs re-encoding.
type UpgradeNeededFn func(method *Method, encodedPassword string) (bool, error)

// GenerateSaltFn is the function type for generating a cryptographic salt.
type GenerateSaltFn func(saltSize int) ([]byte, error)

// HashFunc is the type for hash functions used by pbkdf2.
type HashFunc func() hash.Hash
