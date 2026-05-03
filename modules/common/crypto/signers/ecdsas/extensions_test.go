package ecdsas

import (
	"crypto"
	"crypto/elliptic"
	"errors"
	"testing"
)

func TestRegister(t *testing.T) {
	t.Run("registers a new method", func(t *testing.T) {
		custom := NewMethod("custom-ecdsa", crypto.SHA256, 32, elliptic.P256())

		Register(*custom)

		got, err := Get("custom-ecdsa")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "custom-ecdsa" {
			t.Fatalf("expected 'custom-ecdsa', got %q", got.Name())
		}
	})
}

func TestGet(t *testing.T) {
	t.Run("retrieves predefined method", func(t *testing.T) {
		got, err := Get("ECDSA_with_SHA256_over_P256")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "ECDSA_with_SHA256_over_P256" {
			t.Fatalf("unexpected name: %q", got.Name())
		}
	})

	t.Run("returns error for unknown method", func(t *testing.T) {
		_, err := Get("UNKNOWN")
		if err == nil {
			t.Fatal("expected error")
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

		if len(list) < 2 {
			t.Fatalf("expected at least 2, got %d", len(list))
		}
	})
}
