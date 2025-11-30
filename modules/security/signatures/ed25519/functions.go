package ed25519

import (
	"crypto/ed25519"

	"github.com/guidomantilla/yarumo/common/types"
)

// Sign generates an Ed25519 signature over the given data using the provided
// method and private key.
//
// Parameters:
//   - method: the Ed25519 method descriptor. Must not be nil.
//   - key: pointer to an Ed25519 private key. Must not be nil and must have
//     length ed25519.PrivateKeySize.
//   - data: the message to be signed.
//
// Behavior:
//   - Validates the method and key.
//   - Produces a deterministic Ed25519 signature (EdDSA).
//   - Returns the signature as a byte slice.
//
// Returns:
//   - The generated signature.
//   - An error if the method or key are invalid.
//
// Notes:
//   - Ed25519 performs hashing internally; no external hash function is applied.
//   - The function never returns an error for a simple verification failure
//     (that is handled in Verify).
func Sign(method *Method, key *ed25519.PrivateKey, data types.Bytes) (types.Bytes, error) {
	if method == nil {
		return nil, ErrMethodIsNil
	}
	if key == nil {
		return nil, ErrKeyIsNil
	}
	if len(*key) != ed25519.PrivateKeySize {
		return nil, ErrKeyLengthIsInvalid
	}

	out := ed25519.Sign(*key, data)
	return out, nil
}

// Verify checks an Ed25519 signature over the given data using the provided
// method and public key.
//
// Parameters:
//   - method: the Ed25519 method descriptor. Must not be nil.
//   - key: pointer to an Ed25519 public key. Must not be nil and must have
//     length ed25519.PublicKeySize.
//   - signature: the signature to verify. Must have length ed25519.SignatureSize.
//   - data: the original message that was signed.
//
// Behavior:
//   - Validates the method, key, and signature sizes.
//   - Uses ed25519.Verify to perform a constant-time signature check.
//   - Returns true if the signature is valid, false otherwise.
//
// Returns:
//   - (true, nil)  if the signature is valid.
//   - (false, nil) if the signature is invalid.
//   - (false, err) if method or key are invalid.
//
// Notes:
//   - Ed25519 handles hashing internally; no external hash function is involved.
//   - The function does not treat an invalid signature as an error; it returns
//     (false, nil) in that case.
func Verify(method *Method, key *ed25519.PublicKey, signature, data types.Bytes) (bool, error) {
	if method == nil {
		return false, ErrMethodIsNil
	}
	if key == nil {
		return false, ErrKeyIsNil
	}
	if len(*key) != ed25519.PrivateKeySize {
		return false, ErrKeyLengthIsInvalid
	}
	if len(signature) != ed25519.SignatureSize {
		return false, ErrSignatureLengthInvalid
	}

	ok := ed25519.Verify(*key, data, signature)
	return ok, nil
}
