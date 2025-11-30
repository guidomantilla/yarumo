package hmacs

import (
	"crypto"

	"github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/random"
	"github.com/guidomantilla/yarumo/common/types"
)

var (
	HMAC_with_SHA256 = NewMethod("HMAC_with_SHA256", crypto.SHA256, 32)
	HMAC_with_SHA512 = NewMethod("HMAC_with_SHA512", crypto.SHA512, 64)
)

type Method struct {
	name    string
	kind    crypto.Hash
	keySize int
}

func NewMethod(name string, kind crypto.Hash, keySize int) *Method {
	return &Method{
		name:    name,
		kind:    kind,
		keySize: keySize,
	}
}

func (m *Method) Name() string {
	assert.NotNil(m, "method is nil")
	return m.name
}

// GenerateKey generates a new random key for the HMAC algorithm.
func (m *Method) GenerateKey() types.Bytes {
	assert.NotNil(m, "method is nil")
	return random.Key(m.keySize)
}

// Digest computes an HMAC over the given data using the specified hash
// function and key.
//
// Parameters:
//   - hash: the hash function to use (e.g., crypto.SHA256). Must be available.
//   - key: secret key used for HMAC.
//   - data: message to authenticate.
//
// Behavior:
//   - Asserts that the hash function is available.
//   - Initializes an HMAC instance with the given key.
//   - Writes the data into the HMAC.
//   - Returns the final HMAC digest.
//
// Returns:
//   - The HMAC value as a byte slice.
//
// Notes:
//   - The function never returns an error; it silently ignores Write errors.
//   - Panics only if the hash function is not registered (via assert).
func (m *Method) Digest(key types.Bytes, data types.Bytes) types.Bytes {
	assert.NotNil(m, "method is nil")
	return Digest(m.kind, key, data)
}

// Validate verifies an HMAC digest using the specified hash function and key.
//
// Parameters:
//   - hash: the hash function used by the HMAC (e.g., crypto.SHA256). Must be available.
//   - key: secret key used to compute the HMAC.
//   - digest: the expected HMAC value to compare against.
//   - data: message whose authenticity and integrity are being verified.
//
// Behavior:
//   - Asserts that the hash function is available.
//   - Recomputes the HMAC using Digest.
//   - Uses hmac.Equal for constant-time comparison.
//
// Returns:
//   - true if the digest matches the calculated HMAC.
//   - false otherwise.
//
// Notes:
//   - Panics only if the hash function is not registered (via assert).
func (m *Method) Validate(key types.Bytes, digest types.Bytes, data types.Bytes) bool {
	assert.NotNil(m, "method is nil")
	return Validate(m.kind, key, digest, data)
}
