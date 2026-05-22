package hashes

import (
	"bytes"
	"crypto"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestMethod_NewHasher(t *testing.T) {
	t.Parallel()

	t.Run("returns hash.Hash for available algorithm", func(t *testing.T) {
		t.Parallel()

		h, err := SHA256.NewHasher()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if h == nil {
			t.Fatal("expected non-nil hash.Hash")
		}

		if h.Size() != 32 {
			t.Fatalf("expected hash size 32, got %d", h.Size())
		}
	})

	t.Run("round-trip matches one-shot Hash for in-memory buffer", func(t *testing.T) {
		t.Parallel()

		input := []byte("the quick brown fox jumps over the lazy dog")

		expected, err := SHA256.Hash(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		h, err := SHA256.NewHasher()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		src := bytes.NewReader(input)

		_, copyErr := io.Copy(h, src)
		if copyErr != nil {
			t.Fatalf("unexpected copy error: %v", copyErr)
		}

		got := h.Sum(nil)
		if !bytes.Equal(got, expected) {
			t.Fatalf("streaming digest mismatch: got %x, want %x", got, expected)
		}
	})

	t.Run("round-trip matches one-shot Hash for temp file", func(t *testing.T) {
		t.Parallel()

		input := bytes.Repeat([]byte("yarumo streaming hash test "), 4096)

		expected, err := SHA512.Hash(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		dir := t.TempDir()
		path := filepath.Join(dir, "payload.bin")

		writeErr := os.WriteFile(path, input, 0o600)
		if writeErr != nil {
			t.Fatalf("unexpected write error: %v", writeErr)
		}

		f, openErr := os.Open(path) //nolint:gosec // path is a t.TempDir test artifact, not user-controlled
		if openErr != nil {
			t.Fatalf("unexpected open error: %v", openErr)
		}
		defer func() { _ = f.Close() }()

		h, err := SHA512.NewHasher()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, copyErr := io.Copy(h, f)
		if copyErr != nil {
			t.Fatalf("unexpected copy error: %v", copyErr)
		}

		got := h.Sum(nil)
		if !bytes.Equal(got, expected) {
			t.Fatalf("file streaming digest mismatch: got %x, want %x", got, expected)
		}
	})

	t.Run("supports multiple writes (incremental)", func(t *testing.T) {
		t.Parallel()

		parts := [][]byte{[]byte("hello "), []byte("streaming "), []byte("hash")}

		joined := bytes.Join(parts, nil)

		expected, err := SHA3_256.Hash(joined)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		h, err := SHA3_256.NewHasher()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for _, p := range parts {
			_, writeErr := h.Write(p)
			if writeErr != nil {
				t.Fatalf("unexpected write error: %v", writeErr)
			}
		}

		got := h.Sum(nil)
		if !bytes.Equal(got, expected) {
			t.Fatalf("incremental digest mismatch: got %x, want %x", got, expected)
		}
	})

	t.Run("returns error for unavailable hash kind", func(t *testing.T) {
		t.Parallel()

		// crypto.Hash(99) is not a registered hash function.
		m := NewMethod("unavailable-stream", crypto.Hash(99))

		h, err := m.NewHasher()
		if h != nil {
			t.Fatalf("expected nil hash.Hash, got %v", h)
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
