package rsapss

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"

	"github.com/guidomantilla/yarumo/common/errs"
	"github.com/guidomantilla/yarumo/common/types"
	"github.com/guidomantilla/yarumo/common/utils"
	"github.com/guidomantilla/yarumo/security/hashes"
)

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
func Sign(method *Method, key *rsa.PrivateKey, data types.Bytes) (types.Bytes, error) {
	if method == nil {
		return nil, ErrMethodInvalid
	}
	if key == nil {
		return nil, ErrKeyInvalid
	}
	if utils.NotIn(key.N.BitLen(), method.allowedKeySizes...) {
		return nil, ErrKeyInvalid
	}

	h := hashes.Hash(method.kind, data)
	out, err := rsa.SignPSS(rand.Reader, key, method.kind, h, &rsa.PSSOptions{
		SaltLength: method.saltLength,
		Hash:       method.kind,
	})
	if err != nil {
		return nil, errs.Wrap(ErrSignFailed, err)
	}
	return out, nil
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
func Verify(method *Method, key *rsa.PublicKey, signature types.Bytes, data types.Bytes) (bool, error) {
	if method == nil {
		return false, ErrMethodInvalid
	}
	if key == nil {
		return false, ErrKeyInvalid
	}
	if utils.NotIn(key.N.BitLen(), method.allowedKeySizes...) {
		return false, ErrKeyInvalid
	}

	h := hashes.Hash(method.kind, data)
	err := rsa.VerifyPSS(key, method.kind, h, signature, &rsa.PSSOptions{
		SaltLength: method.saltLength,
		Hash:       method.kind,
	})
	if err != nil {
		if errors.Is(err, rsa.ErrVerification) {
			return false, nil
		}
		return false, errs.Wrap(ErrVerifyFailed, err)
	}

	return true, nil
}
