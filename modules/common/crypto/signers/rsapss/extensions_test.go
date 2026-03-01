package rsapss

import (
	"crypto"
	"crypto/rsa"
	"errors"
	"testing"
)

func TestRegister(t *testing.T) {
	t.Run("registers a new method", func(t *testing.T) {
		custom := NewMethod("custom-rsapss", crypto.SHA256, rsa.PSSSaltLengthEqualsHash, []int{2048})

		Register(*custom)

		got, err := Get("custom-rsapss")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "custom-rsapss" {
			t.Fatalf("expected 'custom-rsapss', got %q", got.Name())
		}
	})
}

func TestGet(t *testing.T) {
	t.Run("retrieves predefined method", func(t *testing.T) {
		got, err := Get("RSASSA_PSS_using_SHA256")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "RSASSA_PSS_using_SHA256" {
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
