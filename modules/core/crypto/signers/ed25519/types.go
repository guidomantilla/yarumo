// Package ed25519 provides Ed25519 key generation, signing, and verification
// with pluggable function fields and a thread-safe name-based registry.
//
// # Config-driven algorithm selection
//
// *Method implements encoding.TextMarshaler / encoding.TextUnmarshaler.
// MarshalText emits the registered algorithm name; UnmarshalText resolves a
// name against the package registry (via Get) and overwrites the receiver.
// This makes Method directly compatible with libraries that honor the
// encoding interfaces — including encoding/json, viper, kong, and koanf —
// so deployments can load Ed25519 algorithm choice from YAML/JSON/TOML config.
//
// Caveat: UnmarshalText resolves against whatever the registry contains at
// the time of the call. Custom methods registered via Register after config
// load will not resolve here; callers that need late-bound lookup should
// call Get(name) directly.
//
// # Recommended entry point for string-named algorithms
//
// Digest(name, key, data) and Validate(name, key, digest, data) are the
// recommended top-level helpers for callers that load the algorithm name
// from config. They each perform a single Get, parse the PEM-encoded
// Ed25519 key, and forward to the corresponding Method operation.
package ed25519

import (
	"crypto/ed25519"

	ctypes "github.com/guidomantilla/yarumo/core/common/types"
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
