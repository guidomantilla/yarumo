package internal

import (
	"crypto/aes"
	"crypto/cipher"

	"golang.org/x/crypto/chacha20poly1305"

	"github.com/guidomantilla/yarumo/common/types"
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

func XChaCha20Poly1305(key types.Bytes, _ int) (cipher.AEAD, error) {
	return chacha20poly1305.NewX(key)
}
