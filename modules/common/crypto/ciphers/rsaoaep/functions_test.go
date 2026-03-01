package rsaoaep

import (
	"crypto"
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
