// Package kdfs provides Key Derivation Function (KDF) algorithms with a
// unified Method.Derive API and a thread-safe name-based registry.
//
// Supported families:
//   - HKDF (RFC 5869) — extract-and-expand KDF over an HMAC primitive. The
//     info argument is used as the HKDF context/label binding.
//   - PBKDF2 (RFC 8018) — iterated HMAC for password-based key derivation.
//     The info argument is ignored.
//   - Scrypt (RFC 7914) — memory-hard KDF. The info argument is ignored.
//
// Use HKDF when expanding a high-entropy master secret into one or more keys
// (e.g. deriving an AEAD key from an ECDH shared secret). Use PBKDF2 or
// Scrypt when the input is a low-entropy password or passphrase.
package kdfs

import (
	ctypes "github.com/guidomantilla/yarumo/common/types"
)

// Type compliance.
var (
	_ DeriveFn = hkdfDerive
	_ DeriveFn = pbkdf2Derive
	_ DeriveFn = scryptDerive
)

// DeriveFn is the function type for key derivation. The secret is the input
// keying material (IKM); salt is the (optional, for HKDF) randomness source;
// info is the HKDF-specific context/label (ignored by PBKDF2 and Scrypt);
// length is the requested number of output bytes.
type DeriveFn func(method *Method, secret, salt, info ctypes.Bytes, length int) (ctypes.Bytes, error)

// ConfigFn is the function type used to configure algorithm-specific
// parameters on a Method via Options.
type ConfigFn func(opts *Options)
