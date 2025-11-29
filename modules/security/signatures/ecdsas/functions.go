package ecdsas

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"

	"github.com/guidomantilla/yarumo/common/errs"
	"github.com/guidomantilla/yarumo/common/types"
	"github.com/guidomantilla/yarumo/security/hashes"
)

func Key(name Name) (*ecdsa.PrivateKey, error) {
	if name == ES256 {
		return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	}
	return nil, nil
}

// ES256 - 32, 256
func ECDSA_P256_SHA256(key *ecdsa.PrivateKey, data types.Bytes) (types.Bytes, error) {
	if key == nil {
		return nil, ErrKeyInvalid
	}
	if key.Curve != elliptic.P256() {
		return nil, ErrKeyInvalid
	}
	if len(data) == 0 {
		return nil, ErrDataEmpty
	}

	hash := hashes.SHA256.Hash(data)
	r, s, err := ecdsa.Sign(rand.Reader, key, hash)
	if err != nil {
		return nil, errs.Wrap(ErrSignFailed, err)
	}

	// We serialize the outputs (r and s) into big-endian byte arrays padded with zeros on the left to make sure the sizes work out.
	// Output must be 2*keyBytes long.
	const keyBytes = 32
	out := make([]byte, 2*keyBytes)

	r.FillBytes(out[0:keyBytes]) // r is assigned to the first half of output.
	s.FillBytes(out[keyBytes:])  // s is assigned to the second half of output.
	return out, nil
}

// ES512 - 66, 521
func ECDSA_P521_SHA512(key *ecdsa.PrivateKey, data types.Bytes) (types.Bytes, error) {
	if key == nil {
		return nil, ErrKeyInvalid
	}
	if key.Curve != elliptic.P521() {
		return nil, ErrKeyInvalid
	}
	if len(data) == 0 {
		return nil, ErrDataEmpty
	}

	hash := hashes.SHA3_512.Hash(data)
	r, s, err := ecdsa.Sign(rand.Reader, key, hash)
	if err != nil {
		return nil, errs.Wrap(ErrSignFailed, err)
	}

	// We serialize the outputs (r and s) into big-endian byte arrays padded with zeros on the left to make sure the sizes work out.
	// Output must be 2*keyBytes long.
	const keyBytes = 66
	out := make([]byte, 2*keyBytes)

	r.FillBytes(out[0:keyBytes]) // r is assigned to the first half of output.
	s.FillBytes(out[keyBytes:])  // s is assigned to the second half of output.
	return out, nil
}
