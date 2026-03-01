// Package ecdsas provides ECDSA key generation, signing, and verification
// using pluggable elliptic curves and a thread-safe name-based registry.
package ecdsas

import (
	"crypto/ecdsa"

	ctypes "github.com/guidomantilla/yarumo/common/types"
)

// Type compliance.
var (
	_ KeyFn    = key
	_ SignFn   = sign
	_ VerifyFn = verify
)

// Format represents the encoding used for ECDSA signatures produced/consumed by this package.
type Format int

const (
	// RS is a raw concatenation format: r || s (big-endian, zero-left padded to key size).
	// Commonly used in JOSE/JWT and WebAuthn.
	RS Format = iota

	// ASN1 is the DER-encoded ASN.1 sequence format used by the standard library.
	ASN1
)

// KeyFn is the function type for key generation.
type KeyFn func(method *Method) (*ecdsa.PrivateKey, error)

// SignFn is the function type for signing.
type SignFn func(method *Method, key *ecdsa.PrivateKey, data ctypes.Bytes, format Format) (ctypes.Bytes, error)

// VerifyFn is the function type for verification.
type VerifyFn func(method *Method, key *ecdsa.PublicKey, signature ctypes.Bytes, data ctypes.Bytes, format Format) (bool, error)
