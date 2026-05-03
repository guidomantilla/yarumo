package tokens

import (
	"errors"
	"testing"

	jwt "github.com/golang-jwt/jwt/v5"
)

func TestRegister(t *testing.T) {

	t.Run("registers a new method", func(t *testing.T) {

		custom := NewMethod("Custom_HS256", jwt.SigningMethodHS256)
		Register(*custom)

		got, err := Get("Custom_HS256")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "Custom_HS256" {
			t.Fatalf("expected 'Custom_HS256', got %q", got.Name())
		}
	})

	t.Run("overwrites existing method", func(t *testing.T) {

		key1 := []byte("key-one")
		key2 := []byte("key-two")

		m1 := NewMethod("Override", jwt.SigningMethodHS256, WithKey(key1))
		Register(*m1)

		m2 := NewMethod("Override", jwt.SigningMethodHS384, WithKey(key2))
		Register(*m2)

		got, err := Get("Override")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.signingMethod != jwt.SigningMethodHS384 {
			t.Fatalf("expected HS384, got %v", got.signingMethod)
		}
	})
}

func TestGet(t *testing.T) {

	t.Run("retrieves predefined JWT_HS256 method", func(t *testing.T) {

		got, err := Get("JWT_HS256")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "JWT_HS256" {
			t.Fatalf("expected 'JWT_HS256', got %q", got.Name())
		}
	})

	t.Run("retrieves predefined JWT_HS512 method", func(t *testing.T) {

		got, err := Get("JWT_HS512")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "JWT_HS512" {
			t.Fatalf("expected 'JWT_HS512', got %q", got.Name())
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
			t.Fatalf("expected at least 3 methods, got %d", len(list))
		}
	})

	t.Run("contains JWT_HS256 method", func(t *testing.T) {

		list := Supported()

		found := false
		for _, m := range list {
			if m.name == "JWT_HS256" {
				found = true
				break
			}
		}

		if !found {
			t.Fatal("expected JWT_HS256 in supported list")
		}
	})
}
