package hashes

import (
	"crypto"

	"github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/types"
)

// Hash computes the digest of the given data using the specified hash function.
//
// Parameters:
//   - hash: a crypto.Hash identifier (e.g., crypto.SHA256). Must be available;
//     otherwise the function panics via assert.
//   - data: the input data to hash.
//
// Behavior:
//   - Ensures the hash function is available.
//   - Creates a new hash instance using hash.New().
//   - Writes the input data into the hasher.
//   - Returns the final digest produced by h.Sum(nil).
//
// Returns:
//   - The hash digest as a byte slice.
//
// Notes:
//   - Write errors are ignored (h.Write never returns an error for standard
//     hash.Hash implementations).
//   - The function does not return an error; it panics only if the hash
//     function is not registered.
func Hash(hash crypto.Hash, data types.Bytes) types.Bytes {
	assert.True(hash.Available(), "hash function not available. call crypto.RegisterHash(...)")
	h := hash.New()
	_, _ = h.Write(data)
	return h.Sum(nil)
}
