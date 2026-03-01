package ecdsas

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	ctypes "github.com/guidomantilla/yarumo/common/types"
)

// Predefined ECDSA methods registered at package init.
var (
	ECDSA_with_SHA256_over_P256 = NewMethod("ECDSA_with_SHA256_over_P256", crypto.SHA256, 32, elliptic.P256())
	ECDSA_with_SHA512_over_P521 = NewMethod("ECDSA_with_SHA512_over_P521", crypto.SHA512, 66, elliptic.P521())
)

// Method holds the configuration for an ECDSA algorithm.
type Method struct {
	name     string
	kind     crypto.Hash
	keySize  int
	curve    elliptic.Curve
	keyFn    KeyFn
	signFn   SignFn
	verifyFn VerifyFn
}

// NewMethod creates a new ECDSA Method with the given name, hash, key size, and curve.
func NewMethod(name string, kind crypto.Hash, keySize int, curve elliptic.Curve, options ...Option) *Method {
	cassert.NotEmpty(name, "name is empty")

	opts := NewOptions(options...)

	return &Method{
		name:     name,
		kind:     kind,
		keySize:  keySize,
		curve:    curve,
		keyFn:    opts.keyFn,
		signFn:   opts.signFn,
		verifyFn: opts.verifyFn,
	}
}

// Name returns the method's algorithm name.
func (m *Method) Name() string {
	cassert.NotNil(m, "method is nil")

	return m.name
}

// GenerateKey generates a new ECDSA private key for this method's curve.
func (m *Method) GenerateKey() (*ecdsa.PrivateKey, error) {
	cassert.NotNil(m, "method is nil")
	cassert.NotNil(m.keyFn, "method keyFn is nil")

	key, err := m.keyFn(m)
	if err != nil {
		return nil, ErrKeyGeneration(err)
	}

	return key, nil
}

// Sign generates an ECDSA signature over the given data.
func (m *Method) Sign(key *ecdsa.PrivateKey, data ctypes.Bytes, format Format) (ctypes.Bytes, error) {
	cassert.NotNil(m, "method is nil")
	cassert.NotNil(m.signFn, "method signFn is nil")

	signature, err := m.signFn(m, key, data, format)
	if err != nil {
		return nil, ErrSigning(err)
	}

	return signature, nil
}

// Verify checks an ECDSA signature over the given data.
func (m *Method) Verify(key *ecdsa.PublicKey, signature ctypes.Bytes, data ctypes.Bytes, format Format) (bool, error) {
	cassert.NotNil(m, "method is nil")
	cassert.NotNil(m.verifyFn, "method verifyFn is nil")

	ok, err := m.verifyFn(m, key, signature, data, format)
	if err != nil {
		return false, ErrVerification(err)
	}

	return ok, nil
}
