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

// GenerateKey generates a new random symmetric key for the given method.
//
// Notes: It delegates to hmacs.KeyFn. This is a function that takes an HMAC Method specifying the required key size.
//
// Parameters: none
//
// Behavior:
//   - Returns an error if method is nil.
//   - Uses random.Bytes to generate a cryptographically secure key of
//     exactly method.keySize bytes.
//
// Returns:
//   - A newly generated key as a byte slice.
//   - An error if the method is invalid.
//
// Notes:
//   - This is intended for symmetric algorithms (AES, ChaCha20, etc.).
//   - The function never panics and never returns a partial key.
//   - The caller is responsible for securely storing the key.
func (m *Method) GenerateKey() (types.Bytes, error) {
	assert.NotNil(m, "method is nil")
	assert.NotNil(m.keyFn, "method keyFn is nil")

	key, err := m.keyFn(m)
	if err != nil {
		return nil, ErrKeyGeneration(err)
	}

	return key, nil
}

// Digest computes an HMAC authentication tag over the provided data using
// the given method and key.
//
// Notes: it delegates to hmacs.DigestFn. This is a function that takes an HMAC Method specifying the underlying hash function.
//
// Parameters:
//   - key: the secret key used for HMAC (any length allowed by HMAC).
//   - data: the message to authenticate.
//
// Behavior:
//   - Returns an error if method is nil.
//   - Returns an error if the underlying hash function is not available
//     (method.kind.Available() == false).
//   - Creates a new HMAC instance using method.kind.New and the provided key.
//   - Writes the input data into the HMAC.
//   - Produces the final authentication tag via h.Sum(nil).
//
// Returns:
//   - A byte slice containing the HMAC tag.
//   - An error if writing to the HMAC fails (extremely rare).
//
// Notes:
//   - HMAC does not encrypt data; it provides integrity and authenticity.
//   - The output size depends on the hash function (e.g., SHA-256 → 32 bytes).
//   - The function never panics and never returns a partial digest.
func (m *Method) Digest(key types.Bytes, data types.Bytes) (types.Bytes, error) {
	assert.NotNil(m, "method is nil")
	assert.NotNil(m.digestFn, "method digestFn is nil")

	digest, err := m.digestFn(m, key, data)
	if err != nil {
		return nil, ErrDigest(err)
	}

	return digest, nil
}

// Validate checks whether the provided HMAC digest matches the HMAC computed
// over the given data using the specified method and key.
//
// Notes: it delegates to hmacs.ValidateFn. This is a function that takes an HMAC Method specifying the underlying hash function.
//
// Parameters:
//   - key: the secret key used for HMAC computation.
//   - digest_: the expected HMAC tag to compare against.
//   - data: the message that should match the provided digest.
//
// Behavior:
//   - Returns an error if method is nil.
//   - Returns an error if the underlying hash function is not available
//     (method.kind.Available() == false).
//   - Recomputes the HMAC tag for the provided data and key using digest().
//   - Compares the expected digest with the computed one using hmac.Equal,
//     which provides constant-time comparison.
//
// Returns:
//   - (true, nil)  if the digest matches.
//   - (false, nil) if the digest does not match.
//   - (false, err) if HMAC computation fails.
//
// Notes:
//   - This function provides integrity/authenticity verification only;
//     it does not check for freshness or prevent replay attacks.
//   - Uses constant-time comparison to avoid timing side-channel leaks.
//   - Never panics and never exposes partial HMAC data.
func (m *Method) Validate(key types.Bytes, digest types.Bytes, data types.Bytes) (bool, error) {
	assert.NotNil(m, "method is nil")
	assert.NotNil(m.validateFn, "method validateFn is nil")

	ok, err := m.validateFn(m, key, digest, data)
	if err != nil {
		return false, ErrValidation(err)
	}

	return ok, nil
}
