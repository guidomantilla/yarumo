package ecdsas

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/asn1"
	"math/big"

	"github.com/guidomantilla/yarumo/common/errs"
	"github.com/guidomantilla/yarumo/common/types"
	"github.com/guidomantilla/yarumo/security/hashes"
)

type Format int

const (
	// RS Format => r || s
	//
	// Used in: JOSE / JWT / WebAuthn
	RS Format = iota

	// ASN1DER Format => SEQUENCE { r INTEGER, s INTEGER }
	//
	// Used in: X.509 / OpenSSL
	ASN1DER
)

// Sign signs the data using the given ECDSA-based method and key.
//
// For format RS:
//
//	Signature format: sig = r || s, where r and s are big-endian integers padded
//	with leading zeros to keySize bytes each; len(sig) == 2*keySize.
//
// For format ASN1DER:
//
//	Signature format: SEQUENCE { r INTEGER, s INTEGER } encoded with ASN.1 DER.
func Sign(method *Method, key *ecdsa.PrivateKey, data types.Bytes, format Format) (types.Bytes, error) {
	if method == nil {
		return nil, ErrMethodInvalid
	}
	if key == nil {
		return nil, ErrKeyInvalid
	}
	if key.Curve != method.curve {
		return nil, ErrKeyInvalid
	}

	h := hashes.Hash(method.kind, data)
	r, s, err := ecdsa.Sign(rand.Reader, key, h)
	if err != nil {
		return nil, errs.Wrap(ErrSignFailed, err)
	}

	switch format {
	case RS:
		// We serialize the outputs (r and s) into big-endian byte arrays padded with zeros on the left to make sure the sizes work out.
		// Output must be 2*keyBytes long.
		keyBytes := method.keySize
		out := make([]byte, 2*keyBytes)
		r.FillBytes(out[0:keyBytes]) // r is assigned to the first half of output.
		s.FillBytes(out[keyBytes:])  // s is assigned to the second half of output.
		return out, nil
	case ASN1DER:
		out, err := asn1.Marshal(tuple{R: r, S: s})
		if err != nil {
			return nil, errs.Wrap(ErrSignFailed, err)
		}
		return out, nil
	default:
		return nil, ErrFormatUnsupported
	}
}

func Verify(method *Method, key *ecdsa.PublicKey, signature types.Bytes, data types.Bytes, format Format) (bool, error) {
	if method == nil {
		return false, ErrMethodInvalid
	}
	if key == nil {
		return false, ErrKeyInvalid
	}
	if key.Curve != method.curve {
		return false, ErrKeyInvalid
	}

	var r, s *big.Int
	switch format {
	case RS:
		keyBytes := method.keySize
		if len(signature) != 2*keyBytes {
			return false, ErrSignatureInvalid
		}
		r = new(big.Int).SetBytes(signature[0:keyBytes])
		s = new(big.Int).SetBytes(signature[keyBytes:])
	case ASN1DER:
		var sig tuple
		_, err := asn1.Unmarshal(signature, &sig)
		if err != nil {
			return false, ErrSignatureInvalid
		}
		if sig.R == nil || sig.S == nil {
			return false, ErrSignatureInvalid
		}
		r, s = sig.R, sig.S
	default:
		return false, ErrFormatUnsupported
	}

	h := hashes.Hash(method.kind, data)
	ok := ecdsa.Verify(key, h, r, s)
	return ok, nil
}

type tuple struct {
	R, S *big.Int
}
