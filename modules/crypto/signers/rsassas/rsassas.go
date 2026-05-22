package rsassas

import (
	"crypto"
	"crypto/rsa"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	ctypes "github.com/guidomantilla/yarumo/common/types"
	cutils "github.com/guidomantilla/yarumo/common/utils"
)

// Predefined RSA signing methods registered at package init.
var (
	RSASSA_PSS_using_SHA256      = NewMethod("RSASSA_PSS_using_SHA256", crypto.SHA256, PSS, []int{2048, 3072, 4096})
	RSASSA_PSS_using_SHA384      = NewMethod("RSASSA_PSS_using_SHA384", crypto.SHA384, PSS, []int{2048, 3072, 4096})
	RSASSA_PSS_using_SHA512      = NewMethod("RSASSA_PSS_using_SHA512", crypto.SHA512, PSS, []int{3072, 4096})
	RSASSA_PKCS1v15_using_SHA256 = NewMethod("RSASSA_PKCS1v15_using_SHA256", crypto.SHA256, PKCS1v15, []int{2048, 3072, 4096})
	RSASSA_PKCS1v15_using_SHA384 = NewMethod("RSASSA_PKCS1v15_using_SHA384", crypto.SHA384, PKCS1v15, []int{2048, 3072, 4096})
	RSASSA_PKCS1v15_using_SHA512 = NewMethod("RSASSA_PKCS1v15_using_SHA512", crypto.SHA512, PKCS1v15, []int{3072, 4096})
)

// Method holds the configuration for an RSA signing algorithm.
type Method struct {
	name            string
	kind            crypto.Hash
	padding         Padding
	allowedKeySizes []int
	keyFn           KeyFn
	signFn          SignFn
	verifyFn        VerifyFn
}

// NewMethod creates a new RSA signing Method with the given name, hash, padding scheme, and allowed key sizes.
func NewMethod(name string, kind crypto.Hash, padding Padding, allowedKeySizes []int, options ...Option) *Method {
	cassert.NotEmpty(name, "name is empty")

	opts := NewOptions(options...)

	return &Method{
		name:            name,
		kind:            kind,
		padding:         padding,
		allowedKeySizes: allowedKeySizes,
		keyFn:           opts.keyFn,
		signFn:          opts.signFn,
		verifyFn:        opts.verifyFn,
	}
}

// Name returns the method's algorithm name.
func (m *Method) Name() string {
	cassert.NotNil(m, "method is nil")

	return m.name
}

// GenerateKey generates a new RSA private key of the specified bit size.
func (m *Method) GenerateKey(size int) (*rsa.PrivateKey, error) {
	cassert.NotNil(m, "method is nil")
	cassert.NotNil(m.keyFn, "method keyFn is nil")

	if cutils.NotIn(size, m.allowedKeySizes...) {
		return nil, ErrKeyGeneration(ErrKeySizeNotAllowed)
	}

	key, err := m.keyFn(size)
	if err != nil {
		return nil, ErrKeyGeneration(err)
	}

	return key, nil
}

// Sign produces an RSA signature over the provided data using the configured padding scheme.
func (m *Method) Sign(key *rsa.PrivateKey, data ctypes.Bytes) (ctypes.Bytes, error) {
	cassert.NotNil(m, "method is nil")
	cassert.NotNil(m.signFn, "method signFn is nil")

	signature, err := m.signFn(m, key, data)
	if err != nil {
		return nil, ErrSigning(err)
	}

	return signature, nil
}

// Verify checks an RSA signature over the provided data using the configured padding scheme.
func (m *Method) Verify(key *rsa.PublicKey, signature ctypes.Bytes, data ctypes.Bytes) (bool, error) {
	cassert.NotNil(m, "method is nil")
	cassert.NotNil(m.verifyFn, "method verifyFn is nil")

	ok, err := m.verifyFn(m, key, signature, data)
	if err != nil {
		return false, ErrVerification(err)
	}

	return ok, nil
}
