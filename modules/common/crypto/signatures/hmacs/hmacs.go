package hmacs

import (
	"crypto"

	"github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/types"
)

var (
	HMAC_with_SHA256 = NewMethod("HMAC_with_SHA256", crypto.SHA256, 32)
	HMAC_with_SHA512 = NewMethod("HMAC_with_SHA512", crypto.SHA512, 64)
)

type Method struct {
	name       string
	kind       crypto.Hash
	keySize    int
	keyFn      KeyFn
	digestFn   DigestFn
	validateFn ValidateFn
}

// NewMethod creates a new HMAC Method definition.
//
// Parameters:
//   - name: a human-readable identifier for the method (e.g., "HMAC_with_SHA256").
//   - kind: the crypto.Hash algorithm to use.
//   - keySize: expected key size in bytes for key generation.
//
// It does not validate the availability of the provided hash at creation time;
// availability is checked when Digest/Validate are called.
func NewMethod(name string, kind crypto.Hash, keySize int, options ...Option) *Method {
	opts := NewOptions(options...)
	return &Method{
		name:       name,
		kind:       kind,
		keySize:    keySize,
		keyFn:      opts.keyFn,
		digestFn:   opts.digestFn,
		validateFn: opts.validateFn,
	}
}

// Name returns the method's configured name.
func (m *Method) Name() string {
	assert.NotNil(m, "method is nil")
	return m.name
}

// GenerateKey generates a new random key suitable for the HMAC method.
// The returned key length equals the method's configured key size.
func (m *Method) GenerateKey() types.Bytes {
	assert.NotNil(m, "method is nil")
	assert.NotNil(m.keyFn, "method keyFn is nil")
	return m.keyFn(m.keySize)
}

// Digest computes the HMAC of data using this Method and the provided key.
//
// Parameters:
//   - key: secret key to initialize the HMAC.
//   - data: message to authenticate.
//
// Behavior:
//   - Verifies that the method and its hash are valid/available.
//   - Initializes an HMAC with the method's hash and the provided key.
//   - Writes the data and returns the computed digest.
//
// Returns:
//   - digest: the calculated HMAC.
//   - error: wrapped with ErrDigest on failure. Possible causes include:
//   - ErrMethodIsNil if the method is nil.
//   - ErrHashNotAvailable if the hash is not available.
//   - I/O style errors returned from the underlying Write.
func (m *Method) Digest(key types.Bytes, data types.Bytes) (types.Bytes, error) {
	assert.NotNil(m, "method is nil")
	assert.NotNil(m.digestFn, "method digestFn is nil")
	digest, err := m.digestFn(m, key, data)
	if err != nil {
		return nil, ErrDigest(err)
	}
	return digest, nil
}

// Validate checks whether the provided digest matches the HMAC of data
// computed with this Method and key.
//
// Parameters:
//   - key: secret key used to compute the HMAC.
//   - digest: expected HMAC to validate against.
//   - data: message whose authenticity/integrity is being verified.
//
// Behavior:
//   - Verifies method and hash availability.
//   - Recomputes the HMAC and compares using constant-time equality.
//
// Returns:
//   - ok: true if the digests are equal, false otherwise.
//   - error: wrapped with ErrValidation on failure. Possible causes include:
//   - ErrMethodIsNil if the method is nil.
//   - ErrHashNotAvailable if the hash is not available.
//   - Errors produced while recomputing the digest.
func (m *Method) Validate(key types.Bytes, digest types.Bytes, data types.Bytes) (bool, error) {
	assert.NotNil(m, "method is nil")
	assert.NotNil(m.validateFn, "method validateFn is nil")
	ok, err := m.validateFn(m, key, digest, data)
	if err != nil {
		return false, ErrValidation(err)
	}
	return ok, nil
}
