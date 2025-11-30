package ed25519

import (
	"crypto/ed25519"

	"github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/types"
)

var (
	Ed25519 = NewMethod("Ed25519")
)

type Method struct {
	name     string
	keyFn    KeyFn
	signFn   SignFn
	verifyFn VerifyFn
}

func NewMethod(name string, options ...Option) *Method {
	opts := NewOptions(options...)
	return &Method{
		name:     name,
		keyFn:    opts.keyFn,
		signFn:   opts.signFn,
		verifyFn: opts.verifyFn,
	}
}

func (m *Method) Name() string {
	assert.NotNil(m, "method is nil")
	return m.name
}

// GenerateKey generates a new Ed25519 public/private key pair.
//
// Notes: it delegates to ed25519.KeyFn.
//
// Parameters: none
//
// Behavior:
//   - Uses crypto/rand as a secure randomness source.
//   - Invokes ed25519.GenerateKey to produce a 32-byte public key and
//     a 64-byte private key (which includes the public key at the end).
//
// Returns:
//   - publicKey: the generated Ed25519 public key.
//   - privateKey: the generated Ed25519 private key.
//   - error: non-nil only if key generation fails.
//
// Notes:
//   - Ed25519 keys always have fixed sizes:
//   - PublicKey:  32 bytes
//   - PrivateKey: 64 bytes
//   - This function never panics.
func (m *Method) GenerateKey() (ed25519.PrivateKey, error) {
	assert.NotNil(m, "method is nil")
	assert.NotNil(m.keyFn, "method keyFn is nil")

	_, key, err := m.keyFn()
	if err != nil {
		return nil, ErrKeyGeneration(err)
	}

	return key, nil
}

// Sign produces an Ed25519 signature over the provided data using the given
// method and private key.
//
// Notes: it delegates to ed25519.SignFn. This is a function that takes an Ed25519 Method (currently unused but validated).
//
// Parameters:
//   - key: Ed25519 private key used to generate the signature.
//   - data: the message to be signed.
//
// Behavior:
//   - Returns an error if method or key are nil.
//   - Returns an error if the private key length is not exactly
//     ed25519.PrivateKeySize (64 bytes).
//   - Uses ed25519.Sign to produce a deterministic 64-byte signature.
//
// Returns:
//   - A 64-byte Ed25519 signature.
//   - An error if the inputs are invalid.
//
// Notes:
//   - Ed25519 does not require hashing beforehand; the API signs the raw message.
//   - The function never panics and never returns a partial signature.
//   - Signature format is always the standard Ed25519 64-byte concatenation.
func (m *Method) Sign(key *ed25519.PrivateKey, data types.Bytes) (types.Bytes, error) {
	assert.NotNil(m, "method is nil")
	assert.NotNil(m.signFn, "method signFn is nil")

	signature, err := m.signFn(m, key, data)
	if err != nil {
		return nil, ErrSigning(err)
	}

	return signature, nil
}

// Verify checks an Ed25519 signature over the given data using the provided
// method and public key.
//
// Notes: it delegates to ed25519.VerifyFn. This is a function that takes an Ed25519 Method (currently unused but validated).
//
// Parameters:
//   - key: Ed25519 public key used for verification.
//   - signature: the 64-byte Ed25519 signature to validate.
//   - data: the original message that was signed.
//
// Behavior:
//   - Returns an error if method or key are nil.
//   - Returns an error if the public key length is invalid.
//   - Returns an error if the signature length is not exactly
//     ed25519.SignatureSize (64 bytes).
//   - Uses ed25519.Verify to check signature validity.
//
// Returns:
//   - (true, nil)  if the signature is valid.
//   - (false, nil) if the signature is invalid.
//   - (false, err) if the input is malformed.
//
// Notes:
//   - Ed25519 uses deterministic signatures and verifies over the raw message
//     (no pre-hashing).
//   - The function never panics and treats a failed verification as a non-error.
func (m *Method) Verify(key *ed25519.PublicKey, signature, data types.Bytes) (bool, error) {
	assert.NotNil(m, "method is nil")
	assert.NotNil(m.verifyFn, "method verifyFn is nil")

	ok, err := m.verifyFn(m, key, signature, data)
	if err != nil {
		return false, ErrVerification(err)
	}

	return ok, nil
}
