package ecdsas

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"math/big"

	chashes "github.com/guidomantilla/yarumo/core/crypto/hashes"
	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
	ctypes "github.com/guidomantilla/yarumo/core/common/types"
)

// pemBlockPrivateKey is the PEM block type used for PKCS#8 private keys.
const pemBlockPrivateKey = "PRIVATE KEY"

// pemBlockPublicKey is the PEM block type used for PKIX/SubjectPublicKeyInfo public keys.
const pemBlockPublicKey = "PUBLIC KEY"

// Digest is the recommended entry point for callers that receive the
// algorithm name as a string (e.g. loaded from config, a request header, or
// a database column). It performs a single registry Get, parses the PKCS#8
// PEM-encoded private key, and forwards to Method.Sign using the ASN.1 DER
// signature format (the standard library default).
//
// Use the explicit Method.Sign API when an alternative signature format
// (e.g. RS for JOSE/JWT or WebAuthn) is required.
func Digest(name string, key, data ctypes.Bytes) (ctypes.Bytes, error) {
	method, err := Get(name)
	if err != nil {
		return nil, err
	}

	priv, err := ParsePrivateKeyPEM(key)
	if err != nil {
		return nil, err
	}

	return method.Sign(priv, data, ASN1)
}

// Validate is the recommended entry point for callers that receive the
// algorithm name as a string. It performs a single registry Get, parses the
// PKIX PEM-encoded public key, and forwards to Method.Verify using the
// ASN.1 DER signature format. Use Method.Verify directly when a different
// signature format is required.
func Validate(name string, key, digest, data ctypes.Bytes) (bool, error) {
	method, err := Get(name)
	if err != nil {
		return false, err
	}

	pub, err := ParsePublicKeyPEM(key)
	if err != nil {
		return false, err
	}

	return method.Verify(pub, digest, data, ASN1)
}

// MarshalPrivateKeyPEM marshals an ECDSA private key into PKCS#8 PEM-encoded bytes.
func MarshalPrivateKeyPEM(key *ecdsa.PrivateKey) ([]byte, error) {
	if key == nil {
		return nil, ErrPEMCodec(ErrKeyIsNil)
	}

	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, ErrPEMCodec(cerrs.Wrap(ErrMarshalKeyFailed, err))
	}

	out := pem.EncodeToMemory(&pem.Block{
		Type:  pemBlockPrivateKey,
		Bytes: der,
	})

	return out, nil
}

// ParsePrivateKeyPEM parses an ECDSA private key from PKCS#8 PEM-encoded bytes.
func ParsePrivateKeyPEM(pemBytes []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, ErrPEMCodec(ErrPEMDecodeFailed)
	}

	if block.Type != pemBlockPrivateKey {
		return nil, ErrPEMCodec(ErrPEMBlockTypeMismatch)
	}

	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, ErrPEMCodec(cerrs.Wrap(ErrParseKeyFailed, err))
	}

	key, ok := parsed.(*ecdsa.PrivateKey)
	if !ok {
		return nil, ErrPEMCodec(ErrKeyTypeMismatch)
	}

	return key, nil
}

// MarshalPublicKeyPEM marshals an ECDSA public key into PKIX/SubjectPublicKeyInfo PEM-encoded bytes.
func MarshalPublicKeyPEM(key *ecdsa.PublicKey) ([]byte, error) {
	if key == nil {
		return nil, ErrPEMCodec(ErrKeyIsNil)
	}

	der, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return nil, ErrPEMCodec(cerrs.Wrap(ErrMarshalKeyFailed, err))
	}

	out := pem.EncodeToMemory(&pem.Block{
		Type:  pemBlockPublicKey,
		Bytes: der,
	})

	return out, nil
}

// ParsePublicKeyPEM parses an ECDSA public key from PKIX/SubjectPublicKeyInfo PEM-encoded bytes.
func ParsePublicKeyPEM(pemBytes []byte) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, ErrPEMCodec(ErrPEMDecodeFailed)
	}

	if block.Type != pemBlockPublicKey {
		return nil, ErrPEMCodec(ErrPEMBlockTypeMismatch)
	}

	parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, ErrPEMCodec(cerrs.Wrap(ErrParseKeyFailed, err))
	}

	key, ok := parsed.(*ecdsa.PublicKey)
	if !ok {
		return nil, ErrPEMCodec(ErrKeyTypeMismatch)
	}

	return key, nil
}

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

	h, err := chashes.Hash(method.kind, data)
	if err != nil {
		return nil, ErrSigning(err)
	}

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

	h, err := chashes.Hash(method.kind, data)
	if err != nil {
		return false, ErrVerification(err)
	}

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
