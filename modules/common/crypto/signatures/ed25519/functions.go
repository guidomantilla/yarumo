package ed25519

import (
	"crypto/ed25519"
	"crypto/rand"

	"github.com/guidomantilla/yarumo/common/types"
)

func key() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	return ed25519.GenerateKey(rand.Reader)
}

func sign(method *Method, key *ed25519.PrivateKey, data types.Bytes) (types.Bytes, error) {
	if method == nil {
		return nil, ErrMethodIsNil
	}
	if key == nil {
		return nil, ErrKeyIsNil
	}
	if len(*key) != ed25519.PrivateKeySize {
		return nil, ErrKeyLengthIsInvalid
	}

	out := ed25519.Sign(*key, data)
	return out, nil
}

func verify(method *Method, key *ed25519.PublicKey, signature, data types.Bytes) (bool, error) {
	if method == nil {
		return false, ErrMethodIsNil
	}
	if key == nil {
		return false, ErrKeyIsNil
	}
	if len(*key) != ed25519.PrivateKeySize {
		return false, ErrKeyLengthIsInvalid
	}
	if len(signature) != ed25519.SignatureSize {
		return false, ErrSignatureLengthInvalid
	}

	ok := ed25519.Verify(*key, data, signature)
	return ok, nil
}
