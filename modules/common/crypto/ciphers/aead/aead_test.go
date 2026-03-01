package aead

import (
	"testing"

	ctypes "github.com/guidomantilla/yarumo/common/types"
)

func TestNewMethod(t *testing.T) {
	t.Parallel()

	t.Run("creates method with defaults", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("test-aead", 32, 12, aesgcm)

		if m == nil {
			t.Fatal("expected non-nil method")
		}

		if m.name != "test-aead" {
			t.Fatalf("expected 'test-aead', got %q", m.name)
		}

		if m.keySize != 32 {
			t.Fatalf("expected keySize 32, got %d", m.keySize)
		}

		if m.nonceSize != 12 {
			t.Fatalf("expected nonceSize 12, got %d", m.nonceSize)
		}
	})

	t.Run("applies custom key function via option", func(t *testing.T) {
		t.Parallel()

		called := false
		custom := func(method *Method) (ctypes.Bytes, error) {
			called = true
			return []byte("01234567890123456789012345678901"), nil
		}

		m := NewMethod("custom", 32, 12, aesgcm, WithKeyFn(custom))

		_, err := m.GenerateKey()
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

		m := NewMethod("my-aead", 32, 12, aesgcm)

		got := m.Name()
		if got != "my-aead" {
			t.Fatalf("expected 'my-aead', got %q", got)
		}
	})
}

func TestMethod_GenerateKey(t *testing.T) {
	t.Parallel()

	t.Run("generates key of correct size", func(t *testing.T) {
		t.Parallel()

		key, err := AES_256_GCM.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(key) != 32 {
			t.Fatalf("expected 32 bytes, got %d", len(key))
		}
	})

	t.Run("wraps error from keyFn", func(t *testing.T) {
		t.Parallel()

		failKey := func(method *Method) (ctypes.Bytes, error) {
			return nil, ErrMethodInvalid
		}

		m := NewMethod("fail", 32, 12, aesgcm, WithKeyFn(failKey))

		_, err := m.GenerateKey()
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestMethod_Encrypt(t *testing.T) {
	t.Parallel()

	t.Run("encrypts data", func(t *testing.T) {
		t.Parallel()

		key, err := AES_256_GCM.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ciphered, err := AES_256_GCM.Encrypt(key, []byte("hello"), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(ciphered) == 0 {
			t.Fatal("expected non-empty ciphertext")
		}
	})

	t.Run("wraps error from encryptFn", func(t *testing.T) {
		t.Parallel()

		failEncrypt := func(method *Method, key ctypes.Bytes, data ctypes.Bytes, aad ctypes.Bytes) (ctypes.Bytes, error) {
			return nil, ErrKeyInvalid
		}

		m := NewMethod("fail-encrypt", 32, 12, aesgcm, WithEncryptFn(failEncrypt))

		_, err := m.Encrypt([]byte("key"), []byte("data"), nil)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestMethod_Decrypt(t *testing.T) {
	t.Parallel()

	t.Run("decrypts data", func(t *testing.T) {
		t.Parallel()

		key, err := AES_256_GCM.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ciphered, err := AES_256_GCM.Encrypt(key, []byte("hello"), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		plain, err := AES_256_GCM.Decrypt(key, ciphered, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if string(plain) != "hello" {
			t.Fatalf("expected 'hello', got %q", string(plain))
		}
	})

	t.Run("encrypts and decrypts with AAD", func(t *testing.T) {
		t.Parallel()

		key, err := AES_256_GCM.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		aad := []byte("additional-data")

		ciphered, err := AES_256_GCM.Encrypt(key, []byte("secret"), aad)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		plain, err := AES_256_GCM.Decrypt(key, ciphered, aad)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if string(plain) != "secret" {
			t.Fatalf("expected 'secret', got %q", string(plain))
		}
	})

	t.Run("wraps error from decryptFn", func(t *testing.T) {
		t.Parallel()

		failDecrypt := func(method *Method, key ctypes.Bytes, ciphered ctypes.Bytes, aad ctypes.Bytes) (ctypes.Bytes, error) {
			return nil, ErrCiphertextTooShort
		}

		m := NewMethod("fail-decrypt", 32, 12, aesgcm, WithDecryptFn(failDecrypt))

		_, err := m.Decrypt([]byte("key"), []byte("ciphered"), nil)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestMethod_AllCipherVariants(t *testing.T) {
	t.Parallel()

	t.Run("AES-128-GCM encrypt and decrypt", func(t *testing.T) {
		t.Parallel()

		key, err := AES_128_GCM.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ciphered, err := AES_128_GCM.Encrypt(key, []byte("aes128-payload"), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		plain, err := AES_128_GCM.Decrypt(key, ciphered, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if string(plain) != "aes128-payload" {
			t.Fatalf("expected 'aes128-payload', got %q", string(plain))
		}
	})

	t.Run("ChaCha20-Poly1305 encrypt and decrypt", func(t *testing.T) {
		t.Parallel()

		key, err := CHACHA20_POLY1305.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ciphered, err := CHACHA20_POLY1305.Encrypt(key, []byte("chacha-payload"), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		plain, err := CHACHA20_POLY1305.Decrypt(key, ciphered, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if string(plain) != "chacha-payload" {
			t.Fatalf("expected 'chacha-payload', got %q", string(plain))
		}
	})

	t.Run("XChaCha20-Poly1305 encrypt and decrypt", func(t *testing.T) {
		t.Parallel()

		key, err := XCHACHA20_POLY1305.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ciphered, err := XCHACHA20_POLY1305.Encrypt(key, []byte("xchacha-payload"), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		plain, err := XCHACHA20_POLY1305.Decrypt(key, ciphered, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if string(plain) != "xchacha-payload" {
			t.Fatalf("expected 'xchacha-payload', got %q", string(plain))
		}
	})
}
