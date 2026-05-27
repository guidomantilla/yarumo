package kdfs

import (
	"crypto"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	ctypes "github.com/guidomantilla/yarumo/core/common/types"
)

// Predefined KDF methods registered at package init.
var (
	// HKDF_with_SHA256 is the RFC 5869 HKDF over HMAC-SHA-256. Use info to
	// bind a derived key to a context label.
	HKDF_with_SHA256 = NewMethod("HKDF_with_SHA256", crypto.SHA256)
	// HKDF_with_SHA384 is the RFC 5869 HKDF over HMAC-SHA-384.
	HKDF_with_SHA384 = NewMethod("HKDF_with_SHA384", crypto.SHA384)
	// HKDF_with_SHA512 is the RFC 5869 HKDF over HMAC-SHA-512.
	HKDF_with_SHA512 = NewMethod("HKDF_with_SHA512", crypto.SHA512)
	// PBKDF2_with_SHA256 is RFC 8018 PBKDF2 with HMAC-SHA-256 and the OWASP
	// 2024 recommended iteration count (600,000).
	PBKDF2_with_SHA256 = NewMethod("PBKDF2_with_SHA256", crypto.SHA256, WithPbkdf2Iterations(Pbkdf2DefaultIterations))
	// PBKDF2_with_SHA512 is RFC 8018 PBKDF2 with HMAC-SHA-512 and a 600,000
	// iteration count. OWASP's SHA-512 baseline is 210,000; we default to
	// 600,000 for parity with PBKDF2_with_SHA256.
	PBKDF2_with_SHA512 = NewMethod("PBKDF2_with_SHA512", crypto.SHA512, WithPbkdf2Iterations(Pbkdf2DefaultIterations))
	// Scrypt_KDF is RFC 7914 scrypt as a key derivation function. The kind
	// is zero because scrypt does not take a crypto.Hash parameter.
	Scrypt_KDF = NewMethod("Scrypt_KDF", 0, WithScryptParams(ScryptDefaultN, ScryptDefaultR, ScryptDefaultP))
)

// Method holds the configuration for a KDF algorithm. For HKDF and PBKDF2
// kind selects the underlying HMAC hash; for Scrypt kind is zero and ignored.
type Method struct {
	name         string
	kind         crypto.Hash
	deriveFn     DeriveFn
	pbkdf2Params *pbkdf2Config
	scryptParams *scryptConfig
}

// NewMethod creates a new KDF Method with the given name and hash algorithm.
// The kind parameter is meaningful for HKDF and PBKDF2; Scrypt-based methods
// should pass 0.
func NewMethod(name string, kind crypto.Hash, options ...Option) *Method {
	cassert.NotEmpty(name, "name is empty")

	opts := NewOptions(options...)

	return &Method{
		name:         name,
		kind:         kind,
		deriveFn:     opts.deriveFn,
		pbkdf2Params: opts.pbkdf2Params,
		scryptParams: opts.scryptParams,
	}
}

// Name returns the method's algorithm name.
func (m *Method) Name() string {
	cassert.NotNil(m, "method is nil")

	return m.name
}

// Derive expands the given input keying material into a key of the requested
// length. The semantics of secret, salt, info and length depend on the
// underlying algorithm:
//
//   - HKDF: secret = IKM, salt is optional but recommended, info is the
//     context/label binding, length is the requested output size in bytes.
//   - PBKDF2: secret = password, salt is required, info is ignored, length
//     is the requested output size in bytes.
//   - Scrypt: secret = password, salt is required, info is ignored, length
//     is the requested output size in bytes.
func (m *Method) Derive(secret, salt, info ctypes.Bytes, length int) (ctypes.Bytes, error) {
	cassert.NotNil(m, "method is nil")
	cassert.NotNil(m.deriveFn, "method deriveFn is nil")

	out, err := m.deriveFn(m, secret, salt, info, length)
	if err != nil {
		return nil, ErrDerive(err)
	}

	return out, nil
}
