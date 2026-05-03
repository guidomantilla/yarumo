package passwords

import (
	"errors"
	"testing"
)

func TestRegister(t *testing.T) {

	t.Run("registers a new method", func(t *testing.T) {

		custom := NewMethod("Custom", "{custom}", WithBcryptParams(BcryptDefaultCost))
		Register(*custom)

		got, err := Get("Custom")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "Custom" {
			t.Fatalf("expected 'Custom', got %q", got.Name())
		}
	})

	t.Run("overwrites existing method", func(t *testing.T) {

		m1 := NewMethod("Override", "{override}", WithBcryptParams(BcryptDefaultCost))
		Register(*m1)

		m2 := NewMethod("Override", "{override-v2}", WithBcryptParams(BcryptDefaultCost+1))
		Register(*m2)

		got, err := Get("Override")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.prefix != "{override-v2}" {
			t.Fatalf("expected prefix '{override-v2}', got %q", got.prefix)
		}
	})
}

func TestGet(t *testing.T) {

	t.Run("retrieves predefined Argon2 method", func(t *testing.T) {

		got, err := Get("Argon2")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != Argon2.Name() {
			t.Fatalf("expected %q, got %q", Argon2.Name(), got.Name())
		}
	})

	t.Run("retrieves predefined Bcrypt method", func(t *testing.T) {

		got, err := Get("Bcrypt")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != Bcrypt.Name() {
			t.Fatalf("expected %q, got %q", Bcrypt.Name(), got.Name())
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
			t.Fatalf("expected at least 4 methods, got %d", len(list))
		}
	})

	t.Run("contains Argon2 method", func(t *testing.T) {

		list := Supported()

		found := false
		for _, m := range list {
			if m.name == "Argon2" {
				found = true
				break
			}
		}

		if !found {
			t.Fatal("expected Argon2 in supported list")
		}
	})
}

func TestByPrefix(t *testing.T) {

	t.Run("returns method matching prefix", func(t *testing.T) {

		encoded, err := Bcrypt.Encode("test-password")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, err := ByPrefix(encoded)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != Bcrypt.Name() {
			t.Fatalf("expected %q, got %q", Bcrypt.Name(), got.Name())
		}
	})

	t.Run("returns error for unknown prefix", func(t *testing.T) {

		_, err := ByPrefix("{unknown}$data")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("returns error for empty string", func(t *testing.T) {

		_, err := ByPrefix("")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}
