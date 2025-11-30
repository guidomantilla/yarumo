package ecdsas

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"

	"github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/types"
)

var (
	ECDSA_with_SHA256_over_P256 = NewMethod("ECDSA_with_SHA256_over_P256", crypto.SHA256, 32, elliptic.P256())
	ECDSA_with_SHA512_over_P521 = NewMethod("ECDSA_with_SHA512_over_P521", crypto.SHA512, 66, elliptic.P521())
)

type Method struct {
	name    string
	kind    crypto.Hash
	keySize int
	curve   elliptic.Curve
}

func NewMethod(name string, kind crypto.Hash, keySize int, curve elliptic.Curve) *Method {
	return &Method{
		name:    name,
		kind:    kind,
		keySize: keySize,
		curve:   curve,
	}
}

func (m *Method) Name() string {
	assert.NotNil(m, "method is nil")
	return m.name
}

func (m *Method) GenerateKey() (*ecdsa.PrivateKey, error) {
	assert.NotNil(m, "method is nil")
	return ecdsa.GenerateKey(m.curve, rand.Reader)
}

// Sign generates an ECDSA signature over the given data using the provided
// method and private key.
//
// Parameters:
//   - method: cryptographic method defining the hash function, curve, and key size.
//   - key: ECDSA private key used to produce the signature.
//   - data: message to be signed.
//   - format: output signature format (RS or ASN1DER).
//
// Behavior:
//   - Returns an error if method or key are nil.
//   - Returns an error if the key's curve does not match the method's curve.
//   - Computes the hash of data using the method's hash function.
//   - Invokes ecdsa.Sign to produce (r, s).
//   - RS format: serializes r||s as fixed-size big-endian byte slices (2*keySize).
//   - ASN1DER format: encodes r and s into a standard ASN.1 DER structure.
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
	return Sign(m, key, data, format)
}

// Verify checks an ECDSA signature over the given data using the provided
// method and public key.
//
// Parameters:
//   - method: cryptographic method defining the hash function and curve.
//   - key: ECDSA public key used for verification.
//   - signature: the signature to verify (in RS or ASN1DER format).
//   - data: the original message that was signed.
//   - format: signature format (RS or ASN1DER).
//
// Behavior:
//   - Returns an error if method or key are nil.
//   - Returns an error if the key's curve does not match the method's curve.
//   - RS format: splits the signature into r||s using method.keySize.
//   - ASN1DER format: decodes a tuple ASN.1 structure containing R and S.
//   - Hashes the data using the method's hash and invokes ecdsa.Verify.
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
// verification failure (it returns false, nil instead).
func (m *Method) Verify(key *ecdsa.PublicKey, signature types.Bytes, data types.Bytes, format Format) (bool, error) {
	assert.NotNil(m, "method is nil")
	return Verify(m, key, signature, data, format)
}
