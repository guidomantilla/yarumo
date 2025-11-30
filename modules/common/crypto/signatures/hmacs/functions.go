package hmacs

import (
	"crypto/hmac"

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
func Digest(method *Method, key types.Bytes, data types.Bytes) (types.Bytes, error) {
	if method == nil {
		return nil, ErrMethodIsNil
	}
	if !method.kind.Available() {
		return nil, ErrHashNotAvailable
	}

	h := hmac.New(method.kind.New, key)
	_, err := h.Write(data)
	if err != nil {
		return nil, err
	}

	out := h.Sum(nil)
	return out, nil
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
func Validate(method *Method, key types.Bytes, digest types.Bytes, data types.Bytes) (bool, error) {
	if method == nil {
		return false, ErrMethodIsNil
	}
	if !method.kind.Available() {
		return false, ErrHashNotAvailable
	}
	calculated, err := Digest(method, key, data)
	if err != nil {
		return false, err
	}
	ok := hmac.Equal(digest, calculated)
	return ok, nil
}
