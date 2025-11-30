package aesgcm

import (
	"github.com/guidomantilla/yarumo/common/errs"
	"github.com/guidomantilla/yarumo/common/random"
	"github.com/guidomantilla/yarumo/common/types"
)

func key(method *Method) types.Bytes {
	return random.Key(method.keySize)
}

func encrypt(method *Method, key types.Bytes, data types.Bytes, aad types.Bytes) (types.Bytes, error) {
	if method == nil {
		return nil, ErrMethodInvalid
	}
	if len(key) != method.keySize {
		return nil, ErrKeyInvalid
	}

	aead, err := method.kind(key, method.nonceSize)
	if err != nil {
		return nil, errs.Wrap(ErrCipherInitFailed, err)
	}

	nonce := random.Key(method.nonceSize)
	ciphered := aead.Seal(nil, nonce, data, aad)

	out := make([]byte, 0, len(nonce)+len(ciphered))
	out = append(out, nonce...)
	out = append(out, ciphered...)

	return out, nil
}

func decrypt(method *Method, key types.Bytes, ciphered types.Bytes, aad types.Bytes) (types.Bytes, error) {
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
		return nil, errs.Wrap(ErrCipherInitFailed, err)
	}

	nonce := ciphered[:method.nonceSize]
	ciphered = ciphered[method.nonceSize:]

	out, err := aead.Open(nil, nonce, ciphered, aad)
	if err != nil {
		return nil, errs.Wrap(ErrDecryptFailed, err)
	}

	return out, nil
}
