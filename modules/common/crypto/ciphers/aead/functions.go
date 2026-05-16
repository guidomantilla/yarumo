package aead

import (
	"crypto/aes"
	"crypto/cipher"

	"golang.org/x/crypto/chacha20poly1305"

	crandom "github.com/guidomantilla/yarumo/common/crypto/random"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
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

	out, err := crandom.Bytes(method.keySize)
	if err != nil {
		return nil, cerrs.Wrap(ErrKeyGenerationFailed, err)
	}

	return out, nil
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

	nonce, nonceErr := crandom.Bytes(method.nonceSize)
	if nonceErr != nil {
		return nil, cerrs.Wrap(ErrNonceGenerationFailed, nonceErr)
	}

	ciphered := aead.Seal(nil, nonce, data, aad)

	out := make([]byte, 0, len(nonce)+len(ciphered))
	out = append(out, nonce...)
	out = append(out, ciphered...)

	return out, nil
}

// Encrypt is the recommended entry point for callers that receive the
// algorithm name as a string (e.g. loaded from config, a request header, or
// a database column). It performs a single registry Get and forwards to
// Method.Encrypt, returning ErrAlgorithmNotSupported when name is not
// registered.
//
// For callers that already hold a *Method (predefined or returned by Get),
// use Method.Encrypt directly; Encrypt exists purely to collapse the
// "Get + Encrypt" boilerplate at the config↔runtime seam.
func Encrypt(name string, key, data, aad ctypes.Bytes) (ctypes.Bytes, error) {
	method, err := Get(name)
	if err != nil {
		return nil, err
	}

	return method.Encrypt(key, data, aad)
}

// Decrypt is the recommended entry point for callers that receive the
// algorithm name as a string. It performs a single registry Get and forwards
// to Method.Decrypt, returning ErrAlgorithmNotSupported when name is not
// registered.
func Decrypt(name string, key, data, aad ctypes.Bytes) (ctypes.Bytes, error) {
	method, err := Get(name)
	if err != nil {
		return nil, err
	}

	return method.Decrypt(key, data, aad)
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
