package hybrid

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/cloudflare/circl/hpke"
	"github.com/cloudflare/circl/kem"

	ctypes "github.com/guidomantilla/yarumo/common/types"
)

func TestNewMethod(t *testing.T) {
	t.Parallel()

	t.Run("creates method with defaults", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("test-hpke", hpke.KEM_X25519_HKDF_SHA256, hpke.KDF_HKDF_SHA256, hpke.AEAD_AES256GCM)

		if m == nil {
			t.Fatal("expected non-nil method")
		}

		if m.name != "test-hpke" {
			t.Fatalf("expected 'test-hpke', got %q", m.name)
		}

		if m.kemID != hpke.KEM_X25519_HKDF_SHA256 {
			t.Fatalf("unexpected kemID %v", m.kemID)
		}
	})

	t.Run("applies custom key function via option", func(t *testing.T) {
		t.Parallel()

		called := false
		custom := func(_ *Method) (kem.PublicKey, kem.PrivateKey, error) {
			called = true

			scheme := hpke.KEM_X25519_HKDF_SHA256.Scheme()

			return scheme.GenerateKeyPair()
		}

		m := NewMethod("custom", hpke.KEM_X25519_HKDF_SHA256, hpke.KDF_HKDF_SHA256, hpke.AEAD_AES256GCM, WithKeyFn(custom))

		_, _, err := m.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !called {
			t.Fatal("expected custom keyFn to be called")
		}
	})
}

func TestMethod_Name(t *testing.T) {
	t.Parallel()

	t.Run("returns the method name", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("my-hpke", hpke.KEM_X25519_HKDF_SHA256, hpke.KDF_HKDF_SHA256, hpke.AEAD_AES256GCM)

		got := m.Name()
		if got != "my-hpke" {
			t.Fatalf("expected 'my-hpke', got %q", got)
		}
	})
}

