package hybrid

import (
	"github.com/cloudflare/circl/hpke"
	"github.com/cloudflare/circl/kem"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	ctypes "github.com/guidomantilla/yarumo/core/common/types"
)

// Predefined HPKE methods registered at package init. Identifiers follow
// RFC 9180 §7.
var (
	// HPKE_X25519_HKDF_SHA256_AES_256_GCM is the RFC 9180 base-mode cipher
	// suite combining DHKEM(X25519, HKDF-SHA256), HKDF-SHA256, and
	// AES-256-GCM. KEM=0x0020, KDF=0x0001, AEAD=0x0002.
	HPKE_X25519_HKDF_SHA256_AES_256_GCM = NewMethod(
		"HPKE_X25519_HKDF_SHA256_AES_256_GCM",
		hpke.KEM_X25519_HKDF_SHA256,
		hpke.KDF_HKDF_SHA256,
		hpke.AEAD_AES256GCM,
	)
)

// Method holds the configuration for an HPKE cipher suite.
type Method struct {
	name      string
	kemID     hpke.KEM
	kdfID     hpke.KDF
	aeadID    hpke.AEAD
	keyFn     KeyFn
	encryptFn EncryptFn
	decryptFn DecryptFn
}

// NewMethod creates a new hybrid Method with the given name and HPKE cipher
// suite identifiers (KEM, KDF, AEAD per RFC 9180 §7).
func NewMethod(name string, kemID hpke.KEM, kdfID hpke.KDF, aeadID hpke.AEAD, options ...Option) *Method {
	cassert.NotEmpty(name, "name is empty")

	opts := NewOptions(options...)

	return &Method{
		name:      name,
		kemID:     kemID,
		kdfID:     kdfID,
		aeadID:    aeadID,
		keyFn:     opts.keyFn,
		encryptFn: opts.encryptFn,
		decryptFn: opts.decryptFn,
	}
}

// Name returns the method's algorithm name.
func (m *Method) Name() string {
	cassert.NotNil(m, "method is nil")

	return m.name
}

// GenerateKey produces a fresh KEM key pair for this method. The public key
// is what senders need to encrypt to the recipient; the private key stays on
// the recipient.
func (m *Method) GenerateKey() (kem.PublicKey, kem.PrivateKey, error) {
	cassert.NotNil(m, "method is nil")
	cassert.NotNil(m.keyFn, "method keyFn is nil")

	pub, priv, err := m.keyFn(m)
	if err != nil {
		return nil, nil, ErrKeyGeneration(err)
	}

	return pub, priv, nil
}

// Encrypt seals plaintext to the recipient's public key. The info argument
// is bound into the HPKE key schedule so callers can scope ciphertexts to a
// protocol or context label; the recipient must pass the same info to
// Decrypt. The returned bytes are the concatenation of the KEM encapsulated
// key and the AEAD ciphertext.
func (m *Method) Encrypt(recipientPub kem.PublicKey, plaintext, info ctypes.Bytes) (ctypes.Bytes, error) {
	cassert.NotNil(m, "method is nil")
	cassert.NotNil(m.encryptFn, "method encryptFn is nil")

	ciphertext, err := m.encryptFn(m, recipientPub, plaintext, info)
	if err != nil {
		return nil, ErrEncrypt(err)
	}

	return ciphertext, nil
}

// Decrypt opens an HPKE ciphertext using the recipient's private key. info
// must match the value used at encryption time, otherwise the AEAD tag
// check will fail.
func (m *Method) Decrypt(recipientPriv kem.PrivateKey, ciphertext, info ctypes.Bytes) (ctypes.Bytes, error) {
	cassert.NotNil(m, "method is nil")
	cassert.NotNil(m.decryptFn, "method decryptFn is nil")

	plaintext, err := m.decryptFn(m, recipientPriv, ciphertext, info)
	if err != nil {
		return nil, ErrDecrypt(err)
	}

	return plaintext, nil
}
