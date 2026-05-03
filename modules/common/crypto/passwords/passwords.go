package passwords

import (
	"crypto/sha512"

	cassert "github.com/guidomantilla/yarumo/common/assert"
)

// Prefix constants for encoded password identification.
const (
	Argon2PrefixKey = "{argon2}"
	BcryptPrefixKey = "{bcrypt}"
	Pbkdf2PrefixKey = "{pbkdf2}"
	ScryptPrefixKey = "{scrypt}"
)

// Predefined methods with default parameters.
var (
	Argon2 = NewMethod("Argon2", Argon2PrefixKey, WithArgon2Params(Argon2Iterations, Argon2Memory, Argon2Threads, Argon2SaltLength, Argon2KeyLength))
	Bcrypt = NewMethod("Bcrypt", BcryptPrefixKey, WithBcryptParams(BcryptDefaultCost))
	Pbkdf2 = NewMethod("Pbkdf2", Pbkdf2PrefixKey, WithPbkdf2Params(Pbkdf2Iterations, Pbkdf2SaltLength, Pbkdf2KeyLength, sha512.New))
	Scrypt = NewMethod("Scrypt", ScryptPrefixKey, WithScryptParams(ScryptN, ScryptR, ScryptP, ScryptSaltLength, ScryptKeyLength))
)

// Method represents a password encoding algorithm with its configuration.
type Method struct {
	name            string
	prefix          string
	argon2Params    *argon2Config
	bcryptParams    *bcryptConfig
	pbkdf2Params    *pbkdf2Config
	scryptParams    *scryptConfig
	encodeFn        EncodeFn
	verifyFn        VerifyFn
	upgradeNeededFn UpgradeNeededFn
}

// NewMethod creates a new password method with the given name, prefix and options.
func NewMethod(name string, prefix string, options ...Option) *Method {
	cassert.NotEmpty(name, "name is empty")
	cassert.NotEmpty(prefix, "prefix is empty")

	opts := NewOptions(options...)

	return &Method{
		name:            name,
		prefix:          prefix,
		argon2Params:    opts.argon2Params,
		bcryptParams:    opts.bcryptParams,
		pbkdf2Params:    opts.pbkdf2Params,
		scryptParams:    opts.scryptParams,
		encodeFn:        opts.encodeFn,
		verifyFn:        opts.verifyFn,
		upgradeNeededFn: opts.upgradeNeededFn,
	}
}

// Name returns the method name.
func (m *Method) Name() string {
	cassert.NotNil(m, "method is nil")
	return m.name
}

// Encode encodes a raw password using this method.
func (m *Method) Encode(rawPassword string) (string, error) {
	cassert.NotNil(m, "method is nil")
	cassert.NotNil(m.encodeFn, "method encodeFn is nil")

	encoded, err := m.encodeFn(m, rawPassword)
	if err != nil {
		return "", ErrEncoding(err)
	}
	return encoded, nil
}

// Verify checks if a raw password matches an encoded password.
func (m *Method) Verify(encodedPassword string, rawPassword string) (bool, error) {
	cassert.NotNil(m, "method is nil")
	cassert.NotNil(m.verifyFn, "method verifyFn is nil")

	ok, err := m.verifyFn(m, encodedPassword, rawPassword)
	if err != nil {
		return false, ErrVerification(err)
	}
	return ok, nil
}

// UpgradeNeeded checks if an encoded password should be re-encoded with current parameters.
func (m *Method) UpgradeNeeded(encodedPassword string) (bool, error) {
	cassert.NotNil(m, "method is nil")
	cassert.NotNil(m.upgradeNeededFn, "method upgradeNeededFn is nil")

	needed, err := m.upgradeNeededFn(m, encodedPassword)
	if err != nil {
		return false, ErrUpgradeCheck(err)
	}
	return needed, nil
}
