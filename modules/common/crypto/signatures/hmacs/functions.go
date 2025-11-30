package hmacs

import (
	"crypto"
	"crypto/hmac"

	"github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/types"
)

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
func Digest(hash crypto.Hash, key types.Bytes, data types.Bytes) types.Bytes {
	assert.True(hash.Available(), "hash function not available. call crypto.RegisterHash(...)")
	h := hmac.New(hash.New, key)
	_, _ = h.Write(data)
	return h.Sum(nil)
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
func Validate(hash crypto.Hash, key types.Bytes, digest types.Bytes, data types.Bytes) bool {
	assert.True(hash.Available(), "hash function not available. call crypto.RegisterHash(...)")
	calculated := Digest(hash, key, data)
	return hmac.Equal(digest, calculated)
}
