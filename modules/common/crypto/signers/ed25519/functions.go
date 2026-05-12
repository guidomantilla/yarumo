package ed25519

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
	ctypes "github.com/guidomantilla/yarumo/common/types"
)

// pemBlockPrivateKey is the PEM block type used for PKCS#8 private keys.
const pemBlockPrivateKey = "PRIVATE KEY"

// pemBlockPublicKey is the PEM block type used for PKIX/SubjectPublicKeyInfo public keys.
const pemBlockPublicKey = "PUBLIC KEY"

// MarshalPrivateKeyPEM marshals an Ed25519 private key into PKCS#8 PEM-encoded bytes.
func MarshalPrivateKeyPEM(key ed25519.PrivateKey) ([]byte, error) {
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

// ParsePrivateKeyPEM parses an Ed25519 private key from PKCS#8 PEM-encoded bytes.
func ParsePrivateKeyPEM(pemBytes []byte) (ed25519.PrivateKey, error) {
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

	key, ok := parsed.(ed25519.PrivateKey)
	if !ok {
		return nil, ErrPEMCodec(ErrKeyTypeMismatch)
	}

	return key, nil
}

// MarshalPublicKeyPEM marshals an Ed25519 public key into PKIX/SubjectPublicKeyInfo PEM-encoded bytes.
func MarshalPublicKeyPEM(key ed25519.PublicKey) ([]byte, error) {
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

// ParsePublicKeyPEM parses an Ed25519 public key from PKIX/SubjectPublicKeyInfo PEM-encoded bytes.
func ParsePublicKeyPEM(pemBytes []byte) (ed25519.PublicKey, error) {
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

	key, ok := parsed.(ed25519.PublicKey)
	if !ok {
		return nil, ErrPEMCodec(ErrKeyTypeMismatch)
	}

	return key, nil
}

func key() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	return ed25519.GenerateKey(rand.Reader)
}

func sign(method *Method, key *ed25519.PrivateKey, data ctypes.Bytes) (ctypes.Bytes, error) {
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

func verify(method *Method, key *ed25519.PublicKey, signature, data ctypes.Bytes) (bool, error) {
	if method == nil {
		return false, ErrMethodIsNil
	}

	if key == nil {
		return false, ErrKeyIsNil
	}

	if len(*key) != ed25519.PublicKeySize {
		return false, ErrKeyLengthIsInvalid
	}

	if len(signature) != ed25519.SignatureSize {
		return false, ErrSignatureLengthInvalid
	}

	ok := ed25519.Verify(*key, data, signature)

	return ok, nil
}