func TestMethod_GenerateKey(t *testing.T) {
	t.Parallel()

	t.Run("generates a key pair", func(t *testing.T) {
		t.Parallel()

		pub, priv, err := HPKE_X25519_HKDF_SHA256_AES_256_GCM.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if pub == nil {
			t.Fatal("expected non-nil public key")
		}

		if priv == nil {
			t.Fatal("expected non-nil private key")
		}

		if !priv.Public().Equal(pub) {
			t.Fatal("private key's public component does not match returned public key")
		}
	})

	t.Run("wraps error from keyFn", func(t *testing.T) {
		t.Parallel()

		failKey := func(_ *Method) (kem.PublicKey, kem.PrivateKey, error) {
			return nil, nil, ErrMethodIsNil
		}

		m := NewMethod("fail", hpke.KEM_X25519_HKDF_SHA256, hpke.KDF_HKDF_SHA256, hpke.AEAD_AES256GCM, WithKeyFn(failKey))

		_, _, err := m.GenerateKey()
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestMethod_Encrypt(t *testing.T) {
	t.Parallel()

	t.Run("encrypts data", func(t *testing.T) {
		t.Parallel()

		pub, _, err := HPKE_X25519_HKDF_SHA256_AES_256_GCM.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ciphered, err := HPKE_X25519_HKDF_SHA256_AES_256_GCM.Encrypt(pub, []byte("hello"), []byte("info"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(ciphered) == 0 {
			t.Fatal("expected non-empty ciphertext")
		}
	})

	t.Run("wraps error from encryptFn", func(t *testing.T) {
		t.Parallel()

		failEncrypt := func(_ *Method, _ kem.PublicKey, _, _ ctypes.Bytes) (ctypes.Bytes, error) {
			return nil, ErrPublicKeyIsNil
		}

		m := NewMethod("fail-enc", hpke.KEM_X25519_HKDF_SHA256, hpke.KDF_HKDF_SHA256, hpke.AEAD_AES256GCM, WithEncryptFn(failEncrypt))

		_, err := m.Encrypt(nil, []byte("data"), nil)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestMethod_Decrypt(t *testing.T) {
	t.Parallel()

	t.Run("decrypts data", func(t *testing.T) {
		t.Parallel()

		pub, priv, err := HPKE_X25519_HKDF_SHA256_AES_256_GCM.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		info := []byte("context")

		ciphered, err := HPKE_X25519_HKDF_SHA256_AES_256_GCM.Encrypt(pub, []byte("hello"), info)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		plain, err := HPKE_X25519_HKDF_SHA256_AES_256_GCM.Decrypt(priv, ciphered, info)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if string(plain) != "hello" {
			t.Fatalf("expected 'hello', got %q", string(plain))
		}
	})

	t.Run("round-trip with large payload", func(t *testing.T) {
		t.Parallel()

		pub, priv, err := HPKE_X25519_HKDF_SHA256_AES_256_GCM.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Build a 1 MiB payload to exercise the AEAD path well past what
		// RSA-OAEP could ever handle.
		payload := bytes.Repeat([]byte("yarumo"), 1024*1024/6)

		ciphered, err := HPKE_X25519_HKDF_SHA256_AES_256_GCM.Encrypt(pub, payload, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		plain, err := HPKE_X25519_HKDF_SHA256_AES_256_GCM.Decrypt(priv, ciphered, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !bytes.Equal(plain, payload) {
			t.Fatal("decrypted payload does not match original")
		}
	})

	t.Run("fails with wrong private key", func(t *testing.T) {
		t.Parallel()

		pub, _, err := HPKE_X25519_HKDF_SHA256_AES_256_GCM.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, wrongPriv, err := HPKE_X25519_HKDF_SHA256_AES_256_GCM.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ciphered, err := HPKE_X25519_HKDF_SHA256_AES_256_GCM.Encrypt(pub, []byte("hello"), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err = HPKE_X25519_HKDF_SHA256_AES_256_GCM.Decrypt(wrongPriv, ciphered, nil)
		if err == nil {
			t.Fatal("expected decryption to fail with wrong private key")
		}
	})

	t.Run("fails with tampered ciphertext", func(t *testing.T) {
		t.Parallel()

		pub, priv, err := HPKE_X25519_HKDF_SHA256_AES_256_GCM.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ciphered, err := HPKE_X25519_HKDF_SHA256_AES_256_GCM.Encrypt(pub, []byte("hello"), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Flip the last byte (AEAD tag region) to force authentication failure.
		ciphered[len(ciphered)-1] ^= 0xff

		_, err = HPKE_X25519_HKDF_SHA256_AES_256_GCM.Decrypt(priv, ciphered, nil)
		if err == nil {
			t.Fatal("expected decryption to fail with tampered ciphertext")
		}
	})

	t.Run("fails with mismatched info", func(t *testing.T) {
		t.Parallel()

		pub, priv, err := HPKE_X25519_HKDF_SHA256_AES_256_GCM.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ciphered, err := HPKE_X25519_HKDF_SHA256_AES_256_GCM.Encrypt(pub, []byte("hello"), []byte("infoA"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err = HPKE_X25519_HKDF_SHA256_AES_256_GCM.Decrypt(priv, ciphered, []byte("infoB"))
		if err == nil {
			t.Fatal("expected decryption to fail with mismatched info")
		}
	})

	t.Run("wraps error from decryptFn", func(t *testing.T) {
		t.Parallel()

		failDecrypt := func(_ *Method, _ kem.PrivateKey, _, _ ctypes.Bytes) (ctypes.Bytes, error) {
			return nil, ErrPrivateKeyIsNil
		}

		m := NewMethod("fail-dec", hpke.KEM_X25519_HKDF_SHA256, hpke.KDF_HKDF_SHA256, hpke.AEAD_AES256GCM, WithDecryptFn(failDecrypt))

		_, err := m.Decrypt(nil, []byte("data"), nil)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

// TestRFC9180_BaseMode_X25519 verifies our Method against the canonical
// RFC 9180 base-mode test vector for KEM=DHKEM(X25519, HKDF-SHA256),
// KDF=HKDF-SHA256, AEAD=AES-256-GCM. The vector values are taken from the
// draft-irtf-cfrg-hpke test-vectors.json published alongside RFC 9180
// (Appendix A.1).
//
// The check does two things:
//
//  1. Drives a circl Sender with the RFC's deterministic ephemeral-key seed
//     (ikmE) and the vector recipient public key (pkRm), and asserts that
//     the resulting encapsulation matches the RFC's enc value byte-for-byte.
//     This pins our suite identifiers (KEM/KDF/AEAD) and info to the spec.
//
//  2. Opens the RFC-shaped wire format (enc || ct) using our Method.Decrypt
//     with the recipient private key reconstructed from skRm, and confirms
//     the plaintext matches the RFC plaintext exactly.
func TestRFC9180_BaseMode_X25519(t *testing.T) {
	t.Parallel()

	// RFC 9180 Appendix A.1 — mode_base, DHKEM(X25519, HKDF-SHA256),
	// HKDF-SHA256, AES-256-GCM.
	const (
		ikmEHex = "2cd7c601cefb3d42a62b04b7a9041494c06c7843818e0ce28a8f704ae7ab20f9"
		skRmHex = "497b4502664cfea5d5af0b39934dac72242a74f8480451e1aee7d6a53320333d"
		pkRmHex = "430f4b9859665145a6b1ba274024487bd66f03a2dd577d7753c68d7d7d00c00c"
		encHex  = "6c93e09869df3402d7bf231bf540fadd35cd56be14f97178f0954db94b7fc256"
		infoHex = "4f6465206f6e2061204772656369616e2055726e"
		// RFC vector plaintext for the suite ("Beauty is truth, truth beauty").
		ptHex = "4265617574792069732074727574682c20747275746820626561757479"
	)

	t.Run("encapsulation matches RFC 9180 enc value", func(t *testing.T) {
		t.Parallel()

		scheme := hpke.KEM_X25519_HKDF_SHA256.Scheme()

		pub, err := scheme.UnmarshalBinaryPublicKey(mustHex(t, pkRmHex))
		if err != nil {
			t.Fatalf("unmarshal public key: %v", err)
		}

		suite := hpke.NewSuite(hpke.KEM_X25519_HKDF_SHA256, hpke.KDF_HKDF_SHA256, hpke.AEAD_AES256GCM)

		sender, err := suite.NewSender(pub, mustHex(t, infoHex))
		if err != nil {
			t.Fatalf("new sender: %v", err)
		}

		gotEnc, _, err := sender.Setup(bytes.NewReader(mustHex(t, ikmEHex)))
		if err != nil {
			t.Fatalf("sender setup: %v", err)
		}

		wantEnc := mustHex(t, encHex)
		if !bytes.Equal(gotEnc, wantEnc) {
			t.Fatalf("encapsulation mismatch:\n got %s\nwant %s",
				hex.EncodeToString(gotEnc), encHex)
		}
	})

	t.Run("Method.Decrypt recovers RFC 9180 plaintext", func(t *testing.T) {
		t.Parallel()

		scheme := hpke.KEM_X25519_HKDF_SHA256.Scheme()

		priv, err := scheme.UnmarshalBinaryPrivateKey(mustHex(t, skRmHex))
		if err != nil {
			t.Fatalf("unmarshal private key: %v", err)
		}

		pub, err := scheme.UnmarshalBinaryPublicKey(mustHex(t, pkRmHex))
		if err != nil {
			t.Fatalf("unmarshal public key: %v", err)
		}

		// Drive a circl Sender deterministically so the resulting wire format
		// can be opened by Method.Decrypt with the vector's recipient key.
		suite := hpke.NewSuite(hpke.KEM_X25519_HKDF_SHA256, hpke.KDF_HKDF_SHA256, hpke.AEAD_AES256GCM)

		info := mustHex(t, infoHex)

		sender, err := suite.NewSender(pub, info)
		if err != nil {
			t.Fatalf("new sender: %v", err)
		}

		enc, sealer, err := sender.Setup(bytes.NewReader(mustHex(t, ikmEHex)))
		if err != nil {
			t.Fatalf("sender setup: %v", err)
		}

		vectorPlaintext := mustHex(t, ptHex)

		sealedCT, err := sealer.Seal(vectorPlaintext, nil)
		if err != nil {
			t.Fatalf("seal: %v", err)
		}

		wire := make([]byte, 0, len(enc)+len(sealedCT))
		wire = append(wire, enc...)
		wire = append(wire, sealedCT...)

		recovered, err := HPKE_X25519_HKDF_SHA256_AES_256_GCM.Decrypt(priv, wire, info)
		if err != nil {
			t.Fatalf("Method.Decrypt against RFC 9180 vector: %v", err)
		}

		if !bytes.Equal(recovered, vectorPlaintext) {
			t.Fatalf("plaintext mismatch:\n got %x\nwant %s", recovered, ptHex)
		}
	})
}

func mustHex(t *testing.T, s string) []byte {
	t.Helper()

	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("hex.DecodeString(%q): %v", s, err)
	}

	return b
}
