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
//
// # Streaming
//
// Method.EncryptStream and Method.DecryptStream implement chunked AEAD
// suitable for multi-megabyte or unbounded io.Reader / io.Writer pipelines.
// The input is split into fixed-size StreamFrameSize (64 KiB) plaintext
// frames; each frame is sealed independently with the underlying AEAD
// primitive and emitted on the wire with a 4-byte big-endian uint32 length
// prefix. A zero-length frame marks end-of-stream and lets the decoder
// distinguish a clean close from truncation.
//
// Frame format on the wire:
//
//	[ 4-byte BE uint32 frame length ][ ciphertext = nonce || enc(plain) || tag ]
//	[ 4-byte BE uint32 frame length ][ ciphertext                              ]
//	...
//	[ 4-byte BE uint32 = 0 ]   ← end-of-stream sentinel
//
// Each frame's per-call AAD is (caller_aad || 8-byte BE frame counter),
// binding the frame's position into the AEAD authentication tag. This
// protects against frame reordering, duplication, and dropping. The
// per-frame random nonce produced by the underlying AEAD primitive is
// embedded inside the ciphertext frame itself, so the streaming API does
// not need to manage nonce derivation.
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
