package kdfs

import (
	"crypto"
	"errors"
	"testing"
)

func TestRegister(t *testing.T) {
	t.Run("registers a new method", func(t *testing.T) {
		custom := NewMethod("custom-kdf", crypto.SHA256)

		Register(*custom)

		got, err := Get("custom-kdf")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "custom-kdf" {
			t.Fatalf("expected 'custom-kdf', got %q", got.Name())
		}
	})

	t.Run("overwrites existing method", func(t *testing.T) {
		m1 := NewMethod("overwrite-kdf", crypto.SHA256)
		m2 := NewMethod("overwrite-kdf", crypto.SHA512)

		Register(*m1)
		Register(*m2)

		got, err := Get("overwrite-kdf")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.kind != crypto.SHA512 {
			t.Fatalf("expected SHA512 after overwrite, got %v", got.kind)
		}
	})
}

func TestGet(t *testing.T) {
	t.Run("retrieves HKDF_with_SHA256", func(t *testing.T) {
		got, err := Get("HKDF_with_SHA256")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "HKDF_with_SHA256" {
			t.Fatalf("expected 'HKDF_with_SHA256', got %q", got.Name())
		}
	})

	t.Run("retrieves HKDF_with_SHA384", func(t *testing.T) {
		got, err := Get("HKDF_with_SHA384")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "HKDF_with_SHA384" {
			t.Fatalf("expected 'HKDF_with_SHA384', got %q", got.Name())
		}
	})

	t.Run("retrieves HKDF_with_SHA512", func(t *testing.T) {
		got, err := Get("HKDF_with_SHA512")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "HKDF_with_SHA512" {
			t.Fatalf("expected 'HKDF_with_SHA512', got %q", got.Name())
		}
	})

	t.Run("retrieves PBKDF2_with_SHA256", func(t *testing.T) {
		got, err := Get("PBKDF2_with_SHA256")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "PBKDF2_with_SHA256" {
			t.Fatalf("expected 'PBKDF2_with_SHA256', got %q", got.Name())
		}
	})

	t.Run("retrieves PBKDF2_with_SHA512", func(t *testing.T) {
		got, err := Get("PBKDF2_with_SHA512")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "PBKDF2_with_SHA512" {
			t.Fatalf("expected 'PBKDF2_with_SHA512', got %q", got.Name())
		}
	})

	t.Run("retrieves Scrypt_KDF", func(t *testing.T) {
		got, err := Get("Scrypt_KDF")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "Scrypt_KDF" {
			t.Fatalf("expected 'Scrypt_KDF', got %q", got.Name())
		}
	})

	t.Run("returns error for unknown method", func(t *testing.T) {
		_, err := Get("UNKNOWN_KDF")
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

		if len(list) < 6 {
			t.Fatalf("expected at least 6 predefined methods, got %d", len(list))
		}

		names := make(map[string]bool, len(list))
		for _, m := range list {
			names[m.Name()] = true
		}

		want := []string{
			"HKDF_with_SHA256",
			"HKDF_with_SHA384",
			"HKDF_with_SHA512",
			"PBKDF2_with_SHA256",
			"PBKDF2_with_SHA512",
			"Scrypt_KDF",
		}

		for _, w := range want {
			if !names[w] {
				t.Fatalf("expected Supported() to include %q", w)
			}
		}
	})
}
