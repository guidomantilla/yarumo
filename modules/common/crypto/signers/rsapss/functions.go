package rsapss

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"

	chashes "github.com/guidomantilla/yarumo/common/crypto/hashes"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
	ctypes "github.com/guidomantilla/yarumo/common/types"
	cutils "github.com/guidomantilla/yarumo/common/utils"
)

func key(bits int) (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, bits)
}

func sign(method *Method, key *rsa.PrivateKey, data ctypes.Bytes) (ctypes.Bytes, error) {
	if method == nil {
		return nil, ErrMethodIsNil
	}

	if key == nil {
		return nil, ErrKeyIsNil
	}

	if cutils.NotIn(key.N.BitLen(), method.allowedKeySizes...) {
		return nil, ErrKeyLengthIsInvalid
	}

	h := chashes.Hash(method.kind, data)

	out, err := rsa.SignPSS(rand.Reader, key, method.kind, h, &rsa.PSSOptions{
		SaltLength: method.saltLength,
		Hash:       method.kind,
	})
	if err != nil {
		return nil, cerrs.Wrap(ErrSignFailed, err)
	}

	return out, nil
}

func verify(method *Method, key *rsa.PublicKey, signature ctypes.Bytes, data ctypes.Bytes) (bool, error) {
	if method == nil {
		return false, ErrMethodIsNil
	}

	if key == nil {
		return false, ErrKeyIsNil
	}

	if cutils.NotIn(key.N.BitLen(), method.allowedKeySizes...) {
		return false, ErrKeyLengthIsInvalid
	}

	h := chashes.Hash(method.kind, data)

	err := rsa.VerifyPSS(key, method.kind, h, signature, &rsa.PSSOptions{
		SaltLength: method.saltLength,
		Hash:       method.kind,
	})
	if err != nil {
		if errors.Is(err, rsa.ErrVerification) {
			return false, nil
		}

		return false, cerrs.Wrap(ErrVerifyFailed, err)
	}

	return true, nil
}
