package rsassas

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"

	chashes "github.com/guidomantilla/yarumo/common/crypto/hashes"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
	ctypes "github.com/guidomantilla/yarumo/common/types"
	cutils "github.com/guidomantilla/yarumo/common/utils"
)

// pemBlockPrivateKey is the PEM block type used for PKCS#8 private keys.
const pemBlockPrivateKey = "PRIVATE KEY"

// pemBlockPublicKey is the PEM block type used for PKIX/SubjectPublicKeyInfo public keys.
const pemBlockPublicKey = "PUBLIC KEY"

// Digest is the recommended entry point for callers that receive the
// algorithm name as a string (e.g. loaded from config, a request header, or
// a database column). It performs a single registry Get, parses the PKCS#8
// PEM-encoded RSA private key, and forwards to Method.Sign.
func Digest(name string, key, data ctypes.Bytes) (ctypes.Bytes, error) {
	method, err := Get(name)
	if err != nil {
		return nil, err
	}

	priv, err := ParsePrivateKeyPEM(key)
	if err != nil {
		return nil, err
	}

	return method.Sign(priv, data)
}

// Validate is the recommended entry point for callers that receive the
// algorithm name as a string. It performs a single registry Get, parses the
// PKIX PEM-encoded RSA public key, and forwards to Method.Verify.
func Validate(name string, key, digest, data ctypes.Bytes) (bool, error) {
	method, err := Get(name)
	if err != nil {
		return false, err
	}

	pub, err := ParsePublicKeyPEM(key)
	if err != nil {
		return false, err
	}

	return method.Verify(pub, digest, data)
}

// MarshalPrivateKeyPEM marshals an RSA private key into PKCS#8 PEM-encoded bytes.
func MarshalPrivateKeyPEM(key *rsa.PrivateKey) ([]byte, error) {
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

// ParsePrivateKeyPEM parses an RSA private key from PKCS#8 PEM-encoded bytes.
func ParsePrivateKeyPEM(pemBytes []byte) (*rsa.PrivateKey, error) {
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

	key, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, ErrPEMCodec(ErrKeyTypeMismatch)
	}

	return key, nil
}

// MarshalPublicKeyPEM marshals an RSA public key into PKIX/SubjectPublicKeyInfo PEM-encoded bytes.
func MarshalPublicKeyPEM(key *rsa.PublicKey) ([]byte, error) {
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

// ParsePublicKeyPEM parses an RSA public key from PKIX/SubjectPublicKeyInfo PEM-encoded bytes.
func ParsePublicKeyPEM(pemBytes []byte) (*rsa.PublicKey, error) {
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

	key, ok := parsed.(*rsa.PublicKey)
	if !ok {
		return nil, ErrPEMCodec(ErrKeyTypeMismatch)
	}

	return key, nil
}

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

	h, err := chashes.Hash(method.kind, data)
	if err != nil {
		return nil, ErrSigning(err)
	}

	switch method.padding {
	case PSS:
		out, err := rsa.SignPSS(rand.Reader, key, method.kind, h, &rsa.PSSOptions{
			SaltLength: rsa.PSSSaltLengthEqualsHash,
			Hash:       method.kind,
		})
		if err != nil {
			return nil, cerrs.Wrap(ErrSignFailed, err)
		}

		return out, nil

	case PKCS1v15:
		out, err := rsa.SignPKCS1v15(rand.Reader, key, method.kind, h)
		if err != nil {
			return nil, cerrs.Wrap(ErrSignFailed, err)
		}

		return out, nil

	default:
		return nil, ErrPaddingNotSupported
	}
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

	h, err := chashes.Hash(method.kind, data)
	if err != nil {
		return false, ErrVerification(err)
	}

	switch method.padding {
	case PSS:
		err = rsa.VerifyPSS(key, method.kind, h, signature, &rsa.PSSOptions{
			SaltLength: rsa.PSSSaltLengthEqualsHash,
			Hash:       method.kind,
		})

	case PKCS1v15:
		err = rsa.VerifyPKCS1v15(key, method.kind, h, signature)

	default:
		return false, ErrPaddingNotSupported
	}

	if err != nil {
		if errors.Is(err, rsa.ErrVerification) {
			return false, nil
		}

		return false, cerrs.Wrap(ErrVerifyFailed, err)
	}

	return true, nil
}
