package hybrid

import (
	"errors"
	"testing"

	"github.com/cloudflare/circl/hpke"
)

func TestGenerateKey(t *testing.T) {
	t.Parallel()

	t.Run("returns matching key pair", func(t *testing.T) {
		t.Parallel()

		pub, priv, err := generateKey(HPKE_X25519_HKDF_SHA256_AES_256_GCM)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !priv.Public().Equal(pub) {
			t.Fatal("public key does not match private key's public component")
		}
	})

	t.Run("returns error for nil method", func(t *testing.T) {
		t.Parallel()

		_, _, err := generateKey(nil)
		if !errors.Is(err, ErrMethodIsNil) {
			t.Fatalf("expected ErrMethodIsNil, got %v", err)
		}
	})
}

func TestEncrypt(t *testing.T) {
	t.Parallel()

	t.Run("encrypts data", func(t *testing.T) {
		t.Parallel()

		pub, _, err := generateKey(HPKE_X25519_HKDF_SHA256_AES_256_GCM)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		out, err := encrypt(HPKE_X25519_HKDF_SHA256_AES_256_GCM, pub, []byte("hello"), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(out) == 0 {
			t.Fatal("expected non-empty output")
		}
	})

	t.Run("returns error for nil method", func(t *testing.T) {
		t.Parallel()

		_, err := encrypt(nil, nil, nil, nil)
		if !errors.Is(err, ErrMethodIsNil) {
			t.Fatalf("expected ErrMethodIsNil, got %v", err)
		}
	})

	t.Run("returns error for nil public key", func(t *testing.T) {
		t.Parallel()

		_, err := encrypt(HPKE_X25519_HKDF_SHA256_AES_256_GCM, nil, nil, nil)
		if !errors.Is(err, ErrPublicKeyIsNil) {
			t.Fatalf("expected ErrPublicKeyIsNil, got %v", err)
		}
	})

	t.Run("returns error for mismatched key scheme", func(t *testing.T) {
		t.Parallel()

		// Generate a P-256 key (different scheme than X25519).
		other := hpke.KEM_P256_HKDF_SHA256.Scheme()

		pub, _, err := other.GenerateKeyPair()
		if err != nil {
			t.Fatalf("unexpected error generating P-256 key: %v", err)
		}

		_, err = encrypt(HPKE_X25519_HKDF_SHA256_AES_256_GCM, pub, []byte("data"), nil)
		if !errors.Is(err, ErrKeyTypeMismatch) {
			t.Fatalf("expected ErrKeyTypeMismatch, got %v", err)
		}
	})
}

func TestDecrypt(t *testing.T) {
	t.Parallel()

	t.Run("decrypts data", func(t *testing.T) {
		t.Parallel()

		pub, priv, err := generateKey(HPKE_X25519_HKDF_SHA256_AES_256_GCM)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ct, err := encrypt(HPKE_X25519_HKDF_SHA256_AES_256_GCM, pub, []byte("hello"), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		pt, err := decrypt(HPKE_X25519_HKDF_SHA256_AES_256_GCM, priv, ct, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if string(pt) != "hello" {
			t.Fatalf("expected 'hello', got %q", string(pt))
		}
	})

	t.Run("returns error for nil method", func(t *testing.T) {
		t.Parallel()

		_, err := decrypt(nil, nil, nil, nil)
		if !errors.Is(err, ErrMethodIsNil) {
			t.Fatalf("expected ErrMethodIsNil, got %v", err)
		}
	})

	t.Run("returns error for nil private key", func(t *testing.T) {
		t.Parallel()

		_, err := decrypt(HPKE_X25519_HKDF_SHA256_AES_256_GCM, nil, nil, nil)
		if !errors.Is(err, ErrPrivateKeyIsNil) {
			t.Fatalf("expected ErrPrivateKeyIsNil, got %v", err)
		}
	})

	t.Run("returns error for mismatched key scheme", func(t *testing.T) {
		t.Parallel()

		other := hpke.KEM_P256_HKDF_SHA256.Scheme()

		_, priv, err := other.GenerateKeyPair()
		if err != nil {
			t.Fatalf("unexpected error generating P-256 key: %v", err)
		}

		_, err = decrypt(HPKE_X25519_HKDF_SHA256_AES_256_GCM, priv, []byte("anything"), nil)
		if !errors.Is(err, ErrKeyTypeMismatch) {
			t.Fatalf("expected ErrKeyTypeMismatch, got %v", err)
		}
	})

	t.Run("returns error for ciphertext too short", func(t *testing.T) {
		t.Parallel()

		_, priv, err := generateKey(HPKE_X25519_HKDF_SHA256_AES_256_GCM)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err = decrypt(HPKE_X25519_HKDF_SHA256_AES_256_GCM, priv, []byte("short"), nil)
		if !errors.Is(err, ErrCiphertextTooShort) {
			t.Fatalf("expected ErrCiphertextTooShort, got %v", err)
		}
	})

	t.Run("returns error for tampered ciphertext", func(t *testing.T) {
		t.Parallel()

		pub, priv, err := generateKey(HPKE_X25519_HKDF_SHA256_AES_256_GCM)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ct, err := encrypt(HPKE_X25519_HKDF_SHA256_AES_256_GCM, pub, []byte("hello"), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ct[len(ct)-1] ^= 0xff

		_, err = decrypt(HPKE_X25519_HKDF_SHA256_AES_256_GCM, priv, ct, nil)
		if !errors.Is(err, ErrDecryptionFailed) {
			t.Fatalf("expected ErrDecryptionFailed, got %v", err)
		}
	})
}

func TestEncrypt_ByName(t *testing.T) {
	t.Parallel()

	const methodName = "HPKE_X25519_HKDF_SHA256_AES_256_GCM"

	t.Run("encrypts and decrypts round trip", func(t *testing.T) {
		t.Parallel()

		method, err := Get(methodName)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		pub, priv, err := method.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error generating key: %v", err)
		}

		pubBin, err := pub.MarshalBinary()
		if err != nil {
			t.Fatalf("unexpected error marshalling public key: %v", err)
		}

		privBin, err := priv.MarshalBinary()
		if err != nil {
			t.Fatalf("unexpected error marshalling private key: %v", err)
		}

		const plaintext = "round-trip"

		ciphered, err := Encrypt(methodName, pubBin, []byte(plaintext), []byte("info"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		plain, err := Decrypt(methodName, privBin, ciphered, []byte("info"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if string(plain) != plaintext {
			t.Fatalf("expected %q, got %q", plaintext, string(plain))
		}
	})

	t.Run("Encrypt returns domain error for unknown name", func(t *testing.T) {
		t.Parallel()

		_, err := Encrypt("UNKNOWN", []byte("k"), []byte("d"), nil)
		if err == nil {
			t.Fatal("expected error for unknown name")
		}

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}
	})

	t.Run("Decrypt returns domain error for unknown name", func(t *testing.T) {
		t.Parallel()

		_, err := Decrypt("UNKNOWN", []byte("k"), []byte("d"), nil)
		if err == nil {
			t.Fatal("expected error for unknown name")
		}

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}
	})

	t.Run("Encrypt returns key-type-mismatch error for invalid key", func(t *testing.T) {
		t.Parallel()

		_, err := Encrypt(methodName, []byte("not a kem key"), []byte("d"), nil)
		if err == nil {
			t.Fatal("expected error for invalid binary key")
		}

		if !errors.Is(err, ErrKeyTypeMismatch) {
			t.Fatalf("expected ErrKeyTypeMismatch, got %v", err)
		}
	})

	t.Run("Decrypt returns key-type-mismatch error for invalid key", func(t *testing.T) {
		t.Parallel()

		_, err := Decrypt(methodName, []byte("not a kem key"), []byte("d"), nil)
		if err == nil {
			t.Fatal("expected error for invalid binary key")
		}

		if !errors.Is(err, ErrKeyTypeMismatch) {
			t.Fatalf("expected ErrKeyTypeMismatch, got %v", err)
		}
	})
}
