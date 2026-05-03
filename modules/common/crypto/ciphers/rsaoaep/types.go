// Package rsaoaep provides RSA-OAEP key generation, encryption, and decryption
// using pluggable hash functions, configurable key sizes, and a thread-safe
// name-based registry.
package rsaoaep

import (
	"crypto/rsa"

	ctypes "github.com/guidomantilla/yarumo/common/types"
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
