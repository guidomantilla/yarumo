package aead

import (
	cassert "github.com/guidomantilla/yarumo/common/assert"
	ctypes "github.com/guidomantilla/yarumo/common/types"
)

// Predefined AEAD methods registered at package init.
var (
	AES_128_GCM        = NewMethod("AES_128_GCM", 16, 12, aesgcm)
	AES_256_GCM        = NewMethod("AES_256_GCM", 32, 12, aesgcm)
	CHACHA20_POLY1305  = NewMethod("ChaCha20-Poly1305", 32, 12, chacha20Poly1305)
	XCHACHA20_POLY1305 = NewMethod("XChaCha20-Poly1305", 32, 24, xchacha20Poly1305)
)

// Method holds the configuration for an AEAD algorithm.
type Method struct {
	name      string
	keySize   int
	nonceSize int
	kind      AeadFn
	keyFn     KeyFn
	encryptFn EncryptFn
	decryptFn DecryptFn
}

// NewMethod creates a new AEAD Method with the given name, key size, nonce size, and cipher factory.
func NewMethod(name string, keySize, nonceSize int, kind AeadFn, options ...Option) *Method {
	cassert.NotEmpty(name, "name is empty")

	opts := NewOptions(options...)

	return &Method{
		name:      name,
		keySize:   keySize,
		nonceSize: nonceSize,
		kind:      kind,
		keyFn:     opts.keyFn,
		encryptFn: opts.encryptFn,
		decryptFn: opts.decryptFn,
	}
}

// Name returns the method's algorithm name.
func (m *Method) Name() string {
	cassert.NotNil(m, "method is nil")
	return m.name
}

// GenerateKey generates a random symmetric key for this AEAD method.
func (m *Method) GenerateKey() (ctypes.Bytes, error) {
	cassert.NotNil(m, "method is nil")
	cassert.NotNil(m.keyFn, "method keyFn is nil")

	key, err := m.keyFn(m)
	if err != nil {
		return nil, ErrKeyGeneration(err)
	}

	return key, nil
}

// Encrypt encrypts data with the given key and additional authenticated data.
func (m *Method) Encrypt(key ctypes.Bytes, data ctypes.Bytes, aad ctypes.Bytes) (ctypes.Bytes, error) {
	cassert.NotNil(m, "method is nil")
	cassert.NotNil(m.encryptFn, "method encryptFn is nil")

	ciphered, err := m.encryptFn(m, key, data, aad)
	if err != nil {
		return nil, ErrEncryption(err)
	}

	return ciphered, nil
}

// Decrypt decrypts ciphered data with the given key and additional authenticated data.
func (m *Method) Decrypt(key ctypes.Bytes, ciphered ctypes.Bytes, aad ctypes.Bytes) (ctypes.Bytes, error) {
	cassert.NotNil(m, "method is nil")
	cassert.NotNil(m.decryptFn, "method decryptFn is nil")

	ciphered, err := m.decryptFn(m, key, ciphered, aad)
	if err != nil {
		return nil, ErrDecryption(err)
	}

	return ciphered, nil
}
