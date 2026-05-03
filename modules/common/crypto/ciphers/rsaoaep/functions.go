package rsaoaep

import (
	"crypto/rand"
	"crypto/rsa"
	_ "crypto/sha256"
	_ "crypto/sha512"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
	ctypes "github.com/guidomantilla/yarumo/common/types"
	cutils "github.com/guidomantilla/yarumo/common/utils"
)

func key(method *Method, bits int) (*rsa.PrivateKey, error) {
	if method == nil {
		return nil, ErrMethodIsNil
	}

	if cutils.NotIn(bits, method.allowedKeySizes...) {
		return nil, ErrKeyLengthIsInvalid
	}

	return rsa.GenerateKey(rand.Reader, bits)
}

func encrypt(method *Method, key *rsa.PublicKey, data ctypes.Bytes, label ctypes.Bytes) (ctypes.Bytes, error) {
	if method == nil {
		return nil, ErrMethodIsNil
	}

	if key == nil {
		return nil, ErrKeyIsNil
	}

	if !method.kind.Available() {
		return nil, ErrHashNotAvailable
	}

	if cutils.NotIn(key.N.BitLen(), method.allowedKeySizes...) {
		return nil, ErrKeyLengthIsInvalid
	}

	h := method.kind.New()

	out, err := rsa.EncryptOAEP(h, rand.Reader, key, data, label)
	if err != nil {
		return nil, cerrs.Wrap(ErrEncryptionFailed, err)
	}

	return out, nil
}

func decrypt(method *Method, priv *rsa.PrivateKey, ciphered ctypes.Bytes, label ctypes.Bytes) (ctypes.Bytes, error) {
	if method == nil {
		return nil, ErrMethodIsNil
	}

	if priv == nil {
		return nil, ErrKeyIsNil
	}

	if !method.kind.Available() {
		return nil, ErrHashNotAvailable
	}

	if cutils.NotIn(priv.N.BitLen(), method.allowedKeySizes...) {
		return nil, ErrKeyLengthIsInvalid
	}

	h := method.kind.New()

	out, err := rsa.DecryptOAEP(h, rand.Reader, priv, ciphered, label)
	if err != nil {
		return nil, cerrs.Wrap(ErrDecryptionFailed, err)
	}

	return out, nil
}
