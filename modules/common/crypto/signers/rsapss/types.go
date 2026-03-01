// Package rsapss provides RSA-PSS key generation, signing, and verification
// using pluggable hash functions, configurable salt lengths, and a thread-safe
// name-based registry.
package rsapss

import (
	"crypto/rsa"

	ctypes "github.com/guidomantilla/yarumo/common/types"
)

// Type compliance.
var (
	_ KeyFn    = key
	_ SignFn   = sign
	_ VerifyFn = verify
)

// KeyFn is the function type for key generation.
type KeyFn func(bits int) (*rsa.PrivateKey, error)

// SignFn is the function type for signing.
type SignFn func(method *Method, key *rsa.PrivateKey, data ctypes.Bytes) (ctypes.Bytes, error)

// VerifyFn is the function type for verification.
type VerifyFn func(method *Method, key *rsa.PublicKey, signature ctypes.Bytes, data ctypes.Bytes) (bool, error)
