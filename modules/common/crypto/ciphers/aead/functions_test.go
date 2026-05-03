package aead

import (
	"crypto/cipher"
	"errors"
	"testing"

	ctypes "github.com/guidomantilla/yarumo/common/types"
)

func TestKey(t *testing.T) {
	t.Parallel()

	t.Run("generates key of correct size", func(t *testing.T) {
		t.Parallel()

		k, err := key(AES_256_GCM)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(k) != 32 {
			t.Fatalf("expected 32 bytes, got %d", len(k))
		}
	})

	t.Run("returns error for nil method", func(t *testing.T) {
		t.Parallel()

		_, err := key(nil)
		if !errors.Is(err, ErrMethodInvalid) {
			t.Fatalf("expected ErrMethodInvalid, got %v", err)
		}
	})

	t.Run("returns error for invalid key size", func(t *testing.T) {
		t.Parallel()

		m := &Method{keySize: 7, nonceSize: 12}

		_, err := key(m)
		if !errors.Is(err, ErrKeySizeInvalid) {
			t.Fatalf("expected ErrKeySizeInvalid, got %v", err)
		}
	})

	t.Run("returns error for invalid nonce size", func(t *testing.T) {
		t.Parallel()

		m := &Method{keySize: 32, nonceSize: 8}

		_, err := key(m)
		if !errors.Is(err, ErrNonceSizeInvalid) {
			t.Fatalf("expected ErrNonceSizeInvalid, got %v", err)
		}
	})
}

func TestEncrypt(t *testing.T) {
	t.Parallel()

	t.Run("encrypts data", func(t *testing.T) {
		t.Parallel()

		k, err := key(AES_256_GCM)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ciphered, err := encrypt(AES_256_GCM, k, []byte("hello"), nil)
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
		if !errors.Is(err, ErrMethodInvalid) {
			t.Fatalf("expected ErrMethodInvalid, got %v", err)
		}
	})

	t.Run("returns error for invalid key length", func(t *testing.T) {
		t.Parallel()

		_, err := encrypt(AES_256_GCM, []byte("short"), []byte("data"), nil)
		if !errors.Is(err, ErrKeyInvalid) {
			t.Fatalf("expected ErrKeyInvalid, got %v", err)
		}
	})

	t.Run("returns error for cipher init failure", func(t *testing.T) {
		t.Parallel()

		badFn := func(_ ctypes.Bytes, _ int) (cipher.AEAD, error) {
			return nil, errors.New("init failed")
		}
		m := NewMethod("bad", 32, 12, badFn)

		k := make([]byte, 32)

		_, err := encrypt(m, k, []byte("data"), nil)
		if !errors.Is(err, ErrCipherInitFailed) {
			t.Fatalf("expected ErrCipherInitFailed, got %v", err)
		}
	})
}

func TestDecrypt(t *testing.T) {
	t.Parallel()

	t.Run("decrypts data", func(t *testing.T) {
		t.Parallel()

		k, err := key(AES_256_GCM)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ciphered, err := encrypt(AES_256_GCM, k, []byte("hello"), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		plain, err := decrypt(AES_256_GCM, k, ciphered, nil)
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
		if !errors.Is(err, ErrMethodInvalid) {
			t.Fatalf("expected ErrMethodInvalid, got %v", err)
		}
	})

	t.Run("returns error for invalid key length", func(t *testing.T) {
		t.Parallel()

		_, err := decrypt(AES_256_GCM, []byte("short"), []byte("ciphered-data-here"), nil)
		if !errors.Is(err, ErrKeyInvalid) {
			t.Fatalf("expected ErrKeyInvalid, got %v", err)
		}
	})

	t.Run("returns error for ciphertext too short", func(t *testing.T) {
		t.Parallel()

		k := make([]byte, 32)

		_, err := decrypt(AES_256_GCM, k, []byte("short"), nil)
		if !errors.Is(err, ErrCiphertextTooShort) {
			t.Fatalf("expected ErrCiphertextTooShort, got %v", err)
		}
	})

	t.Run("returns error for cipher init failure", func(t *testing.T) {
		t.Parallel()

		badFn := func(_ ctypes.Bytes, _ int) (cipher.AEAD, error) {
			return nil, errors.New("init failed")
		}
		m := NewMethod("bad", 32, 12, badFn)

		k := make([]byte, 32)
		ciphered := make([]byte, 20) // longer than nonceSize

		_, err := decrypt(m, k, ciphered, nil)
		if !errors.Is(err, ErrCipherInitFailed) {
			t.Fatalf("expected ErrCipherInitFailed, got %v", err)
		}
	})

	t.Run("returns error for tampered ciphertext", func(t *testing.T) {
		t.Parallel()

		k, err := key(AES_256_GCM)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ciphered, err := encrypt(AES_256_GCM, k, []byte("hello"), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Tamper with the ciphertext (after nonce)
		ciphered[len(ciphered)-1] ^= 0xff

		_, err = decrypt(AES_256_GCM, k, ciphered, nil)
		if !errors.Is(err, ErrDecryptFailed) {
			t.Fatalf("expected ErrDecryptFailed, got %v", err)
		}
	})
}
