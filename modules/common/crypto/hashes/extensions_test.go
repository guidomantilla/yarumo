package hashes

import (
	"crypto"
	"errors"
	"testing"
)

func TestRegister(t *testing.T) {
	t.Run("registers a new method", func(t *testing.T) {
		custom := NewMethod("custom-hash", crypto.SHA256)

		Register(*custom)

		got, err := Get("custom-hash")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "custom-hash" {
			t.Fatalf("expected 'custom-hash', got %q", got.Name())
		}
	})

	t.Run("overwrites existing method", func(t *testing.T) {
		m1 := NewMethod("overwrite-test", crypto.SHA256)
		m2 := NewMethod("overwrite-test", crypto.SHA512)

		Register(*m1)
		Register(*m2)

		got, err := Get("overwrite-test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.kind != crypto.SHA512 {
			t.Fatalf("expected SHA512 after overwrite, got %v", got.kind)
		}
	})
}

func TestGet(t *testing.T) {
	t.Run("retrieves predefined SHA256", func(t *testing.T) {
		got, err := Get("SHA256")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "SHA256" {
			t.Fatalf("expected 'SHA256', got %q", got.Name())
		}
	})

	t.Run("retrieves predefined BLAKE2b_512", func(t *testing.T) {
		got, err := Get("BLAKE2b_512")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "BLAKE2b_512" {
			t.Fatalf("expected 'BLAKE2b_512', got %q", got.Name())
		}
	})

	t.Run("returns error for unknown method", func(t *testing.T) {
		_, err := Get("UNKNOWN")
		if err == nil {
			t.Fatal("expected error for unknown method")
		}

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}
	})
}

func TestSupported(t *testing.T) {
	t.Run("returns at least the predefined methods", func(t *testing.T) {
		list := Supported()

		if len(list) < 6 {
			t.Fatalf("expected at least 6 predefined methods, got %d", len(list))
		}
	})

	t.Run("contains SHA256", func(t *testing.T) {
		list := Supported()

		found := false

		for _, m := range list {
			if m.name == "SHA256" {
				found = true
				break
			}
		}

		if !found {
			t.Fatal("expected SHA256 in supported list")
		}
	})
}
