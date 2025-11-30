package aesgcm

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"

	"github.com/guidomantilla/yarumo/common/errs"
	"github.com/guidomantilla/yarumo/common/types"
)

func Encrypt(method *Method, key types.Bytes, plaintext types.Bytes, aad types.Bytes) ([]byte, error) {
	if method == nil {
		return nil, nil
	}
	if len(key) != method.keySize {
		return nil, nil
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCMWithNonceSize(block, method.nonceSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM cipher: %w", err)
	}

	nonce := make([]byte, method.nonceSize)
	if _, err = rand.Read(nonce); err != nil {
		return nil, errs.Wrap(nil, err)
	}

	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, aad)
	return ciphertext, nil
}
