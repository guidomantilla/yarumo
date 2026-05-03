package ed25519

import (
	"crypto/ed25519"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	ctypes "github.com/guidomantilla/yarumo/common/types"
)

// Predefined Ed25519 method registered at package init.
var (
	Ed25519 = NewMethod("Ed25519")
)

// Method holds the configuration for an Ed25519 algorithm.
type Method struct {
	name     string
	keyFn    KeyFn
	signFn   SignFn
	verifyFn VerifyFn
}

// NewMethod creates a new Ed25519 Method with the given name.
func NewMethod(name string, options ...Option) *Method {
	cassert.NotEmpty(name, "name is empty")

	opts := NewOptions(options...)

	return &Method{
		name:     name,
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

// GenerateKey generates a new Ed25519 private key.
func (m *Method) GenerateKey() (ed25519.PrivateKey, error) {
	cassert.NotNil(m, "method is nil")
	cassert.NotNil(m.keyFn, "method keyFn is nil")

	_, key, err := m.keyFn()
	if err != nil {
		return nil, ErrKeyGeneration(err)
	}

	return key, nil
}

// Sign produces an Ed25519 signature over the provided data.
func (m *Method) Sign(key *ed25519.PrivateKey, data ctypes.Bytes) (ctypes.Bytes, error) {
	cassert.NotNil(m, "method is nil")
	cassert.NotNil(m.signFn, "method signFn is nil")

	signature, err := m.signFn(m, key, data)
	if err != nil {
		return nil, ErrSigning(err)
	}

	return signature, nil
}

// Verify checks an Ed25519 signature over the given data.
func (m *Method) Verify(key *ed25519.PublicKey, signature, data ctypes.Bytes) (bool, error) {
	cassert.NotNil(m, "method is nil")
	cassert.NotNil(m.verifyFn, "method verifyFn is nil")

	ok, err := m.verifyFn(m, key, signature, data)
	if err != nil {
		return false, ErrVerification(err)
	}

	return ok, nil
}
