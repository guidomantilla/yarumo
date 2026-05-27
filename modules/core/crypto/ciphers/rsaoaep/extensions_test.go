package rsaoaep

import (
	"crypto"
	"errors"
	"testing"
)

func TestRegister(t *testing.T) {
	t.Run("registers a new method", func(t *testing.T) {
		custom := NewMethod("custom-rsaoaep", crypto.SHA256, []int{2048})

		Register(*custom)

		got, err := Get("custom-rsaoaep")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "custom-rsaoaep" {
			t.Fatalf("expected 'custom-rsaoaep', got %q", got.Name())
		}
	})
}

func TestGet(t *testing.T) {
	t.Run("retrieves predefined method", func(t *testing.T) {
		got, err := Get("RSA-OAEP-SHA256")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "RSA-OAEP-SHA256" {
			t.Fatalf("unexpected name: %q", got.Name())
		}
	})

	t.Run("retrieves RSA-OAEP-SHA384", func(t *testing.T) {
		got, err := Get("RSA-OAEP-SHA384")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "RSA-OAEP-SHA384" {
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

		if len(list) < 3 {
			t.Fatalf("expected at least 3, got %d", len(list))
		}

		found := make(map[string]bool, len(list))
		for _, m := range list {
			found[m.Name()] = true
		}

		for _, name := range []string{"RSA-OAEP-SHA256", "RSA-OAEP-SHA384", "RSA-OAEP-SHA512"} {
			if !found[name] {
				t.Fatalf("expected %q in Supported(), missing", name)
			}
		}
	})
}
