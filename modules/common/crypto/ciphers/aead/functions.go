package aead

import (
	"crypto/aes"
	"crypto/cipher"

	"golang.org/x/crypto/chacha20poly1305"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
	crandom "github.com/guidomantilla/yarumo/common/random"
	ctypes "github.com/guidomantilla/yarumo/common/types"
	cutils "github.com/guidomantilla/yarumo/common/utils"
)

func aesgcm(key ctypes.Bytes, nonceSize int) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return cipher.NewGCMWithNonceSize(block, nonceSize)
}

func chacha20Poly1305(key ctypes.Bytes, _ int) (cipher.AEAD, error) {
	return chacha20poly1305.New(key)
}

func xchacha20Poly1305(key ctypes.Bytes, _ int) (cipher.AEAD, error) {
	return chacha20poly1305.NewX(key)
}

func key(method *Method) (ctypes.Bytes, error) {
	if method == nil {
		return nil, ErrMethodInvalid
	}

	if cutils.NotIn(method.keySize, 16, 32) {
		return nil, ErrKeySizeInvalid
	}

	if cutils.NotIn(method.nonceSize, 12, 24) {
		return nil, ErrNonceSizeInvalid
	}

	return crandom.Bytes(method.keySize), nil
}

func encrypt(method *Method, key ctypes.Bytes, data ctypes.Bytes, aad ctypes.Bytes) (ctypes.Bytes, error) {
	if method == nil {
		return nil, ErrMethodInvalid
	}

	if len(key) != method.keySize {
		return nil, ErrKeyInvalid
	}

	aead, err := method.kind(key, method.nonceSize)
	if err != nil {
		return nil, cerrs.Wrap(ErrCipherInitFailed, err)
	}

	nonce := crandom.Bytes(method.nonceSize)
	ciphered := aead.Seal(nil, nonce, data, aad)

	out := make([]byte, 0, len(nonce)+len(ciphered))
	out = append(out, nonce...)
	out = append(out, ciphered...)

	return out, nil
}

func decrypt(method *Method, key ctypes.Bytes, ciphered ctypes.Bytes, aad ctypes.Bytes) (ctypes.Bytes, error) {
	if method == nil {
		return nil, ErrMethodInvalid
	}

	if len(key) != method.keySize {
		return nil, ErrKeyInvalid
	}

	if len(ciphered) < method.nonceSize {
		return nil, ErrCiphertextTooShort
	}

	aead, err := method.kind(key, method.nonceSize)
	if err != nil {
		return nil, cerrs.Wrap(ErrCipherInitFailed, err)
	}

	nonce := ciphered[:method.nonceSize]
	ciphered = ciphered[method.nonceSize:]

	out, err := aead.Open(nil, nonce, ciphered, aad)
	if err != nil {
		return nil, cerrs.Wrap(ErrDecryptFailed, err)
	}

	return out, nil
}
