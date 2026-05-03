package aead

import (
	"errors"
	"testing"
)

func TestRegister(t *testing.T) {
	t.Run("registers a new method", func(t *testing.T) {
		custom := NewMethod("custom-aead", 32, 12, aesgcm)

		Register(*custom)

		got, err := Get("custom-aead")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "custom-aead" {
			t.Fatalf("expected 'custom-aead', got %q", got.Name())
		}
	})
}

func TestGet(t *testing.T) {
	t.Run("retrieves predefined method", func(t *testing.T) {
		got, err := Get("AES_256_GCM")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "AES_256_GCM" {
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

		if len(list) < 4 {
			t.Fatalf("expected at least 4, got %d", len(list))
		}
	})
}
