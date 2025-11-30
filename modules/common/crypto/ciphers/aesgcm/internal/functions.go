package internal

import (
	"crypto/aes"
	"crypto/cipher"

	"github.com/guidomantilla/yarumo/common/types"
	"golang.org/x/crypto/chacha20poly1305"
)

func AESGCM(key types.Bytes, nonceSize int) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCMWithNonceSize(block, nonceSize)
}

func ChaCha20Poly1305(key types.Bytes, _ int) (cipher.AEAD, error) {
	return chacha20poly1305.New(key)
}
