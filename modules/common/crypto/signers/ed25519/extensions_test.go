package ed25519

import (
	"errors"
	"testing"
)

func TestRegister(t *testing.T) {
	t.Run("registers a new method", func(t *testing.T) {
		custom := NewMethod("custom-ed25519")

		Register(*custom)

		got, err := Get("custom-ed25519")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "custom-ed25519" {
			t.Fatalf("expected 'custom-ed25519', got %q", got.Name())
		}
	})
}

func TestGet(t *testing.T) {
	t.Run("retrieves predefined method", func(t *testing.T) {
		got, err := Get("Ed25519")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "Ed25519" {
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
	t.Run("returns at least the predefined method", func(t *testing.T) {
		list := Supported()

		if len(list) < 1 {
			t.Fatalf("expected at least 1, got %d", len(list))
		}
	})
}
