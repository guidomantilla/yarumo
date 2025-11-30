package ecdsas

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"

	"github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/types"
)

var (
	ECDSA_with_SHA256_over_P256 = NewMethod("ECDSA_with_SHA256_over_P256", crypto.SHA256, 32, elliptic.P256())
	ECDSA_with_SHA512_over_P521 = NewMethod("ECDSA_with_SHA512_over_P521", crypto.SHA512, 66, elliptic.P521())
)

type Method struct {
	name     string
	kind     crypto.Hash
	keySize  int
	curve    elliptic.Curve
	keyFn    KeyFn
	signFn   SignFn
	verifyFn VerifyFn
}

func NewMethod(name string, kind crypto.Hash, keySize int, curve elliptic.Curve, options ...Option) *Method {
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

func (m *Method) Name() string {
	assert.NotNil(m, "method is nil")
	return m.name
}

// GenerateKey generates a new ECDSA private key for the given method.
//
// Notes: It delegates to ecdsas.KeyFn. This is a function that takes an ECDSA Method specifying the elliptic curve to use.
//
// Parameters: none
//
// Behavior:
//   - Returns an error if method is nil.
//   - Invokes ecdsa.GenerateKey using the curve defined in method.curve.
//   - Uses crypto/rand as the secure randomness source.
//
// Returns:
//   - A newly generated *ecdsa.PrivateKey.
//   - An error if key generation fails.
//
// Notes:
//   - The key size and security level depend entirely on the selected curve.
//   - This function never panics.
func (m *Method) GenerateKey() (*ecdsa.PrivateKey, error) {
	assert.NotNil(m, "method is nil")
	assert.NotNil(m.keyFn, "method keyFn is nil")

	key, err := m.keyFn(m)
	if err != nil {
		return nil, ErrKeyGeneration(err)
	}

	return key, nil
}

// Sign generates an ECDSA signature over the given data using the provided
// method, private key, and output format.
//
// Notes: it delegates to ecdsas.SignFn. This is a function that takes an ECDSA Method specifying the elliptic curve, the hash function, and key size to use.
//
// Parameters:
//   - key: ECDSA private key used to produce the signature.
//   - data: message to be signed.
//   - format: signature output format (RS or ASN1).
//
// Behavior:
//   - Returns an error if method or key are nil.
//   - Returns an error if the private key’s curve does not match method.curve.
//   - Hashes the input data using method.kind.
//   - For RS:
//   - Produces (r, s) using ecdsa.Sign.
//   - Serializes them as r||s, each big-endian and padded to method.keySize.
//   - Output length is exactly 2*keySize.
//   - For ASN1:
//   - Produces a standard ASN.1 DER-encoded ECDSA signature via ecdsa.SignASN1.
//
// Returns:
//   - A serialized ECDSA signature in the selected format.
//   - An error if signing fails or if the format is unsupported.
//
// Notes:
//   - RS format is commonly used in JOSE/JWT/WebAuthn.
//   - ASN1 format matches Go’s standard library and X.509 expectations.
//   - The function never panics.
func (m *Method) Sign(key *ecdsa.PrivateKey, data types.Bytes, format Format) (types.Bytes, error) {
	assert.NotNil(m, "method is nil")
	assert.NotNil(m.signFn, "method signFn is nil")

	signature, err := m.signFn(m, key, data, format)
	if err != nil {
		return nil, ErrSigning(err)
	}

	return signature, nil
}

// Verify checks an ECDSA signature over the given data using the specified
// method, public key, and signature format.
//
// Notes: it delegates to ecdsas.VerifyFn. This is a function that takes an ECDSA Method specifying the elliptic curve, the hash function, and key size to use.
//
// Parameters:
//   - key: ECDSA public key used for signature verification.
//   - signature: the signature to verify.
//   - data: the original message that was signed.
//   - format: signature format (RS or ASN1).
//
// Behavior:
//   - Returns an error if method or key are nil.
//   - Returns an error if the key’s curve does not match method.curve.
//   - Hashes the input data using method.kind.
//   - For RS:
//   - Expects signature length == 2*keySize.
//   - Splits signature into r||s (each padded big-endian integer).
//   - Uses ecdsa.Verify to validate (r, s).
//   - For ASN1:
//   - Uses ecdsa.VerifyASN1 to validate a DER-encoded ASN.1 signature.
//
// Returns:
//   - (true, nil)  if the signature is valid.
//   - (false, nil) if the signature is invalid.
//   - (false, err) if the signature format is unsupported or the input is malformed.
//
// Notes:
//   - RS format matches JOSE/JWT/WebAuthn conventions.
//   - ASN1 format matches Go’s crypto/x509 expectations.
//   - The function never panics and treats verification failure as a non-error.
func (m *Method) Verify(key *ecdsa.PublicKey, signature types.Bytes, data types.Bytes, format Format) (bool, error) {
	assert.NotNil(m, "method is nil")
	assert.NotNil(m.verifyFn, "method verifyFn is nil")

	ok, err := m.verifyFn(m, key, signature, data, format)
	if err != nil {
		return false, ErrVerification(err)
	}

	return ok, nil
}
