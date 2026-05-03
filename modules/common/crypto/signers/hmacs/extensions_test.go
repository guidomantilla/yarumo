package hmacs

import (
	"crypto"
	"errors"
	"testing"
)

func TestRegister(t *testing.T) {
	t.Run("registers a new method", func(t *testing.T) {
		custom := NewMethod("custom-hmac", crypto.SHA256, 32)

		Register(*custom)

		got, err := Get("custom-hmac")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "custom-hmac" {
			t.Fatalf("expected 'custom-hmac', got %q", got.Name())
		}
	})
}

func TestGet(t *testing.T) {
	t.Run("retrieves predefined HMAC_with_SHA256", func(t *testing.T) {
		got, err := Get("HMAC_with_SHA256")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "HMAC_with_SHA256" {
			t.Fatalf("expected 'HMAC_with_SHA256', got %q", got.Name())
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

		if len(list) < 2 {
			t.Fatalf("expected at least 2 predefined methods, got %d", len(list))
		}
	})
}
