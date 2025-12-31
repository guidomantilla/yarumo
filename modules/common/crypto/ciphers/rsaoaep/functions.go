package rsaoaep

import (
	"crypto/rand"
	"crypto/rsa"

	"github.com/guidomantilla/yarumo/common/errs"
	"github.com/guidomantilla/yarumo/common/types"
	"github.com/guidomantilla/yarumo/common/utils"
)

// key RSA-OAEP
func key(method *Method, bits int) (*rsa.PrivateKey, error) {
	if method == nil {
		return nil, ErrMethodIsNil
	}

	if utils.NotIn(bits, method.allowedKeySizes...) {
		return nil, ErrKeyLengthIsInvalid
	}

	return rsa.GenerateKey(rand.Reader, bits)
}

// encrypt RSA-OAEP
func encrypt(method *Method, key *rsa.PublicKey, data types.Bytes, label types.Bytes) (types.Bytes, error) {
	if method == nil {
		return nil, ErrMethodIsNil
	}

	if key == nil {
		return nil, ErrKeyIsNil
	}

	if !method.kind.Available() {
		return nil, ErrHashNotAvailable
	}

	if utils.NotIn(key.N.BitLen(), method.allowedKeySizes...) {
		return nil, ErrKeyLengthIsInvalid
	}

	h := method.kind.New()

	out, err := rsa.EncryptOAEP(h, rand.Reader, key, data, label)
	if err != nil {
		return nil, errs.Wrap(ErrEncryptFailed, err)
	}

	return out, nil
}

// decrypt RSA-OAEP
func decrypt(method *Method, priv *rsa.PrivateKey, ciphered types.Bytes, label types.Bytes) (types.Bytes, error) {
	if method == nil {
		return nil, ErrMethodIsNil
	}

	if priv == nil {
		return nil, ErrKeyIsNil
	}

	if !method.kind.Available() {
		return nil, ErrHashNotAvailable
	}

	if utils.NotIn(priv.N.BitLen(), method.allowedKeySizes...) {
		return nil, ErrKeyLengthIsInvalid
	}

	h := method.kind.New()

	out, err := rsa.DecryptOAEP(h, rand.Reader, priv, ciphered, label)
	if err != nil {
		return nil, errs.Wrap(ErrDecryptFailed, err)
	}

	return out, nil
}
