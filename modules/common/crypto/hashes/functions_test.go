package hashes

import (
	"crypto"
	"errors"
	"testing"
)

func TestHash(t *testing.T) {
	t.Parallel()

	t.Run("computes SHA256 digest", func(t *testing.T) {
		t.Parallel()

		result, err := Hash(crypto.SHA256, []byte("hello"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result) != 32 {
			t.Fatalf("expected 32 bytes, got %d", len(result))
		}
	})

	t.Run("computes SHA512 digest", func(t *testing.T) {
		t.Parallel()

		result, err := Hash(crypto.SHA512, []byte("hello"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result) != 64 {
			t.Fatalf("expected 64 bytes, got %d", len(result))
		}
	})

	t.Run("produces deterministic output", func(t *testing.T) {
		t.Parallel()

		a, err := Hash(crypto.SHA256, []byte("test"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		b, err := Hash(crypto.SHA256, []byte("test"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if a.ToHex() != b.ToHex() {
			t.Fatal("expected identical digests for same input")
		}
	})

	t.Run("handles empty input", func(t *testing.T) {
		t.Parallel()

		result, err := Hash(crypto.SHA256, []byte{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result) != 32 {
			t.Fatalf("expected 32 bytes for empty input, got %d", len(result))
		}
	})

	t.Run("handles nil input", func(t *testing.T) {
		t.Parallel()

		result, err := Hash(crypto.SHA256, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result) != 32 {
			t.Fatalf("expected 32 bytes for nil input, got %d", len(result))
		}
	})

	t.Run("returns error for unavailable hash", func(t *testing.T) {
		t.Parallel()

		result, err := Hash(crypto.Hash(99), []byte("data"))
		if result != nil {
			t.Fatalf("expected nil result, got %v", result)
		}

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, ErrHashFunctionUnavailable) {
			t.Fatalf("expected ErrHashFunctionUnavailable, got %v", err)
		}
	})
}
