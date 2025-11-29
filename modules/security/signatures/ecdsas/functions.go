package ecdsas

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"hash"

	"github.com/guidomantilla/yarumo/common/errs"
	"github.com/guidomantilla/yarumo/common/types"
	"github.com/guidomantilla/yarumo/security/hashes"
)

func Sign(method *Method, key *ecdsa.PrivateKey, data types.Bytes) (types.Bytes, error) {
	if method == nil {
		return nil, ErrMethodInvalid
	}
	if key == nil {
		return nil, ErrKeyInvalid
	}
	if key.Curve != method.curve {
		return nil, ErrKeyInvalid
	}
	//if len(data) == 0 {
	//	return nil, ErrDataEmpty
	//}

	h := hashes.Hash(method.kind, data)
	r, s, err := ecdsa.Sign(rand.Reader, key, h)
	if err != nil {
		return nil, errs.Wrap(ErrSignFailed, err)
	}

	// We serialize the outputs (r and s) into big-endian byte arrays padded with zeros on the left to make sure the sizes work out.
	// Output must be 2*keyBytes long.
	var keyBytes = method.keySize
	out := make([]byte, 2*keyBytes)

	r.FillBytes(out[0:keyBytes]) // r is assigned to the first half of output.
	s.FillBytes(out[keyBytes:])  // s is assigned to the second half of output.
	return out, nil
}

// internal use only

func sha_new256() hash.Hash {
	return sha256.New()
}

func sha_new512() hash.Hash {
	return sha512.New()
}
