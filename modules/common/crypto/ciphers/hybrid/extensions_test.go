package hybrid

import (
	"errors"
	"testing"

	"github.com/cloudflare/circl/hpke"
)

func TestRegister(t *testing.T) {
	t.Run("registers a new method", func(t *testing.T) {
		custom := NewMethod("custom-hpke", hpke.KEM_X25519_HKDF_SHA256, hpke.KDF_HKDF_SHA256, hpke.AEAD_AES256GCM)

		Register(*custom)

		got, err := Get("custom-hpke")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "custom-hpke" {
			t.Fatalf("expected 'custom-hpke', got %q", got.Name())
		}
	})
}

func TestGet(t *testing.T) {
	t.Run("retrieves predefined method", func(t *testing.T) {
		got, err := Get("HPKE_X25519_HKDF_SHA256_AES_256_GCM")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "HPKE_X25519_HKDF_SHA256_AES_256_GCM" {
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

		if len(list) < 1 {
			t.Fatalf("expected at least 1, got %d", len(list))
		}
	})
}
