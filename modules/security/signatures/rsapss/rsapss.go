package rsapss

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"

	"github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/types"
	"github.com/guidomantilla/yarumo/common/utils"
)

var (
	RSASSA_PSS_SHA256 = NewMethod("RSASSA_PSS_SHA256", crypto.SHA256, rsa.PSSSaltLengthEqualsHash, 2048, 3072, 4096)
	RSASSA_PSS_SHA512 = NewMethod("RSASSA_PSS_SHA512", crypto.SHA512, rsa.PSSSaltLengthEqualsHash, 3072, 4096)
)

type Method struct {
	name            string
	kind            crypto.Hash
	saltLength      int
	allowedKeySizes []int
}

func NewMethod(name string, kind crypto.Hash, saltLength int, allowedKeySizes ...int) *Method {
	return &Method{
		name:            name,
		kind:            kind,
		saltLength:      saltLength,
		allowedKeySizes: allowedKeySizes,
	}
}
func (m *Method) Name() string {
	assert.NotNil(m, "method is nil")
	return m.name
}

// GenerateKey generates a new RSA private key.
func (m *Method) GenerateKey(size int) (*rsa.PrivateKey, error) {
	assert.NotNil(m, "method is nil")
	if utils.NotIn(size, m.allowedKeySizes...) {
		return nil, ErrKeyGeneration(ErrKeySizeNotAllowed)
	}
	key, err := rsa.GenerateKey(rand.Reader, size)
	if err != nil {
		return nil, ErrKeyGeneration(err)
	}
	return key, nil
}

// Sign creates an RSA-PSS signature over the given data using the specified
// method and private key.
//
// Parameters:
//   - method: the RSA-PSS method descriptor, defining the hash function,
//     allowed key sizes, and salt length requirements. Must not be nil.
//   - key: RSA private key used to produce the signature. Must not be nil and
//     its modulus size (in bits) must be included in method.allowedKeySizes.
//   - data: the message to be signed.
//
// Behavior:
//   - Validates the method and the RSA key size.
//   - Hashes the data using the hash function defined by the method.
//   - Produces an RSA-PSS signature using rsa.SignPSS with the method's
//     configured salt length and hash function.
//
// Returns:
//   - The generated signature as a byte slice.
//   - An error if the method or key are invalid, or if signing fails.
//
// Notes:
//   - RSA-PSS is the modern recommended scheme for RSA signatures (RFC 8017).
//   - A failure to sign returns (nil, ErrSignFailed).
func (m *Method) Sign(key *rsa.PrivateKey, data types.Bytes) (types.Bytes, error) {
	assert.NotNil(m, "method is nil")
	signature, err := Sign(m, key, data)
	if err != nil {
		return nil, ErrSigning(err)
	}
	return signature, nil
}

// Verify checks an RSA-PSS signature over the given data using the specified
// method and public key.
//
// Parameters:
//   - method: the RSA-PSS method descriptor, defining the hash function,
//     allowed key sizes, and salt length. Must not be nil.
//   - key: RSA public key used to verify the signature. Must not be nil and
//     its modulus size (in bits) must be included in method.allowedKeySizes.
//   - signature: the RSA-PSS signature to verify.
//   - data: the original message that was signed.
//
// Behavior:
//   - Validates the method and RSA key size.
//   - Hashes the data using the method's hash function.
//   - Invokes rsa.VerifyPSS with the method's salt length and hash algorithm.
//   - Returns (false, nil) if the signature is simply invalid.
//   - Returns (false, err) for verification errors unrelated to signature validity.
//
// Returns:
//   - (true, nil)  if the signature is valid.
//   - (false, nil) if the signature is invalid.
//   - (false, err) if the method or key are invalid, or if verification fails
//     due to an internal error.
//
// Notes:
//   - RSA-PSS is the recommended signature scheme according to RFC 8017.
//   - The function distinguishes between "invalid signature" and actual errors.
func (m *Method) Verify(key *rsa.PublicKey, signature types.Bytes, data types.Bytes) (bool, error) {
	assert.NotNil(m, "method is nil")
	ok, err := Verify(m, key, signature, data)
	if err != nil {
		return false, ErrVerification(err)
	}
	return ok, nil
}
