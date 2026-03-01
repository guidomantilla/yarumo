package hashes

import (
	"crypto"
	"testing"

	ctypes "github.com/guidomantilla/yarumo/common/types"
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
		custom := func(hash crypto.Hash, data ctypes.Bytes) ctypes.Bytes {
			called = true
			return data
		}

		m := NewMethod("custom", crypto.SHA512, WithHashFn(custom))

		result := m.Hash([]byte("data"))

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

	t.Run("computes SHA256 digest", func(t *testing.T) {
		t.Parallel()

		result := SHA256.Hash([]byte("hello"))

		if len(result) != 32 {
			t.Fatalf("expected 32 bytes for SHA256, got %d", len(result))
		}
	})

	t.Run("computes SHA512 digest", func(t *testing.T) {
		t.Parallel()

		result := SHA512.Hash([]byte("hello"))

		if len(result) != 64 {
			t.Fatalf("expected 64 bytes for SHA512, got %d", len(result))
		}
	})

	t.Run("computes SHA3_256 digest", func(t *testing.T) {
		t.Parallel()

		result := SHA3_256.Hash([]byte("hello"))

		if len(result) != 32 {
			t.Fatalf("expected 32 bytes for SHA3_256, got %d", len(result))
		}
	})

	t.Run("computes SHA3_512 digest", func(t *testing.T) {
		t.Parallel()

		result := SHA3_512.Hash([]byte("hello"))

		if len(result) != 64 {
			t.Fatalf("expected 64 bytes for SHA3_512, got %d", len(result))
		}
	})

	t.Run("computes BLAKE2b_256 digest", func(t *testing.T) {
		t.Parallel()

		result := BLAKE2b_256.Hash([]byte("hello"))

		if len(result) != 32 {
			t.Fatalf("expected 32 bytes for BLAKE2b_256, got %d", len(result))
		}
	})

	t.Run("computes BLAKE2b_512 digest", func(t *testing.T) {
		t.Parallel()

		result := BLAKE2b_512.Hash([]byte("hello"))

		if len(result) != 64 {
			t.Fatalf("expected 64 bytes for BLAKE2b_512, got %d", len(result))
		}
	})

	t.Run("produces deterministic output", func(t *testing.T) {
		t.Parallel()

		a := SHA256.Hash([]byte("test"))
		b := SHA256.Hash([]byte("test"))

		if a.ToHex() != b.ToHex() {
			t.Fatal("expected identical digests for same input")
		}
	})

	t.Run("produces different output for different input", func(t *testing.T) {
		t.Parallel()

		a := SHA256.Hash([]byte("hello"))
		b := SHA256.Hash([]byte("world"))

		if a.ToHex() == b.ToHex() {
			t.Fatal("expected different digests for different input")
		}
	})

	t.Run("handles empty input", func(t *testing.T) {
		t.Parallel()

		result := SHA256.Hash([]byte{})

		if len(result) != 32 {
			t.Fatalf("expected 32 bytes for empty input, got %d", len(result))
		}
	})

	t.Run("handles nil input", func(t *testing.T) {
		t.Parallel()

		result := SHA256.Hash(nil)

		if len(result) != 32 {
			t.Fatalf("expected 32 bytes for nil input, got %d", len(result))
		}
	})
}
