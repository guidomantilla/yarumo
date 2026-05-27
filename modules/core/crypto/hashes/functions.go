package hashes

import (
	"crypto"

	ctypes "github.com/guidomantilla/yarumo/core/common/types"
)

// Hash computes the digest of the given data using the specified hash function.
// It returns an error if the hash function driver was not registered (e.g. the
// caller did not blank-import the implementation package). The signature
// matches the workspace-wide crypto operation pattern of (result, error).
func Hash(hash crypto.Hash, data ctypes.Bytes) (ctypes.Bytes, error) {
	if !hash.Available() {
		return nil, ErrHashFunctionUnavailable
	}

	h := hash.New()
	_, _ = h.Write(data)

	return h.Sum(nil), nil
}

// Compute is the recommended entry point for callers that receive the
// algorithm name as a string (e.g. loaded from config, a request header, or
// a database column). It performs a single registry Get and forwards to
// Method.Hash, returning ErrAlgorithmNotSupported when name is not
// registered.
//
// For callers that already hold a *Method (predefined or returned by Get),
// use Method.Hash directly; Compute exists purely to collapse the
// "Get + Hash" boilerplate at the config↔runtime seam.
func Compute(name string, data ctypes.Bytes) (ctypes.Bytes, error) {
	method, err := Get(name)
	if err != nil {
		return nil, err
	}

	return method.Hash(data)
}
