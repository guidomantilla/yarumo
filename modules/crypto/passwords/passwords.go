package passwords

import (
	"crypto/sha512"

	cassert "github.com/guidomantilla/yarumo/common/assert"
)

// Prefix constants for encoded password identification.
const (
	// Argon2PrefixKey is the legacy prefix produced by the original Argon2
	// predefined method.
	//
	// Deprecated: use Argon2idPrefixKey instead. The underlying implementation
	// was always argon2.IDKey (argon2id) so the generic "{argon2}" prefix was
	// ambiguous. Stored hashes carrying this prefix continue to verify because
	// ByPrefix matches both {argon2} and {argon2id} and routes them to the
	// Argon2id method. Newly encoded passwords use {argon2id}. See the
	// package doc "Migration: Argon2 → Argon2id" section for details.
	Argon2PrefixKey   = "{argon2}"
	Argon2idPrefixKey = "{argon2id}"
	Argon2iPrefixKey  = "{argon2i}"
	BcryptPrefixKey   = "{bcrypt}"
	Pbkdf2PrefixKey   = "{pbkdf2}"
	ScryptPrefixKey   = "{scrypt}"
)

// Predefined methods with default parameters.
var (
	// Argon2id is the OWASP-recommended argon2 variant for password storage,
	// implemented via argon2.IDKey. New encodes use the {argon2id} prefix;
	// ByPrefix also routes legacy {argon2} hashes to this method so existing
	// stored hashes continue to verify after the YA-0030 rename.
	Argon2id = NewMethod("Argon2id", Argon2idPrefixKey, WithArgon2Params(Argon2Iterations, Argon2Memory, Argon2Threads, Argon2SaltLength, Argon2KeyLength))
	// Argon2i is the side-channel-resistant argon2 variant, implemented via
	// argon2.Key. Useful in threat models that include cache-timing or
	// side-channel adversaries; for general-purpose password storage prefer
	// Argon2id, which is the OWASP recommendation.
	Argon2i = NewMethod("Argon2i", Argon2iPrefixKey, WithArgon2iParams(Argon2Iterations, Argon2Memory, Argon2Threads, Argon2SaltLength, Argon2KeyLength))
	// Argon2 is a deprecated alias of Argon2id retained for one release to
	// ease migration. Direct consumers will see staticcheck SA1019.
	//
	// Deprecated: use Argon2id. The original Argon2 predefined was always
	// backed by argon2.IDKey, so the name was ambiguous. Note: the alias is
	// not separately registered in the algorithm map — Get("Argon2") returns
	// ErrAlgorithmNotSupported; callers using registry lookup must migrate
	// to Get("Argon2id"). Stored hashes with the legacy {argon2} prefix
	// continue to verify via ByPrefix's dual-match. See the package doc
	// "Migration: Argon2 → Argon2id" section for details.
	Argon2 = Argon2id
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

	opts := newOptionsForMethod(prefix, options...)

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
