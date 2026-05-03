package hmacs

import (
	"crypto"
	"testing"

	ctypes "github.com/guidomantilla/yarumo/common/types"
)

func TestNewMethod(t *testing.T) {
	t.Parallel()

	t.Run("creates method with defaults", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("test-hmac", crypto.SHA256, 32)

		if m == nil {
			t.Fatal("expected non-nil method")
		}

		if m.name != "test-hmac" {
			t.Fatalf("expected name 'test-hmac', got %q", m.name)
		}

		if m.keySize != 32 {
			t.Fatalf("expected keySize 32, got %d", m.keySize)
		}
	})

	t.Run("applies custom key function via option", func(t *testing.T) {
		t.Parallel()

		called := false
		custom := func(method *Method) (ctypes.Bytes, error) {
			called = true
			return []byte("key"), nil
		}

		m := NewMethod("custom", crypto.SHA256, 32, WithKeyFn(custom))

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

		m := NewMethod("my-hmac", crypto.SHA256, 32)

		got := m.Name()
		if got != "my-hmac" {
			t.Fatalf("expected 'my-hmac', got %q", got)
		}
	})
}

func TestMethod_GenerateKey(t *testing.T) {
	t.Parallel()

	t.Run("generates key of correct size", func(t *testing.T) {
		t.Parallel()

		key, err := HMAC_with_SHA256.GenerateKey()
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
			return nil, ErrMethodIsNil
		}

		m := NewMethod("fail", crypto.SHA256, 32, WithKeyFn(failKey))

		_, err := m.GenerateKey()
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestMethod_Digest(t *testing.T) {
	t.Parallel()

	t.Run("computes HMAC-SHA256 digest", func(t *testing.T) {
		t.Parallel()

		key, err := HMAC_with_SHA256.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error generating key: %v", err)
		}

		digest, err := HMAC_with_SHA256.Digest(key, []byte("hello"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(digest) != 32 {
			t.Fatalf("expected 32 bytes, got %d", len(digest))
		}
	})

	t.Run("produces deterministic output", func(t *testing.T) {
		t.Parallel()

		key := ctypes.Bytes("fixed-key-for-test-1234567890ab")

		a, err := HMAC_with_SHA256.Digest(key, []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		b, err := HMAC_with_SHA256.Digest(key, []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if a.ToHex() != b.ToHex() {
			t.Fatal("expected identical digests")
		}
	})

	t.Run("wraps error from digestFn", func(t *testing.T) {
		t.Parallel()

		failDigest := func(method *Method, key ctypes.Bytes, data ctypes.Bytes) (ctypes.Bytes, error) {
			return nil, ErrHashNotAvailable
		}

		m := NewMethod("fail-digest", crypto.SHA256, 32, WithDigestFn(failDigest))

		_, err := m.Digest([]byte("key"), []byte("data"))
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestMethod_Validate(t *testing.T) {
	t.Parallel()

	t.Run("validates correct digest", func(t *testing.T) {
		t.Parallel()

		key, err := HMAC_with_SHA256.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		digest, err := HMAC_with_SHA256.Digest(key, []byte("message"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ok, err := HMAC_with_SHA256.Validate(key, digest, []byte("message"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatal("expected validation to succeed")
		}
	})

	t.Run("rejects wrong digest", func(t *testing.T) {
		t.Parallel()

		key, err := HMAC_with_SHA256.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ok, err := HMAC_with_SHA256.Validate(key, []byte("wrong-digest"), []byte("message"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ok {
			t.Fatal("expected validation to fail")
		}
	})

	t.Run("wraps error from validateFn", func(t *testing.T) {
		t.Parallel()

		failValidate := func(method *Method, key ctypes.Bytes, sig ctypes.Bytes, data ctypes.Bytes) (bool, error) {
			return false, ErrHashNotAvailable
		}

		m := NewMethod("fail-validate", crypto.SHA256, 32, WithValidateFn(failValidate))

		_, err := m.Validate([]byte("key"), []byte("sig"), []byte("data"))
		if err == nil {
			t.Fatal("expected error")
		}
	})
}
