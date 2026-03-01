package hashes

import (
	"crypto"
	"testing"
)

func TestHash(t *testing.T) {
	t.Parallel()

	t.Run("computes SHA256 digest", func(t *testing.T) {
		t.Parallel()

		result := Hash(crypto.SHA256, []byte("hello"))

		if len(result) != 32 {
			t.Fatalf("expected 32 bytes, got %d", len(result))
		}
	})

	t.Run("computes SHA512 digest", func(t *testing.T) {
		t.Parallel()

		result := Hash(crypto.SHA512, []byte("hello"))

		if len(result) != 64 {
			t.Fatalf("expected 64 bytes, got %d", len(result))
		}
	})

	t.Run("produces deterministic output", func(t *testing.T) {
		t.Parallel()

		a := Hash(crypto.SHA256, []byte("test"))
		b := Hash(crypto.SHA256, []byte("test"))

		if a.ToHex() != b.ToHex() {
			t.Fatal("expected identical digests for same input")
		}
	})

	t.Run("handles empty input", func(t *testing.T) {
		t.Parallel()

		result := Hash(crypto.SHA256, []byte{})

		if len(result) != 32 {
			t.Fatalf("expected 32 bytes for empty input, got %d", len(result))
		}
	})

	t.Run("handles nil input", func(t *testing.T) {
		t.Parallel()

		result := Hash(crypto.SHA256, nil)

		if len(result) != 32 {
			t.Fatalf("expected 32 bytes for nil input, got %d", len(result))
		}
	})
}
