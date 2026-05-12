package rsassas

import (
	"crypto"
	"errors"
	"testing"
)

func TestRegister(t *testing.T) {
	t.Run("registers a new method", func(t *testing.T) {
		custom := NewMethod("custom-rsassas", crypto.SHA256, PSS, []int{2048})

		Register(*custom)

		got, err := Get("custom-rsassas")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "custom-rsassas" {
			t.Fatalf("expected 'custom-rsassas', got %q", got.Name())
		}
	})
}

func TestGet(t *testing.T) {
	t.Run("retrieves predefined PSS method", func(t *testing.T) {
		got, err := Get("RSASSA_PSS_using_SHA256")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "RSASSA_PSS_using_SHA256" {
			t.Fatalf("unexpected name: %q", got.Name())
		}
	})

	t.Run("retrieves predefined PKCS1v15 method", func(t *testing.T) {
		got, err := Get("RSASSA_PKCS1v15_using_SHA256")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "RSASSA_PKCS1v15_using_SHA256" {
			t.Fatalf("unexpected name: %q", got.Name())
		}
	})

	t.Run("retrieves predefined PKCS1v15 SHA384 method", func(t *testing.T) {
		got, err := Get("RSASSA_PKCS1v15_using_SHA384")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "RSASSA_PKCS1v15_using_SHA384" {
			t.Fatalf("unexpected name: %q", got.Name())
		}
	})

	t.Run("retrieves predefined PSS SHA384 method", func(t *testing.T) {
		got, err := Get("RSASSA_PSS_using_SHA384")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "RSASSA_PSS_using_SHA384" {
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

		if len(list) < 6 {
			t.Fatalf("expected at least 6, got %d", len(list))
		}
	})

	t.Run("includes RSASSA_PSS_using_SHA384", func(t *testing.T) {
		list := Supported()

		var found bool
		for _, m := range list {
			if m.Name() == "RSASSA_PSS_using_SHA384" {
				found = true
				break
			}
		}

		if !found {
			t.Fatal("expected Supported() to include RSASSA_PSS_using_SHA384")
		}
	})
}
