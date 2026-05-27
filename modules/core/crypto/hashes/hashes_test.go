package hashes

import (
	"crypto"
	"errors"
	"testing"

	ctypes "github.com/guidomantilla/yarumo/core/common/types"
)

func TestNewMethod(t *testing.T) {
	t.Parallel()

	t.Run("creates method with default hash function", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("test", crypto.SHA256)

		if m == nil {
			t.Fatal("expected non-nil method")
		}

		if m.name != "test" {
			t.Fatalf("expected name 'test', got %q", m.name)
		}

		if m.kind != crypto.SHA256 {
			t.Fatalf("expected kind crypto.SHA256, got %v", m.kind)
		}

		if m.hashFn == nil {
			t.Fatal("expected hashFn to be set")
		}
	})

	t.Run("applies custom hash function via option", func(t *testing.T) {
		t.Parallel()

		called := false
		custom := func(hash crypto.Hash, data ctypes.Bytes) (ctypes.Bytes, error) {
			called = true
			return data, nil
		}

		m := NewMethod("custom", crypto.SHA512, WithHashFn(custom))

		result, err := m.Hash([]byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !called {
			t.Fatal("expected custom hashFn to be called")
		}

		if string(result) != "data" {
			t.Fatalf("expected 'data', got %q", string(result))
		}
	})
}

func TestMethod_Name(t *testing.T) {
	t.Parallel()

	t.Run("returns the method name", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("my-hash", crypto.SHA256)

		got := m.Name()
		if got != "my-hash" {
			t.Fatalf("expected 'my-hash', got %q", got)
		}
	})
}

func TestMethod_Hash(t *testing.T) {
	t.Parallel()

	t.Run("computes SHA1 digest", func(t *testing.T) {
		t.Parallel()

		result, err := SHA1.Hash([]byte("hello"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result) != 20 {
			t.Fatalf("expected 20 bytes for SHA1, got %d", len(result))
		}
	})

	t.Run("computes SHA224 digest", func(t *testing.T) {
		t.Parallel()

		result, err := SHA224.Hash([]byte("hello"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result) != 28 {
			t.Fatalf("expected 28 bytes for SHA224, got %d", len(result))
		}
	})

	t.Run("computes SHA256 digest", func(t *testing.T) {
		t.Parallel()

		result, err := SHA256.Hash([]byte("hello"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result) != 32 {
			t.Fatalf("expected 32 bytes for SHA256, got %d", len(result))
		}
	})

	t.Run("computes SHA384 digest", func(t *testing.T) {
		t.Parallel()

		result, err := SHA384.Hash([]byte("hello"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result) != 48 {
			t.Fatalf("expected 48 bytes for SHA384, got %d", len(result))
		}
	})

	t.Run("computes SHA512 digest", func(t *testing.T) {
		t.Parallel()

		result, err := SHA512.Hash([]byte("hello"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result) != 64 {
			t.Fatalf("expected 64 bytes for SHA512, got %d", len(result))
		}
	})

	t.Run("computes SHA3_256 digest", func(t *testing.T) {
		t.Parallel()

		result, err := SHA3_256.Hash([]byte("hello"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result) != 32 {
			t.Fatalf("expected 32 bytes for SHA3_256, got %d", len(result))
		}
	})

	t.Run("computes SHA3_384 digest", func(t *testing.T) {
		t.Parallel()

		result, err := SHA3_384.Hash([]byte("hello"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result) != 48 {
			t.Fatalf("expected 48 bytes for SHA3_384, got %d", len(result))
		}
	})

	t.Run("computes SHA3_512 digest", func(t *testing.T) {
		t.Parallel()

		result, err := SHA3_512.Hash([]byte("hello"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result) != 64 {
			t.Fatalf("expected 64 bytes for SHA3_512, got %d", len(result))
		}
	})

	t.Run("computes BLAKE2b_256 digest", func(t *testing.T) {
		t.Parallel()

		result, err := BLAKE2b_256.Hash([]byte("hello"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result) != 32 {
			t.Fatalf("expected 32 bytes for BLAKE2b_256, got %d", len(result))
		}
	})

	t.Run("computes BLAKE2b_512 digest", func(t *testing.T) {
		t.Parallel()

		result, err := BLAKE2b_512.Hash([]byte("hello"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result) != 64 {
			t.Fatalf("expected 64 bytes for BLAKE2b_512, got %d", len(result))
		}
	})

	t.Run("produces deterministic output", func(t *testing.T) {
		t.Parallel()

		a, err := SHA256.Hash([]byte("test"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		b, err := SHA256.Hash([]byte("test"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if a.ToHex() != b.ToHex() {
			t.Fatal("expected identical digests for same input")
		}
	})

	t.Run("produces different output for different input", func(t *testing.T) {
		t.Parallel()

		a, err := SHA256.Hash([]byte("hello"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		b, err := SHA256.Hash([]byte("world"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if a.ToHex() == b.ToHex() {
			t.Fatal("expected different digests for different input")
		}
	})

	t.Run("handles empty input", func(t *testing.T) {
		t.Parallel()

		result, err := SHA256.Hash([]byte{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result) != 32 {
			t.Fatalf("expected 32 bytes for empty input, got %d", len(result))
		}
	})

	t.Run("handles nil input", func(t *testing.T) {
		t.Parallel()

		result, err := SHA256.Hash(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result) != 32 {
			t.Fatalf("expected 32 bytes for nil input, got %d", len(result))
		}
	})

	t.Run("returns error when crypto.Hash driver is unavailable", func(t *testing.T) {
		t.Parallel()

		// crypto.Hash(99) is not a registered hash function — the driver was
		// not blank-imported, so hash.Available() returns false.
		m := NewMethod("unavailable", crypto.Hash(99))

		result, err := m.Hash([]byte("data"))
		if result != nil {
			t.Fatalf("expected nil result, got %v", result)
		}

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, ErrHashFunctionUnavailable) {
			t.Fatalf("expected errors.Is to match ErrHashFunctionUnavailable, got %v", err)
		}

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error via errors.As, got %T", err)
		}

		if domErr.Type != HashUnavailable {
			t.Fatalf("expected type %q, got %q", HashUnavailable, domErr.Type)
		}
	})
}
