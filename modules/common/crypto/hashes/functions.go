package hashes

import (
	"crypto"

	ctypes "github.com/guidomantilla/yarumo/common/types"
)

// Hash computes the digest of the given data using the specified hash function.
// It returns an error if the hash function driver was not registered (e.g. the
// caller did not blank-import the implementation package). The signature
// matches the workspace-wide crypto operation pattern of (result, error).
func Hash(hash crypto.Hash, data ctypes.Bytes) (ctypes.Bytes, error) {
	if !hash.Available() {
		return nil, ErrDigest(ErrHashFunctionUnavailable)
	}

	h := hash.New()
	_, _ = h.Write(data)

	return h.Sum(nil), nil
}
