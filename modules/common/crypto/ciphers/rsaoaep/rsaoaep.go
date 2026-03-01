package rsaoaep

import (
	"crypto"
	"crypto/rsa"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	ctypes "github.com/guidomantilla/yarumo/common/types"
	cutils "github.com/guidomantilla/yarumo/common/utils"
)

// Predefined RSA-OAEP methods registered at package init.
var (
	RSA_OAEP_SHA256 = NewMethod("RSA-OAEP-SHA256", crypto.SHA256, []int{2048, 3072, 4096})
	RSA_OAEP_SHA512 = NewMethod("RSA-OAEP-SHA512", crypto.SHA512, []int{3072, 4096})
)

// Method holds the configuration for an RSA-OAEP algorithm.
type Method struct {
	name            string
	kind            crypto.Hash
	allowedKeySizes []int
	keyFn           KeyFn
	encryptFn       EncryptFn
	decryptFn       DecryptFn
}

// NewMethod creates a new RSA-OAEP Method with the given name, hash, and allowed key sizes.
func NewMethod(name string, kind crypto.Hash, allowedKeySizes []int, options ...Option) *Method {
	cassert.NotEmpty(name, "name is empty")

	opts := NewOptions(options...)

	return &Method{
		name:            name,
		kind:            kind,
		allowedKeySizes: allowedKeySizes,
		keyFn:           opts.keyFn,
		encryptFn:       opts.encryptFn,
		decryptFn:       opts.decryptFn,
	}
}

// Name returns the method's algorithm name.
func (m *Method) Name() string {
	cassert.NotNil(m, "method is nil")

	return m.name
}

// GenerateKey generates a new RSA private key of the specified bit size.
func (m *Method) GenerateKey(bits int) (*rsa.PrivateKey, error) {
	cassert.NotNil(m, "method is nil")
	cassert.NotNil(m.keyFn, "method keyFn is nil")

	if cutils.NotIn(bits, m.allowedKeySizes...) {
		return nil, ErrKeyGeneration(ErrKeySizeNotAllowed)
	}

	k, err := m.keyFn(m, bits)
	if err != nil {
		return nil, ErrKeyGeneration(err)
	}

	return k, nil
}

// Encrypt encrypts data with the given public key and optional label.
func (m *Method) Encrypt(key *rsa.PublicKey, data ctypes.Bytes, label ctypes.Bytes) (ctypes.Bytes, error) {
	cassert.NotNil(m, "method is nil")
	cassert.NotNil(m.encryptFn, "method encryptFn is nil")

	ciphered, err := m.encryptFn(m, key, data, label)
	if err != nil {
		return nil, ErrEncryption(err)
	}

	return ciphered, nil
}

// Decrypt decrypts ciphered data with the given private key and optional label.
func (m *Method) Decrypt(key *rsa.PrivateKey, ciphered ctypes.Bytes, label ctypes.Bytes) (ctypes.Bytes, error) {
	cassert.NotNil(m, "method is nil")
	cassert.NotNil(m.decryptFn, "method decryptFn is nil")

	plain, err := m.decryptFn(m, key, ciphered, label)
	if err != nil {
		return nil, ErrDecryption(err)
	}

	return plain, nil
}
