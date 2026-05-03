package rsaoaep

import (
	"crypto"
	"crypto/rsa"
	"testing"

	ctypes "github.com/guidomantilla/yarumo/common/types"
)

func TestNewMethod(t *testing.T) {
	t.Parallel()

	t.Run("creates method with defaults", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("test-rsaoaep", crypto.SHA256, []int{2048})

		if m == nil {
			t.Fatal("expected non-nil method")
		}

		if m.name != "test-rsaoaep" {
			t.Fatalf("expected 'test-rsaoaep', got %q", m.name)
		}
	})
}

func TestMethod_Name(t *testing.T) {
	t.Parallel()

	t.Run("returns the method name", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("my-rsaoaep", crypto.SHA256, []int{2048})

		got := m.Name()
		if got != "my-rsaoaep" {
			t.Fatalf("expected 'my-rsaoaep', got %q", got)
		}
	})
}

func TestMethod_GenerateKey(t *testing.T) {
	t.Parallel()

	t.Run("generates RSA key", func(t *testing.T) {
		t.Parallel()

		key, err := RSA_OAEP_SHA256.GenerateKey(2048)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if key == nil {
			t.Fatal("expected non-nil key")
		}

		if key.N.BitLen() != 2048 {
			t.Fatalf("expected 2048-bit key, got %d", key.N.BitLen())
		}
	})

	t.Run("rejects disallowed key size", func(t *testing.T) {
		t.Parallel()

		_, err := RSA_OAEP_SHA256.GenerateKey(1024)
		if err == nil {
			t.Fatal("expected error for disallowed key size")
		}
	})

	t.Run("wraps error from keyFn", func(t *testing.T) {
		t.Parallel()

		failKey := func(method *Method, bits int) (*rsa.PrivateKey, error) {
			return nil, ErrMethodIsNil
		}

		m := NewMethod("fail", crypto.SHA256, []int{2048}, WithKeyFn(failKey))

		_, err := m.GenerateKey(2048)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestMethod_Encrypt(t *testing.T) {
	t.Parallel()

	t.Run("encrypts data", func(t *testing.T) {
		t.Parallel()

		key, err := RSA_OAEP_SHA256.GenerateKey(2048)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ciphered, err := RSA_OAEP_SHA256.Encrypt(&key.PublicKey, []byte("hello"), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(ciphered) == 0 {
			t.Fatal("expected non-empty ciphertext")
		}
	})

	t.Run("wraps error from encryptFn", func(t *testing.T) {
		t.Parallel()

		failEncrypt := func(method *Method, key *rsa.PublicKey, data, label ctypes.Bytes) (ctypes.Bytes, error) {
			return nil, ErrKeyIsNil
		}

		m := NewMethod("fail-encrypt", crypto.SHA256, []int{2048}, WithEncryptFn(failEncrypt))

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

		key, err := RSA_OAEP_SHA256.GenerateKey(2048)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ciphered, err := RSA_OAEP_SHA256.Encrypt(&key.PublicKey, []byte("hello"), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		plain, err := RSA_OAEP_SHA256.Decrypt(key, ciphered, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if string(plain) != "hello" {
			t.Fatalf("expected 'hello', got %q", string(plain))
		}
	})

	t.Run("encrypts and decrypts with label", func(t *testing.T) {
		t.Parallel()

		key, err := RSA_OAEP_SHA256.GenerateKey(2048)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		label := []byte("my-label")

		ciphered, err := RSA_OAEP_SHA256.Encrypt(&key.PublicKey, []byte("secret"), label)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		plain, err := RSA_OAEP_SHA256.Decrypt(key, ciphered, label)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if string(plain) != "secret" {
			t.Fatalf("expected 'secret', got %q", string(plain))
		}
	})

	t.Run("wraps error from decryptFn", func(t *testing.T) {
		t.Parallel()

		failDecrypt := func(method *Method, key *rsa.PrivateKey, ciphered, label ctypes.Bytes) (ctypes.Bytes, error) {
			return nil, ErrKeyIsNil
		}

		m := NewMethod("fail-decrypt", crypto.SHA256, []int{2048}, WithDecryptFn(failDecrypt))

		_, err := m.Decrypt(nil, []byte("data"), nil)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestMethod_SHA512Variant(t *testing.T) {
	t.Parallel()

	t.Run("RSA-OAEP-SHA512 encrypt and decrypt", func(t *testing.T) {
		t.Parallel()

		key, err := RSA_OAEP_SHA512.GenerateKey(4096)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ciphered, err := RSA_OAEP_SHA512.Encrypt(&key.PublicKey, []byte("sha512-payload"), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		plain, err := RSA_OAEP_SHA512.Decrypt(key, ciphered, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if string(plain) != "sha512-payload" {
			t.Fatalf("expected 'sha512-payload', got %q", string(plain))
		}
	})
}
