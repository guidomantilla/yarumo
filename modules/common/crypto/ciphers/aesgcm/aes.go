package aesgcm

import "github.com/guidomantilla/yarumo/common/crypto/ciphers/aesgcm/internal"

var (
	AES_128_GCM       = NewMethod("AES_128_GCM", 16, 12, internal.AESGCM)
	AES_256_GCM       = NewMethod("AES_256_GCM", 32, 12, internal.AESGCM)
	CHACHA20_POLY1305 = NewMethod("CHACHA20_POLY1305", 32, 12, internal.ChaCha20Poly1305)
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

func NewMethod(name string, keySize, nonceSize int, kind AeadFn) *Method {
	return &Method{
		name:      name,
		keySize:   keySize,
		nonceSize: nonceSize,
		kind:      kind,
	}
}
