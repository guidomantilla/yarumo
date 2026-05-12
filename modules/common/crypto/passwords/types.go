// Package passwords provides password encoding, verification, upgrade checking and generation
// using argon2, bcrypt, pbkdf2 and scrypt algorithms.
package passwords

import (
	"hash"

	"golang.org/x/crypto/bcrypt"
)

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
