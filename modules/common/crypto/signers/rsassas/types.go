// Package rsassas provides RSA key generation, signing, and verification
// supporting both PSS and PKCS#1 v1.5 padding schemes, with pluggable hash
// functions and a thread-safe name-based registry.
package rsassas

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

// Padding represents the RSA signature padding scheme.
type Padding int

const (
	// PSS is the Probabilistic Signature Scheme (RSA-PSS).
	PSS Padding = iota

	// PKCS1v15 is the deterministic PKCS#1 v1.5 padding scheme.
	PKCS1v15
)

// KeyFn is the function type for key generation.
type KeyFn func(bits int) (*rsa.PrivateKey, error)

// SignFn is the function type for signing.
type SignFn func(method *Method, key *rsa.PrivateKey, data ctypes.Bytes) (ctypes.Bytes, error)

// VerifyFn is the function type for verification.
type VerifyFn func(method *Method, key *rsa.PublicKey, signature ctypes.Bytes, data ctypes.Bytes) (bool, error)
