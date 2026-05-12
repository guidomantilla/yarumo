// Package rsassas provides RSA key generation, signing, and verification
// supporting both PSS and PKCS#1 v1.5 padding schemes, with pluggable hash
// functions and a thread-safe name-based registry.
//
// # Config-driven algorithm selection
//
// *Method implements encoding.TextMarshaler / encoding.TextUnmarshaler.
// MarshalText emits the registered algorithm name; UnmarshalText resolves a
// name against the package registry (via Get) and overwrites the receiver.
// This makes Method directly compatible with libraries that honor the
// encoding interfaces — including encoding/json, viper, kong, and koanf —
// so deployments can load RSA signing algorithm choice from YAML/JSON/TOML config.
//
// Caveat: UnmarshalText resolves against whatever the registry contains at
// the time of the call. Custom methods registered via Register after config
// load will not resolve here; callers that need late-bound lookup should
// call Get(name) directly.
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
