package rsaoaep

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"testing"
)

func TestKeyFn(t *testing.T) {
	t.Parallel()

	t.Run("generates RSA key", func(t *testing.T) {
		t.Parallel()

		k, err := key(RSA_OAEP_SHA256, 2048)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if k.N.BitLen() != 2048 {
			t.Fatalf("expected 2048-bit key, got %d", k.N.BitLen())
		}
	})

	t.Run("returns error for nil method", func(t *testing.T) {
		t.Parallel()

		_, err := key(nil, 2048)
		if !errors.Is(err, ErrMethodIsNil) {
			t.Fatalf("expected ErrMethodIsNil, got %v", err)
		}
	})

	t.Run("returns error for invalid key size", func(t *testing.T) {
		t.Parallel()

		_, err := key(RSA_OAEP_SHA256, 1024)
		if !errors.Is(err, ErrKeyLengthIsInvalid) {
			t.Fatalf("expected ErrKeyLengthIsInvalid, got %v", err)
		}
	})
}

func TestEncryptFn(t *testing.T) {
	t.Parallel()

	t.Run("encrypts data", func(t *testing.T) {
		t.Parallel()

		k, err := key(RSA_OAEP_SHA256, 2048)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ciphered, err := encrypt(RSA_OAEP_SHA256, &k.PublicKey, []byte("hello"), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(ciphered) == 0 {
			t.Fatal("expected non-empty ciphertext")
		}
	})

	t.Run("returns error for nil method", func(t *testing.T) {
		t.Parallel()

		_, err := encrypt(nil, nil, nil, nil)
		if !errors.Is(err, ErrMethodIsNil) {
			t.Fatalf("expected ErrMethodIsNil, got %v", err)
		}
	})

	t.Run("returns error for nil key", func(t *testing.T) {
		t.Parallel()

		_, err := encrypt(RSA_OAEP_SHA256, nil, []byte("data"), nil)
		if !errors.Is(err, ErrKeyIsNil) {
			t.Fatalf("expected ErrKeyIsNil, got %v", err)
		}
	})

	t.Run("returns error for unavailable hash", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("bad-hash", crypto.Hash(0), []int{2048})

		k, err := key(RSA_OAEP_SHA256, 2048)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err = encrypt(m, &k.PublicKey, []byte("data"), nil)
		if !errors.Is(err, ErrHashNotAvailable) {
			t.Fatalf("expected ErrHashNotAvailable, got %v", err)
		}
	})

	t.Run("returns error for invalid key size", func(t *testing.T) {
		t.Parallel()

		k, err := key(RSA_OAEP_SHA256, 2048)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Use a method that only allows 4096
		m := NewMethod("strict", crypto.SHA256, []int{4096})

		_, err = encrypt(m, &k.PublicKey, []byte("data"), nil)
		if !errors.Is(err, ErrKeyLengthIsInvalid) {
			t.Fatalf("expected ErrKeyLengthIsInvalid, got %v", err)
		}
	})
}

