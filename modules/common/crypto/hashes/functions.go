package hashes

import (
	"crypto"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	ctypes "github.com/guidomantilla/yarumo/common/types"
)

// Hash computes the digest of the given data using the specified hash function.
// It panics if the hash function is not available.
func Hash(hash crypto.Hash, data ctypes.Bytes) ctypes.Bytes {
	cassert.True(hash.Available(), "hash function not available. call crypto.RegisterHash(...)")
	h := hash.New()
	_, _ = h.Write(data)

	return h.Sum(nil)
}
