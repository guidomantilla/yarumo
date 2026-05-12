package hashes

import (
	"crypto"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	ctypes "github.com/guidomantilla/yarumo/common/types"
)

// Predefined hash methods registered at package init.
var (
	// SHA1 is the legacy SHA-1 hash (160-bit digest).
	//
	// Deprecated: SHA-1 is cryptographically broken for collision resistance
	// (SHAttered, 2017) and MUST NOT be used for new code that relies on
	// collision or pre-image resistance — including digital signatures,
	// certificate fingerprints, content addressing, or password derivation.
	// Prefer SHA-256 or stronger (SHA-384, SHA-512, SHA3-*, BLAKE2b) for
	// new applications. This method is provided solely for interoperability
	// with legacy protocols and formats that still mandate SHA-1, such as
	// TLS 1.0/1.1 handshakes, Git object identifiers, HMAC-SHA1 (still
	// considered safe as a MAC), older PGP signatures, and other long-lived
	// wire formats. Callers using SHA-1 outside of those constraints should
	// migrate to a SHA-2 or SHA-3 variant.
	SHA1 = NewMethod("SHA1", crypto.SHA1)
	// SHA224 is the SHA-2 224-bit hash. Prefer SHA-256 unless a fixed
	// 224-bit output width is required (e.g. NIST suite-B interop).
	SHA224      = NewMethod("SHA224", crypto.SHA224)
	SHA256      = NewMethod("SHA256", crypto.SHA256)
	SHA384      = NewMethod("SHA384", crypto.SHA384)
	SHA512      = NewMethod("SHA512", crypto.SHA512)
	SHA3_256    = NewMethod("SHA3_256", crypto.SHA3_256)
	SHA3_384    = NewMethod("SHA3_384", crypto.SHA3_384)
	SHA3_512    = NewMethod("SHA3_512", crypto.SHA3_512)
	BLAKE2b_256 = NewMethod("BLAKE2b_256", crypto.BLAKE2b_256)
	BLAKE2b_512 = NewMethod("BLAKE2b_512", crypto.BLAKE2b_512)
)

// Method holds the configuration for a hash algorithm.
type Method struct {
	name   string
	kind   crypto.Hash
	hashFn HashFn
}

// NewMethod creates a new Method with the given name and hash algorithm.
func NewMethod(name string, kind crypto.Hash, options ...Option) *Method {
	cassert.NotEmpty(name, "name is empty")

	opts := NewOptions(options...)

	return &Method{
		name:   name,
		kind:   kind,
		hashFn: opts.hashFn,
	}
}

// Name returns the method's algorithm name.
func (m *Method) Name() string {
	cassert.NotNil(m, "method is nil")

	return m.name
}

// Hash computes the digest of the given data using this method's hash function.
// It returns an error if the underlying crypto.Hash driver is unavailable.
func (m *Method) Hash(data ctypes.Bytes) (ctypes.Bytes, error) {
	cassert.NotNil(m, "method is nil")
	cassert.NotNil(m.hashFn, "method hashFn is nil")

	digest, err := m.hashFn(m.kind, data)
	if err != nil {
		return nil, ErrDigest(err)
	}

	return digest, nil
}
