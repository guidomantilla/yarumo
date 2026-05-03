package hmacs

import (
	"crypto"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	ctypes "github.com/guidomantilla/yarumo/common/types"
)

// Predefined HMAC methods registered at package init.
var (
	HMAC_with_SHA256 = NewMethod("HMAC_with_SHA256", crypto.SHA256, 32)
	HMAC_with_SHA512 = NewMethod("HMAC_with_SHA512", crypto.SHA512, 64)
)

// Method holds the configuration for an HMAC algorithm.
type Method struct {
	name       string
	kind       crypto.Hash
	keySize    int
	keyFn      KeyFn
	digestFn   DigestFn
	validateFn ValidateFn
}

// NewMethod creates a new HMAC Method with the given name, hash algorithm, and key size.
func NewMethod(name string, kind crypto.Hash, keySize int, options ...Option) *Method {
	cassert.NotEmpty(name, "name is empty")

	opts := NewOptions(options...)

	return &Method{
		name:       name,
		kind:       kind,
		keySize:    keySize,
		keyFn:      opts.keyFn,
		digestFn:   opts.digestFn,
		validateFn: opts.validateFn,
	}
}

// Name returns the method's algorithm name.
func (m *Method) Name() string {
	cassert.NotNil(m, "method is nil")

	return m.name
}

// GenerateKey generates a new random symmetric key for this HMAC method.
func (m *Method) GenerateKey() (ctypes.Bytes, error) {
	cassert.NotNil(m, "method is nil")
	cassert.NotNil(m.keyFn, "method keyFn is nil")

	key, err := m.keyFn(m)
	if err != nil {
		return nil, ErrKeyGeneration(err)
	}

	return key, nil
}

// Digest computes an HMAC authentication tag over the provided data.
func (m *Method) Digest(key ctypes.Bytes, data ctypes.Bytes) (ctypes.Bytes, error) {
	cassert.NotNil(m, "method is nil")
	cassert.NotNil(m.digestFn, "method digestFn is nil")

	digest, err := m.digestFn(m, key, data)
	if err != nil {
		return nil, ErrDigest(err)
	}

	return digest, nil
}

// Validate checks whether the provided HMAC digest matches the computed one.
func (m *Method) Validate(key ctypes.Bytes, digest ctypes.Bytes, data ctypes.Bytes) (bool, error) {
	cassert.NotNil(m, "method is nil")
	cassert.NotNil(m.validateFn, "method validateFn is nil")

	ok, err := m.validateFn(m, key, digest, data)
	if err != nil {
		return false, ErrValidation(err)
	}

	return ok, nil
}
