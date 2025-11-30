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

	ASN1
)

// Sign generates an ECDSA signature over the given data using the provided
// method and private key.
//
// Parameters:
//   - method: cryptographic method defining the hash function, curve, and key size.
//   - key: ECDSA private key used to produce the signature.
//   - data: message to be signed.
//   - format: output signature format (RS or ASN1DER).
//
// Behavior:
//   - Returns an error if method or key are nil.
//   - Returns an error if the key's curve does not match the method's curve.
//   - Computes the hash of data using the method's hash function.
//   - Invokes ecdsa.Sign to produce (r, s).
//   - RS format: serializes r||s as fixed-size big-endian byte slices (2*keySize).
//   - ASN1DER format: encodes r and s into a standard ASN.1 DER structure.
//
// Returns:
//   - A byte slice containing the encoded signature.
//   - An error if signing fails or if the format is not supported.
//
// Possible errors:
//   - ErrMethodInvalid
//   - ErrKeyInvalid
//   - ErrSignFailed
//   - ErrFormatUnsupported
//
// The function guarantees consistent output formatting and never panics.
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

	switch format {
	case RS:
		h := hashes.Hash(method.kind, data)
		r, s, err := ecdsa.Sign(rand.Reader, key, h)
		//ecdsa.SignASN1()
		if err != nil {
			return nil, errs.Wrap(ErrSignFailed, err)
		}
		// We serialize the outputs (r and s) into big-endian byte arrays padded with zeros on the left to make sure the sizes work out.
		// Output must be 2*keyBytes long.
		keyBytes := method.keySize
		out := make([]byte, 2*keyBytes)
		r.FillBytes(out[0:keyBytes]) // r is assigned to the first half of output.
		s.FillBytes(out[keyBytes:])  // s is assigned to the second half of output.
		return out, nil
		
	case ASN1DER:
		h := hashes.Hash(method.kind, data)
		r, s, err := ecdsa.Sign(rand.Reader, key, h)
		if err != nil {
			return nil, errs.Wrap(ErrSignFailed, err)
		}
		out, err := asn1.Marshal(tuple{R: r, S: s})
		if err != nil {
			return nil, errs.Wrap(ErrSignFailed, err)
		}
		return out, nil

	case ASN1:
		out, err := ecdsa.SignASN1(rand.Reader, key, data)
		if err != nil {
			return nil, errs.Wrap(ErrSignFailed, err)
		}
		return out, nil
	}

	return nil, ErrFormatUnsupported
}

// Verify checks an ECDSA signature over the given data using the provided
// method and public key.
//
// Parameters:
//   - method: cryptographic method defining the hash function and curve.
//   - key: ECDSA public key used for verification.
//   - signature: the signature to verify (in RS or ASN1DER format).
//   - data: the original message that was signed.
//   - format: signature format (RS or ASN1DER).
//
// Behavior:
//   - Returns an error if method or key are nil.
//   - Returns an error if the key's curve does not match the method's curve.
//   - RS format: splits the signature into r||s using method.keySize.
//   - ASN1DER format: decodes a tuple ASN.1 structure containing R and S.
//   - Hashes the data using the method's hash and invokes ecdsa.Verify.
//
// Returns:
//   - (true, nil)  if the signature is valid.
//   - (false, nil) if the signature is invalid.
//   - (false, err) if the signature format is invalid or incompatible.
//
// Possible errors:
//   - ErrMethodInvalid
//   - ErrKeyInvalid
//   - ErrSignatureInvalid
//   - ErrFormatUnsupported
//
// The function never panics and does not return an error for a simple
// verification failure (it returns false, nil instead).
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
	//ecdsa.VerifyASN1()
	return ok, nil
}

type tuple struct {
	R, S *big.Int
}
