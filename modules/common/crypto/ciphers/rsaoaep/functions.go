package rsaoaep

import (
	"crypto/rand"
	"crypto/rsa"
	_ "crypto/sha256"
	_ "crypto/sha512"
	"crypto/x509"
	"encoding/pem"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
	ctypes "github.com/guidomantilla/yarumo/common/types"
	cutils "github.com/guidomantilla/yarumo/common/utils"
)

// pemBlockPrivateKey is the PEM block type used for PKCS#8 private keys.
const pemBlockPrivateKey = "PRIVATE KEY"

// pemBlockPublicKey is the PEM block type used for PKIX/SubjectPublicKeyInfo public keys.
const pemBlockPublicKey = "PUBLIC KEY"

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

func key(method *Method, bits int) (*rsa.PrivateKey, error) {
	if method == nil {
		return nil, ErrMethodIsNil
	}

	if cutils.NotIn(bits, method.allowedKeySizes...) {
		return nil, ErrKeyLengthIsInvalid
	}

	return rsa.GenerateKey(rand.Reader, bits)
}

func encrypt(method *Method, key *rsa.PublicKey, data ctypes.Bytes, label ctypes.Bytes) (ctypes.Bytes, error) {
	if method == nil {
		return nil, ErrMethodIsNil
	}

	if key == nil {
		return nil, ErrKeyIsNil
	}

	if !method.kind.Available() {
		return nil, ErrHashNotAvailable
	}

	if cutils.NotIn(key.N.BitLen(), method.allowedKeySizes...) {
		return nil, ErrKeyLengthIsInvalid
	}

	h := method.kind.New()

	out, err := rsa.EncryptOAEP(h, rand.Reader, key, data, label)
	if err != nil {
		return nil, cerrs.Wrap(ErrEncryptionFailed, err)
	}

	return out, nil
}

func decrypt(method *Method, priv *rsa.PrivateKey, ciphered ctypes.Bytes, label ctypes.Bytes) (ctypes.Bytes, error) {
	if method == nil {
		return nil, ErrMethodIsNil
	}

	if priv == nil {
		return nil, ErrKeyIsNil
	}

	if !method.kind.Available() {
		return nil, ErrHashNotAvailable
	}

	if cutils.NotIn(priv.N.BitLen(), method.allowedKeySizes...) {
		return nil, ErrKeyLengthIsInvalid
	}

	h := method.kind.New()

	out, err := rsa.DecryptOAEP(h, rand.Reader, priv, ciphered, label)
	if err != nil {
		return nil, cerrs.Wrap(ErrDecryptionFailed, err)
	}

	return out, nil
}