func TestDecryptFn(t *testing.T) {
	t.Parallel()

	t.Run("decrypts data", func(t *testing.T) {
		t.Parallel()

		k, err := key(RSA_OAEP_SHA256, 2048)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ciphered, err := encrypt(RSA_OAEP_SHA256, &k.PublicKey, []byte("hello"), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		plain, err := decrypt(RSA_OAEP_SHA256, k, ciphered, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if string(plain) != "hello" {
			t.Fatalf("expected 'hello', got %q", string(plain))
		}
	})

	t.Run("returns error for nil method", func(t *testing.T) {
		t.Parallel()

		_, err := decrypt(nil, nil, nil, nil)
		if !errors.Is(err, ErrMethodIsNil) {
			t.Fatalf("expected ErrMethodIsNil, got %v", err)
		}
	})

	t.Run("returns error for nil key", func(t *testing.T) {
		t.Parallel()

		_, err := decrypt(RSA_OAEP_SHA256, nil, []byte("data"), nil)
		if !errors.Is(err, ErrKeyIsNil) {
			t.Fatalf("expected ErrKeyIsNil, got %v", err)
		}
	})

	t.Run("returns error for unavailable hash", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("bad-hash", crypto.Hash(0), []int{2048})

		k, err := key(RSA_OAEP_SHA256, 2048)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err = decrypt(m, k, []byte("data"), nil)
		if !errors.Is(err, ErrHashNotAvailable) {
			t.Fatalf("expected ErrHashNotAvailable, got %v", err)
		}
	})

	t.Run("returns error for invalid key size", func(t *testing.T) {
		t.Parallel()

		k, err := key(RSA_OAEP_SHA256, 2048)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Use a method that only allows 4096
		m := NewMethod("strict", crypto.SHA256, []int{4096})

		_, err = decrypt(m, k, []byte("data"), nil)
		if !errors.Is(err, ErrKeyLengthIsInvalid) {
			t.Fatalf("expected ErrKeyLengthIsInvalid, got %v", err)
		}
	})

	t.Run("returns error for tampered ciphertext", func(t *testing.T) {
		t.Parallel()

		k, err := key(RSA_OAEP_SHA256, 2048)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err = decrypt(RSA_OAEP_SHA256, k, []byte("invalid-ciphertext"), nil)
		if !errors.Is(err, ErrDecryptionFailed) {
			t.Fatalf("expected ErrDecryptionFailed, got %v", err)
		}
	})
}

func TestMarshalPrivateKeyPEM(t *testing.T) {
	t.Parallel()

	t.Run("returns error for nil key", func(t *testing.T) {
		t.Parallel()

		_, err := MarshalPrivateKeyPEM(nil)

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}

		if !errors.Is(err, ErrKeyIsNil) {
			t.Fatalf("expected ErrKeyIsNil, got %v", err)
		}
	})
}

func TestParsePrivateKeyPEM(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrPEMDecodeFailed on malformed PEM", func(t *testing.T) {
		t.Parallel()

		_, err := ParsePrivateKeyPEM([]byte("not a pem"))

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}

		if !errors.Is(err, ErrPEMDecodeFailed) {
			t.Fatalf("expected ErrPEMDecodeFailed, got %v", err)
		}
	})

	t.Run("returns ErrPEMBlockTypeMismatch on wrong block type", func(t *testing.T) {
		t.Parallel()

		block := &pem.Block{Type: "CERTIFICATE", Bytes: []byte("dummy")}

		_, err := ParsePrivateKeyPEM(pem.EncodeToMemory(block))
		if !errors.Is(err, ErrPEMBlockTypeMismatch) {
			t.Fatalf("expected ErrPEMBlockTypeMismatch, got %v", err)
		}
	})

	t.Run("returns ErrKeyTypeMismatch when parsing ECDSA key as RSA", func(t *testing.T) {
		t.Parallel()

		ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		der, err := x509.MarshalPKCS8PrivateKey(ecKey)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ecPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})

		_, err = ParsePrivateKeyPEM(ecPEM)
		if !errors.Is(err, ErrKeyTypeMismatch) {
			t.Fatalf("expected ErrKeyTypeMismatch, got %v", err)
		}
	})
}

func TestPrivateKeyPEMRoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("round-trips an RSA private key and decrypts", func(t *testing.T) {
		t.Parallel()

		orig, err := RSA_OAEP_SHA256.GenerateKey(2048)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		pemBytes, err := MarshalPrivateKeyPEM(orig)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		parsed, err := ParsePrivateKeyPEM(pemBytes)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ciphered, err := RSA_OAEP_SHA256.Encrypt(&parsed.PublicKey, []byte("hello"), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		plain, err := RSA_OAEP_SHA256.Decrypt(parsed, ciphered, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if string(plain) != "hello" {
			t.Fatalf("expected 'hello', got %q", string(plain))
		}
	})
}

func TestMarshalPublicKeyPEM(t *testing.T) {
	t.Parallel()

	t.Run("returns error for nil key", func(t *testing.T) {
		t.Parallel()

		_, err := MarshalPublicKeyPEM(nil)
		if !errors.Is(err, ErrKeyIsNil) {
			t.Fatalf("expected ErrKeyIsNil, got %v", err)
		}
	})
}

func TestParsePublicKeyPEM(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrPEMDecodeFailed on malformed PEM", func(t *testing.T) {
		t.Parallel()

		_, err := ParsePublicKeyPEM([]byte("garbage"))
		if !errors.Is(err, ErrPEMDecodeFailed) {
			t.Fatalf("expected ErrPEMDecodeFailed, got %v", err)
		}
	})

	t.Run("returns ErrPEMBlockTypeMismatch on wrong block type", func(t *testing.T) {
		t.Parallel()

		block := &pem.Block{Type: "CERTIFICATE", Bytes: []byte("dummy")}

		_, err := ParsePublicKeyPEM(pem.EncodeToMemory(block))
		if !errors.Is(err, ErrPEMBlockTypeMismatch) {
			t.Fatalf("expected ErrPEMBlockTypeMismatch, got %v", err)
		}
	})

	t.Run("returns ErrKeyTypeMismatch when parsing ECDSA public key as RSA", func(t *testing.T) {
		t.Parallel()

		ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		der, err := x509.MarshalPKIXPublicKey(&ecKey.PublicKey)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ecPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})

		_, err = ParsePublicKeyPEM(ecPEM)
		if !errors.Is(err, ErrKeyTypeMismatch) {
			t.Fatalf("expected ErrKeyTypeMismatch, got %v", err)
		}
	})
}

func TestPublicKeyPEMRoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("round-trips an RSA public key and encrypts", func(t *testing.T) {
		t.Parallel()

		priv, err := RSA_OAEP_SHA256.GenerateKey(2048)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		pubPEM, err := MarshalPublicKeyPEM(&priv.PublicKey)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		parsedPub, err := ParsePublicKeyPEM(pubPEM)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ciphered, err := RSA_OAEP_SHA256.Encrypt(parsedPub, []byte("hello"), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		plain, err := RSA_OAEP_SHA256.Decrypt(priv, ciphered, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if string(plain) != "hello" {
			t.Fatalf("expected 'hello', got %q", string(plain))
		}
	})
}

func TestEncrypt_ByName(t *testing.T) {
	t.Parallel()

	t.Run("encrypts and decrypts round trip", func(t *testing.T) {
		t.Parallel()

		method, err := Get("RSA-OAEP-SHA256")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		priv, err := method.GenerateKey(2048)
		if err != nil {
			t.Fatalf("unexpected error generating key: %v", err)
		}

		privPEM, err := MarshalPrivateKeyPEM(priv)
		if err != nil {
			t.Fatalf("unexpected error marshalling private key: %v", err)
		}

		pubPEM, err := MarshalPublicKeyPEM(&priv.PublicKey)
		if err != nil {
			t.Fatalf("unexpected error marshalling public key: %v", err)
		}

		const plaintext = "round-trip"

		ciphered, err := Encrypt("RSA-OAEP-SHA256", pubPEM, []byte(plaintext), []byte("label"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		plain, err := Decrypt("RSA-OAEP-SHA256", privPEM, ciphered, []byte("label"))
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

	t.Run("Encrypt returns PEM codec error for invalid key", func(t *testing.T) {
		t.Parallel()

		_, err := Encrypt("RSA-OAEP-SHA256", []byte("not a pem"), []byte("d"), nil)
		if err == nil {
			t.Fatal("expected error for invalid PEM")
		}

		if !errors.Is(err, ErrPEMDecodeFailed) {
			t.Fatalf("expected ErrPEMDecodeFailed, got %v", err)
		}
	})

	t.Run("Decrypt returns PEM codec error for invalid key", func(t *testing.T) {
		t.Parallel()

		_, err := Decrypt("RSA-OAEP-SHA256", []byte("not a pem"), []byte("d"), nil)
		if err == nil {
			t.Fatal("expected error for invalid PEM")
		}

		if !errors.Is(err, ErrPEMDecodeFailed) {
			t.Fatalf("expected ErrPEMDecodeFailed, got %v", err)
		}
	})
}
