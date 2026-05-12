// Package hybrid provides hybrid public-key encryption (HPKE, RFC 9180) with
// pluggable cipher-suite combinations and a thread-safe name-based registry.
//
// # Why hybrid encryption?
//
// Pure asymmetric ciphers like RSA-OAEP cannot encrypt payloads larger than
// roughly `keySize - 2*hashSize - 2` bytes (about 190 bytes for RSA-2048 +
// SHA-256). Hybrid encryption sidesteps that limit by combining:
//
//  1. An asymmetric Key Encapsulation Mechanism (KEM) that ships an ephemeral
//     symmetric key to the recipient under their long-term public key.
//  2. A symmetric Authenticated Encryption with Associated Data (AEAD) cipher
//     that encrypts the actual payload using that ephemeral key.
//
// The resulting wire format is therefore the encapsulated key followed by the
// AEAD ciphertext. Decryption requires the matching long-term private key.
//
// # Cipher suites
//
// This package wraps github.com/cloudflare/circl/hpke and exposes RFC 9180
// mode 0 (base mode). The package-level predefined Method is:
//
//   - HPKE_X25519_HKDF_SHA256_AES_256_GCM — KEM=DHKEM(X25519, HKDF-SHA256),
//     KDF=HKDF-SHA256, AEAD=AES-256-GCM. Identifiers 0x0020 / 0x0001 / 0x0002
//     per RFC 9180 §7.1, §7.2, §7.3.
//
// The HPKE Method works with X25519 key pairs represented as
// hpke.KEM_X25519_HKDF_SHA256 KEM scheme keys (kem.PublicKey /
// kem.PrivateKey from circl). GenerateKey returns both halves.
//
// # Caveats
//
// HPKE base mode does not authenticate the sender — anybody with the
// recipient's public key can produce a valid ciphertext. Pair this with a
// signature (see common/crypto/signers) when sender authenticity matters.
//
// The info argument supplied to Encrypt/Decrypt is bound into the HPKE key
// schedule; mismatched info between sender and receiver causes Decrypt to
// fail. Use info to bind the ciphertext to a protocol context.
//
// # Config-driven algorithm selection
//
// *Method implements encoding.TextMarshaler / encoding.TextUnmarshaler.
// MarshalText emits the registered algorithm name; UnmarshalText resolves a
// name against the package registry (via Get) and overwrites the receiver.
// This makes Method directly compatible with libraries that honor the
// encoding interfaces — including encoding/json, viper, kong, and koanf —
// so deployments can load HPKE cipher-suite choice from YAML/JSON/TOML
// config.
//
// Caveat: UnmarshalText resolves against whatever the registry contains at
// the time of the call. Custom methods registered via Register after config
// load will not resolve here; callers that need late-bound lookup should
// call Get(name) directly.
package hybrid

import (
	"github.com/cloudflare/circl/kem"

	ctypes "github.com/guidomantilla/yarumo/common/types"
)

// Type compliance.
var (
	_ KeyFn     = generateKey
	_ EncryptFn = encrypt
	_ DecryptFn = decrypt
)

// KeyFn is the function type for KEM key pair generation. It returns the
// recipient's public/private key pair encoded in the KEM-specific wire form
// (X25519 raw bytes for KEM_X25519_HKDF_SHA256).
type KeyFn func(method *Method) (kem.PublicKey, kem.PrivateKey, error)

// EncryptFn is the function type for hybrid encryption. The output is the
// concatenation of the KEM encapsulated key and the AEAD ciphertext.
type EncryptFn func(method *Method, recipientPub kem.PublicKey, plaintext, info ctypes.Bytes) (ctypes.Bytes, error)

// DecryptFn is the function type for hybrid decryption. It expects the wire
// format produced by EncryptFn and the recipient's private key matching the
// public key used to encrypt.
type DecryptFn func(method *Method, recipientPriv kem.PrivateKey, ciphertext, info ctypes.Bytes) (ctypes.Bytes, error)
