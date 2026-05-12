// Package aead provides authenticated encryption with associated data (AEAD)
// using pluggable cipher factories, configurable key and nonce sizes, and a
// thread-safe name-based registry.
//
// # Config-driven algorithm selection
//
// *Method implements encoding.TextMarshaler / encoding.TextUnmarshaler.
// MarshalText emits the registered algorithm name; UnmarshalText resolves a
// name against the package registry (via Get) and overwrites the receiver.
// This makes Method directly compatible with libraries that honor the
// encoding interfaces — including encoding/json, viper, kong, and koanf —
// so deployments can load cipher choice from YAML/JSON/TOML config.
//
// Caveat: UnmarshalText resolves against whatever the registry contains at
// the time of the call. Custom methods registered via Register after config
// load will not resolve here; callers that need late-bound lookup should
// call Get(name) directly.
package aead

import (
	"crypto/cipher"

	ctypes "github.com/guidomantilla/yarumo/common/types"
)

// Type compliance.
var (
	_ AeadFn    = aesgcm
	_ AeadFn    = chacha20Poly1305
	_ AeadFn    = xchacha20Poly1305
	_ KeyFn     = key
	_ EncryptFn = encrypt
	_ DecryptFn = decrypt
)

// AeadFn is the function type for creating an AEAD cipher from a key and nonce size.
type AeadFn func(key ctypes.Bytes, nonceSize int) (cipher.AEAD, error)

// KeyFn is the function type for key generation.
type KeyFn func(method *Method) (ctypes.Bytes, error)

// EncryptFn is the function type for encryption.
type EncryptFn func(method *Method, key ctypes.Bytes, data ctypes.Bytes, aad ctypes.Bytes) (ctypes.Bytes, error)

// DecryptFn is the function type for decryption.
type DecryptFn func(method *Method, key, ciphered ctypes.Bytes, aad ctypes.Bytes) (ctypes.Bytes, error)
