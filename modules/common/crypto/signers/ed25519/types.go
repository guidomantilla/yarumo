// Package ed25519 provides Ed25519 key generation, signing, and verification
// with pluggable function fields and a thread-safe name-based registry.
package ed25519

import (
	"crypto/ed25519"

	ctypes "github.com/guidomantilla/yarumo/common/types"
)

// Type compliance.
var (
	_ KeyFn    = key
	_ SignFn   = sign
	_ VerifyFn = verify
)

// KeyFn is the function type for key generation.
type KeyFn func() (ed25519.PublicKey, ed25519.PrivateKey, error)

// SignFn is the function type for signing.
type SignFn func(method *Method, key *ed25519.PrivateKey, data ctypes.Bytes) (ctypes.Bytes, error)

// VerifyFn is the function type for verification.
type VerifyFn func(method *Method, key *ed25519.PublicKey, signature ctypes.Bytes, data ctypes.Bytes) (bool, error)
