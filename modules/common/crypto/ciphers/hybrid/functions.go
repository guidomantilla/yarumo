package hybrid

import (
	"crypto/rand"

	"github.com/cloudflare/circl/hpke"
	"github.com/cloudflare/circl/kem"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
	ctypes "github.com/guidomantilla/yarumo/common/types"
)

// suite builds the circl HPKE Suite for the given method.
func suite(method *Method) hpke.Suite {
	return hpke.NewSuite(method.kemID, method.kdfID, method.aeadID)
}

// generateKey produces an HPKE key pair for the method's KEM.
func generateKey(method *Method) (kem.PublicKey, kem.PrivateKey, error) {
	if method == nil {
		return nil, nil, ErrMethodIsNil
	}

	scheme := method.kemID.Scheme()

	pub, priv, err := scheme.GenerateKeyPair()
	if err != nil {
		return nil, nil, cerrs.Wrap(ErrKeyGenerationFailed, err)
	}

	return pub, priv, nil
}

// encrypt performs HPKE base-mode encryption. The output is the concatenation
// of the KEM encapsulated key and the AEAD ciphertext.
func encrypt(method *Method, recipientPub kem.PublicKey, plaintext, info ctypes.Bytes) (ctypes.Bytes, error) {
	if method == nil {
		return nil, ErrMethodIsNil
	}

	if recipientPub == nil {
		return nil, ErrPublicKeyIsNil
	}

	if recipientPub.Scheme().Name() != method.kemID.Scheme().Name() {
		return nil, ErrKeyTypeMismatch
	}

	s := suite(method)

	sender, err := s.NewSender(recipientPub, info)
	if err != nil {
		return nil, cerrs.Wrap(ErrSuiteSetupFailed, err)
	}

	enc, sealer, err := sender.Setup(rand.Reader)
	if err != nil {
		return nil, cerrs.Wrap(ErrEncapsulationFailed, err)
	}

	ct, err := sealer.Seal(plaintext, nil)
	if err != nil {
		return nil, cerrs.Wrap(ErrEncryptionFailed, err)
	}

	out := make([]byte, 0, len(enc)+len(ct))
	out = append(out, enc...)
	out = append(out, ct...)

	return out, nil
}

// decrypt performs HPKE base-mode decryption. It expects the wire format
// produced by encrypt (encapsulated key || AEAD ciphertext).
func decrypt(method *Method, recipientPriv kem.PrivateKey, ciphertext, info ctypes.Bytes) (ctypes.Bytes, error) {
	if method == nil {
		return nil, ErrMethodIsNil
	}

	if recipientPriv == nil {
		return nil, ErrPrivateKeyIsNil
	}

	if recipientPriv.Scheme().Name() != method.kemID.Scheme().Name() {
		return nil, ErrKeyTypeMismatch
	}

	scheme := method.kemID.Scheme()
	encSize := scheme.CiphertextSize()

	if len(ciphertext) < encSize {
		return nil, ErrCiphertextTooShort
	}

	enc := ciphertext[:encSize]
	ct := ciphertext[encSize:]

	s := suite(method)

	receiver, err := s.NewReceiver(recipientPriv, info)
	if err != nil {
		return nil, cerrs.Wrap(ErrSuiteSetupFailed, err)
	}

	opener, err := receiver.Setup(enc)
	if err != nil {
		return nil, cerrs.Wrap(ErrDecapsulationFailed, err)
	}

	pt, err := opener.Open(ct, nil)
	if err != nil {
		return nil, cerrs.Wrap(ErrDecryptionFailed, err)
	}

	return pt, nil
}
