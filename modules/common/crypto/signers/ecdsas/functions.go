package ecdsas

import (
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"

	chashes "github.com/guidomantilla/yarumo/common/crypto/hashes"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
	ctypes "github.com/guidomantilla/yarumo/common/types"
)

func key(method *Method) (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(method.curve, rand.Reader)
}

func sign(method *Method, key *ecdsa.PrivateKey, data ctypes.Bytes, format Format) (ctypes.Bytes, error) {
	if method == nil {
		return nil, ErrMethodIsNil
	}

	if key == nil {
		return nil, ErrKeyIsNil
	}

	if key.Curve != method.curve {
		return nil, ErrKeyCurveIsInvalid
	}

	h := chashes.Hash(method.kind, data)

	switch format {
	case RS:
		r, s, err := ecdsa.Sign(rand.Reader, key, h)
		if err != nil {
			return nil, cerrs.Wrap(ErrSignFailed, err)
		}
		// Serialize r and s into big-endian byte arrays padded with zeros on the left.
		// Output must be 2*keyBytes long: first r, then s.
		keyBytes := method.keySize
		out := make([]byte, 2*keyBytes)
		r.FillBytes(out[0:keyBytes]) // r is assigned to the first half of output.
		s.FillBytes(out[keyBytes:])  // s is assigned to the second half of output.

		return out, nil

	case ASN1:
		out, err := ecdsa.SignASN1(rand.Reader, key, h)
		if err != nil {
			return nil, cerrs.Wrap(ErrSignFailed, err)
		}

		return out, nil
	}

	return nil, ErrFormatUnsupported
}

func verify(method *Method, key *ecdsa.PublicKey, signature ctypes.Bytes, data ctypes.Bytes, format Format) (bool, error) {
	if method == nil {
		return false, ErrMethodIsNil
	}

	if key == nil {
		return false, ErrKeyIsNil
	}

	if key.Curve != method.curve {
		return false, ErrKeyCurveIsInvalid
	}

	h := chashes.Hash(method.kind, data)

	switch format {
	case RS:
		keyBytes := method.keySize
		if len(signature) != 2*keyBytes {
			return false, ErrSignatureInvalid
		}

		r := new(big.Int).SetBytes(signature[0:keyBytes])
		s := new(big.Int).SetBytes(signature[keyBytes:])
		ok := ecdsa.Verify(key, h, r, s)

		return ok, nil
	case ASN1:
		ok := ecdsa.VerifyASN1(key, h, signature)
		return ok, nil
	}

	return false, ErrFormatUnsupported
}
