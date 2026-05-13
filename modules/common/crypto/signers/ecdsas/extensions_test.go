package ecdsas

import (
	"crypto"
	"crypto/elliptic"
	"errors"
	"testing"
)

func TestRegister(t *testing.T) {
	t.Parallel()

	t.Run("registers a new method", func(t *testing.T) {
		t.Parallel()
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
	t.Parallel()

	t.Run("retrieves predefined method", func(t *testing.T) {
		t.Parallel()

		got, err := Get("ECDSA_with_SHA256_over_P256")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "ECDSA_with_SHA256_over_P256" {
			t.Fatalf("unexpected name: %q", got.Name())
		}
	})

	t.Run("retrieves predefined P-384 method", func(t *testing.T) {
		t.Parallel()

		got, err := Get("ECDSA_with_SHA384_over_P384")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "ECDSA_with_SHA384_over_P384" {
			t.Fatalf("unexpected name: %q", got.Name())
		}
	})

	t.Run("returns error for unknown method", func(t *testing.T) {
		t.Parallel()

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
	t.Parallel()

	t.Run("returns at least the predefined methods", func(t *testing.T) {
		t.Parallel()

		list := Supported()

		if len(list) < 3 {
			t.Fatalf("expected at least 3, got %d", len(list))
		}

		want := map[string]bool{
			"ECDSA_with_SHA256_over_P256": false,
			"ECDSA_with_SHA384_over_P384": false,
			"ECDSA_with_SHA512_over_P521": false,
		}

		for _, m := range list {
			_, ok := want[m.Name()]
			if ok {
				want[m.Name()] = true
			}
		}

		for name, seen := range want {
			if !seen {
				t.Fatalf("expected Supported() to include %q", name)
			}
		}
	})
}
