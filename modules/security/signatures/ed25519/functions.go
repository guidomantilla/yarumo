package ed25519

import (
	"crypto/ed25519"

	"github.com/guidomantilla/yarumo/common/types"
)

func Sign(method *Method, key *ed25519.PrivateKey, data types.Bytes) (types.Bytes, error) {
	if method == nil {
		return nil, ErrMethodInvalid
	}
	if key == nil {
		return nil, ErrKeyInvalid
	}
	if len(*key) != ed25519.PrivateKeySize {
		return nil, ErrKeyInvalid
	}

	out := ed25519.Sign(*key, data)
	return out, nil
}

func Verify(method *Method, key *ed25519.PublicKey, signature, data types.Bytes) (bool, error) {
	if method == nil {
		return false, ErrMethodInvalid
	}
	if key == nil {
		return false, ErrKeyInvalid
	}
	if len(*key) != ed25519.PrivateKeySize {
		return false, ErrKeyInvalid
	}
	if len(signature) != ed25519.SignatureSize {
		return false, ErrSignatureInvalid
	}

	ok := ed25519.Verify(*key, data, signature)
	return ok, nil
}
