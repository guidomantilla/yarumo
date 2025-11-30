package rsapss

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"

	"github.com/guidomantilla/yarumo/common/crypto/hashes"
	"github.com/guidomantilla/yarumo/common/errs"
	"github.com/guidomantilla/yarumo/common/types"
	"github.com/guidomantilla/yarumo/common/utils"
)

func key(bits int) (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, bits)
}

func sign(method *Method, key *rsa.PrivateKey, data types.Bytes) (types.Bytes, error) {
	if method == nil {
		return nil, ErrMethodIsNil
	}
	if key == nil {
		return nil, ErrKeyIsNil
	}
	if utils.NotIn(key.N.BitLen(), method.allowedKeySizes...) {
		return nil, ErrKeyLengthIsInvalid
	}

	h := hashes.Hash(method.kind, data)
	out, err := rsa.SignPSS(rand.Reader, key, method.kind, h, &rsa.PSSOptions{
		SaltLength: method.saltLength,
		Hash:       method.kind,
	})
	if err != nil {
		return nil, errs.Wrap(ErrSignFailed, err)
	}
	return out, nil
}

func verify(method *Method, key *rsa.PublicKey, signature types.Bytes, data types.Bytes) (bool, error) {
	if method == nil {
		return false, ErrMethodIsNil
	}
	if key == nil {
		return false, ErrKeyIsNil
	}
	if utils.NotIn(key.N.BitLen(), method.allowedKeySizes...) {
		return false, ErrKeyLengthIsInvalid
	}

	h := hashes.Hash(method.kind, data)
	err := rsa.VerifyPSS(key, method.kind, h, signature, &rsa.PSSOptions{
		SaltLength: method.saltLength,
		Hash:       method.kind,
	})
	if err != nil {
		if errors.Is(err, rsa.ErrVerification) {
			return false, nil
		}
		return false, errs.Wrap(ErrVerifyFailed, err)
	}

	return true, nil
}
