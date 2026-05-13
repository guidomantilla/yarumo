package hashes

import (
	"crypto"
	"errors"
	"testing"
)

func TestCompute(t *testing.T) {
	t.Parallel()

	t.Run("computes SHA256 digest by name", func(t *testing.T) {
		t.Parallel()

		got, err := Compute("SHA256", []byte("hello"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != 32 {
			t.Fatalf("expected 32 bytes, got %d", len(got))
		}
	})

	t.Run("matches Method.Hash for the same name", func(t *testing.T) {
		t.Parallel()

		viaHelper, err := Compute("SHA512", []byte("payload"))
		if err != nil {
			t.Fatalf("unexpected error from helper: %v", err)
		}

		method, err := Get("SHA512")
		if err != nil {
			t.Fatalf("unexpected error from Get: %v", err)
		}

		viaMethod, err := method.Hash([]byte("payload"))
		if err != nil {
			t.Fatalf("unexpected error from Method.Hash: %v", err)
		}

		if viaHelper.ToHex() != viaMethod.ToHex() {
			t.Fatal("expected helper and Method.Hash to agree")
		}
	})

	t.Run("returns domain error for unknown name", func(t *testing.T) {
		t.Parallel()

		got, err := Compute("UNKNOWN", []byte("data"))
		if err == nil {
			t.Fatal("expected error for unknown name")
		}

		if got != nil {
			t.Fatalf("expected nil bytes, got %v", got)
		}

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}

		if domErr.Type != HashNotFound {
			t.Fatalf("expected type %q, got %q", HashNotFound, domErr.Type)
		}
	})
}

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
