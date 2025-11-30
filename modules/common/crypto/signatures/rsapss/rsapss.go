package rsapss

import (
	"crypto"
	"crypto/rsa"

	"github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/types"
	"github.com/guidomantilla/yarumo/common/utils"
)

var (
	RSASSA_PSS_SHA256 = NewMethod("RSASSA_PSS_SHA256", crypto.SHA256, rsa.PSSSaltLengthEqualsHash, []int{2048, 3072, 4096})
	RSASSA_PSS_SHA512 = NewMethod("RSASSA_PSS_SHA512", crypto.SHA512, rsa.PSSSaltLengthEqualsHash, []int{3072, 4096})
)

type Method struct {
	name            string
	kind            crypto.Hash
	saltLength      int
	allowedKeySizes []int
	keyFn           KeyFn
	signFn          SignFn
	verifyFn        VerifyFn
}

func NewMethod(name string, kind crypto.Hash, saltLength int, allowedKeySizes []int, options ...Option) *Method {
	opts := NewOptions(options...)
	return &Method{
		name:            name,
		kind:            kind,
		saltLength:      saltLength,
		allowedKeySizes: allowedKeySizes,
		keyFn:           opts.keyFn,
		signFn:          opts.signFn,
		verifyFn:        opts.verifyFn,
	}
}

func (m *Method) Name() string {
	assert.NotNil(m, "method is nil")
	return m.name
}

// GenerateKey generates a new RSA private key of the specified bit size.
//
// Parameters:
//   - bits: the RSA modulus size in bits (commonly 2048, 3072, or 4096).
//
// Behavior:
//   - Uses crypto/rand as the secure randomness source.
//   - Invokes rsa.GenerateKey to create a new RSA private key.
//   - The public key is included within the returned *rsa.PrivateKey.
//
// Returns:
//   - A newly generated *rsa.PrivateKey.
//   - An error if key generation fails or if an invalid bit size is provided.
//
// Notes:
//   - Larger key sizes provide stronger security but are slower
//     (especially for signing).
//   - The function never panics.
//   - Call key.PublicKey to access the associated public key.
func (m *Method) GenerateKey(size int) (*rsa.PrivateKey, error) {
	assert.NotNil(m, "method is nil")
	assert.NotNil(m.keyFn, "method keyFn is nil")
	if utils.NotIn(size, m.allowedKeySizes...) {
		return nil, ErrKeyGeneration(ErrKeySizeNotAllowed)
	}

	key, err := m.keyFn(size)
	if err != nil {
		return nil, ErrKeyGeneration(err)
	}

	return key, nil
}

// Sign produces an RSA-PSS signature over the provided data using the given
// method configuration and RSA private key.
//
// Notes:it delegates to rsapss.SignFn. This is a function that takes an RSA-PSS Method specifying the hash function, allowed key sizes, and salt length policy.
//
// Parameters:
//   - key: RSA private key used to generate the signature.
//   - data: the message to be signed.
//
// Behavior:
//   - Returns an error if method or key are nil.
//   - Validates that the RSA key size (key.N.BitLen()) is included in
//     method.allowedKeySizes.
//   - Hashes the input data using method.kind.
//   - Uses rsa.SignPSS with the configured salt length and hash algorithm to
//     produce a probabilistic RSA-PSS signature.
//
// Returns:
//   - A byte slice containing the RSA-PSS signature.
//   - An error if signing fails or if the key size is not allowed.
//
// Notes:
//   - RSA-PSS is a modern, recommended signature scheme providing stronger
//     security than older PKCS#1 v1.5 signatures.
//   - Signature size is equal to the RSA modulus size.
//   - The function never panics and never returns a partial signature.
//   - Salt length policy (e.g., PSSSaltLengthEqualsHash) is controlled by
//     method.saltLength.
func (m *Method) Sign(key *rsa.PrivateKey, data types.Bytes) (types.Bytes, error) {
	assert.NotNil(m, "method is nil")
	assert.NotNil(m.signFn, "method signFn is nil")

	signature, err := m.signFn(m, key, data)
	if err != nil {
		return nil, ErrSigning(err)
	}

	return signature, nil
}

// Verify checks an RSA-PSS signature over the provided data using the given
// method configuration and RSA public key.
//
// Notes: it delegates to rsapss.VerifyFn. This is a function that takes an RSA-PSS Method specifying the hash function, allowed key sizes, and salt length policy.
//
// Parameters:
//   - key: RSA public key used for verification.
//   - signature: the RSA-PSS signature to verify.
//   - data: the original message that was signed.
//
// Behavior:
//
//   - Returns an error if method or key are nil.
//
//   - Validates that the RSA key size (key.N.BitLen()) is included in
//     method.allowedKeySizes.
//
//   - Hashes the input data using method.kind.
//
//   - Invokes rsa.VerifyPSS with the configured salt length and hash algorithm.
//
//   - Distinguishes between “invalid signature” and “internal error”:
//
//   - If rsa.VerifyPSS returns rsa.ErrVerification:
//     → returns (false, nil).
//
//   - Otherwise:
//     → wraps and returns the verification error.
//
// Returns:
//   - (true, nil)  if the signature is valid.
//   - (false, nil) if the signature is invalid.
//   - (false, err) if verification cannot be completed due to an error.
//
// Notes:
//   - RSA-PSS is a probabilistic signature scheme recommended by modern
//     cryptographic standards.
//   - Signature length must match the RSA modulus size.
//   - The function never panics and always treats signature mismatch as a
//     non-error condition.
func (m *Method) Verify(key *rsa.PublicKey, signature types.Bytes, data types.Bytes) (bool, error) {
	assert.NotNil(m, "method is nil")
	assert.NotNil(m.verifyFn, "method verifyFn is nil")

	ok, err := m.verifyFn(m, key, signature, data)
	if err != nil {
		return false, ErrVerification(err)
	}

	return ok, nil
}
