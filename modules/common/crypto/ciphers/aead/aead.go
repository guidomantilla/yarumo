package aead

import (
	"github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/crypto/ciphers/aead/internal"
	"github.com/guidomantilla/yarumo/common/types"
)

var (
	AES_128_GCM        = NewMethod("AES_128_GCM", 16, 12, internal.AESGCM)
	AES_256_GCM        = NewMethod("AES_256_GCM", 32, 12, internal.AESGCM)
	CHACHA20_POLY1305  = NewMethod("ChaCha20-Poly1305", 32, 12, internal.ChaCha20Poly1305)
	XCHACHA20_POLY1305 = NewMethod("XChaCha20-Poly1305", 32, 24, internal.XChaCha20Poly1305)
)

type Method struct {
	name      string
	keySize   int
	nonceSize int
	kind      AeadFn
	keyFn     KeyFn
	encryptFn EncryptFn
	decryptFn DecryptFn
}

func NewMethod(name string, keySize, nonceSize int, kind AeadFn, options ...Option) *Method {
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

func (m *Method) Name() string {
	assert.NotNil(m, "method is nil")
	return m.name
}

func (m *Method) GenerateKey() (types.Bytes, error) {
	assert.NotNil(m, "method is nil")
	assert.NotNil(m.keyFn, "method keyFn is nil")

	key, err := m.keyFn(m)
	if err != nil {
		return nil, ErrKeyGeneration(err)
	}

	return key, nil
}

func (m *Method) Encrypt(key types.Bytes, data types.Bytes, aad types.Bytes) (types.Bytes, error) {
	assert.NotNil(m, "method is nil")
	assert.NotNil(m.encryptFn, "method encryptFn is nil")

	ciphered, err := m.encryptFn(m, key, data, aad)
	if err != nil {
		return nil, ErrEncryption(err)
	}

	return ciphered, nil
}

func (m *Method) Decrypt(key types.Bytes, ciphered types.Bytes, aad types.Bytes) (types.Bytes, error) {
	assert.NotNil(m, "method is nil")
	assert.NotNil(m.decryptFn, "method decryptFn is nil")

	ciphered, err := m.decryptFn(m, key, ciphered, aad)
	if err != nil {
		return nil, ErrDecryption(err)
	}

	return ciphered, nil
}
