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

// GenerateKey generates a new ECDSA private key.
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
// method and private key.
//
// Parameters:
//   - method: cryptographic method defining the hash function, curve, and key size.
//   - key: ECDSA private key used to produce the signature. Must not be nil and
//     must use the same curve as method.curve.
//   - data: message to be signed.
//   - format: output signature format (RS or ASN1).
//
// Behavior:
//   - Returns an error if method or key are nil.
//   - Returns an error if the key's curve does not match the method's curve.
//   - Computes the hash of data using the method's hash function.
//   - For RS: calls ecdsa.Sign to obtain (r, s) and serializes r||s as
//     fixed-size big-endian byte slices (2*keySize).
//   - For ASN1: calls ecdsa.SignASN1 to produce a standard ASN.1 DER-encoded
//     ECDSA signature.
//
// Returns:
//   - A byte slice containing the encoded signature.
//   - An error if signing fails or if the format is not supported.
//
// Possible errors:
//   - ErrMethodInvalid
//   - ErrKeyInvalid
//   - ErrSignFailed
//   - ErrFormatUnsupported
//
// The function guarantees consistent output formatting and never panics.
func (m *Method) Sign(key *ecdsa.PrivateKey, data types.Bytes, format Format) (types.Bytes, error) {
	assert.NotNil(m, "method is nil")
	assert.NotNil(m.signFn, "method signFn is nil")

	signature, err := m.signFn(m, key, data, format)
	if err != nil {
		return nil, ErrSigning(err)
	}

	return signature, nil
}

// Verify checks an ECDSA signature over the given data using the provided
// method and public key.
//
// Parameters:
//   - method: cryptographic method defining the hash function and curve.
//   - key: ECDSA public key used for verification. Must not be nil and must
//     use the same curve as method.curve.
//   - signature: the signature to verify (in RS or ASN1 format).
//   - data: the original message that was signed.
//   - format: signature format (RS or ASN1).
//
// Behavior:
//   - Returns an error if method or key are nil.
//   - Returns an error if the key's curve does not match the method's curve.
//   - Computes the hash of data using the method's hash function.
//   - For RS: splits the signature into r||s using method.keySize and calls
//     ecdsa.Verify.
//   - For ASN1: calls ecdsa.VerifyASN1 with the hash and the ASN.1 DER-encoded
//     signature.
//
// Returns:
//   - (true, nil)  if the signature is valid.
//   - (false, nil) if the signature is invalid.
//   - (false, err) if the signature format is invalid or incompatible.
//
// Possible errors:
//   - ErrMethodInvalid
//   - ErrKeyInvalid
//   - ErrSignatureInvalid
//   - ErrFormatUnsupported
//
// The function never panics and does not return an error for a simple
// verification failure; it returns (false, nil) instead.
func (m *Method) Verify(key *ecdsa.PublicKey, signature types.Bytes, data types.Bytes, format Format) (bool, error) {
	assert.NotNil(m, "method is nil")
	assert.NotNil(m.verifyFn, "method verifyFn is nil")

	ok, err := m.verifyFn(m, key, signature, data, format)
	if err != nil {
		return false, ErrVerification(err)
	}

	return ok, nil
}
