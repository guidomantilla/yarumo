package hashes

import (
	"crypto"

	"github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/types"
)

var (
	SHA256      = NewMethod("SHA256", crypto.SHA256)
	SHA512      = NewMethod("SHA512", crypto.SHA512)
	SHA3_256    = NewMethod("SHA3_256", crypto.SHA3_256)
	SHA3_512    = NewMethod("SHA3_512", crypto.SHA3_512)
	BLAKE2b_256 = NewMethod("BLAKE2b_256", crypto.BLAKE2b_256)
	BLAKE2b_512 = NewMethod("BLAKE2b_512", crypto.BLAKE2b_512)
)

type Method struct {
	name   string
	kind   crypto.Hash
	hashFn HashFn
}

func NewMethod(name string, kind crypto.Hash, options ...Option) *Method {
	opts := NewOptions(options...)
	return &Method{
		name:   name,
		kind:   kind,
		hashFn: opts.hashFn,
	}
}

func (m *Method) Name() string {
	assert.NotNil(m, "method is nil")

	return m.name
}

// Hash computes the digest of the given data using the specified hash function.
//
// Notes: It delegates to hashes.HashFn. This is a function that takes a crypto.Hash identifier (e.g., crypto.SHA256). Must be available.
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
func (m *Method) Hash(data types.Bytes) types.Bytes {
	assert.NotNil(m, "method is nil")
	assert.NotNil(m.hashFn, "method hashFn is nil")

	return m.hashFn(m.kind, data)
}
