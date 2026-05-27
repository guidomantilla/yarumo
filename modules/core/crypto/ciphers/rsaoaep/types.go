// Package rsaoaep provides RSA-OAEP key generation, encryption, and decryption
// using pluggable hash functions, configurable key sizes, and a thread-safe
// name-based registry.
//
// # Config-driven algorithm selection
//
// *Method implements encoding.TextMarshaler / encoding.TextUnmarshaler.
// MarshalText emits the registered algorithm name; UnmarshalText resolves a
// name against the package registry (via Get) and overwrites the receiver.
// This makes Method directly compatible with libraries that honor the
// encoding interfaces — including encoding/json, viper, kong, and koanf —
// so deployments can load RSA-OAEP algorithm choice from YAML/JSON/TOML config.
//
// Caveat: UnmarshalText resolves against whatever the registry contains at
// the time of the call. Custom methods registered via Register after config
// load will not resolve here; callers that need late-bound lookup should
// call Get(name) directly.
//
// # Recommended entry point for string-named algorithms
//
// Encrypt(name, key, data, aad) and Decrypt(name, key, data, aad) are the
// recommended top-level helpers for callers that load the algorithm name
// from config. They each perform a single Get, parse the PEM-encoded RSA
// key, and forward to the corresponding Method operation. The aad argument
// is passed through as the OAEP label.
package rsaoaep

import (
	"crypto/rsa"

	ctypes "github.com/guidomantilla/yarumo/core/common/types"
)

// Type compliance.
var (
	_ KeyFn     = key
	_ EncryptFn = encrypt
	_ DecryptFn = decrypt
)

// KeyFn is the function type for key generation.
type KeyFn func(method *Method, bits int) (*rsa.PrivateKey, error)

// EncryptFn is the function type for encryption.
type EncryptFn func(method *Method, key *rsa.PublicKey, data, label ctypes.Bytes) (ctypes.Bytes, error)

// DecryptFn is the function type for decryption.
type DecryptFn func(method *Method, key *rsa.PrivateKey, ciphered, label ctypes.Bytes) (ctypes.Bytes, error)
