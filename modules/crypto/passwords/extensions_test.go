package passwords

import (
	"errors"
	"strings"
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

	t.Run("retrieves predefined Argon2id method", func(t *testing.T) {

		got, err := Get("Argon2id")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != Argon2id.Name() {
			t.Fatalf("expected %q, got %q", Argon2id.Name(), got.Name())
		}
	})

	t.Run("retrieves predefined Argon2i method", func(t *testing.T) {

		got, err := Get("Argon2i")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != Argon2i.Name() {
			t.Fatalf("expected %q, got %q", Argon2i.Name(), got.Name())
		}
	})

	t.Run("legacy Argon2 name is not registered", func(t *testing.T) {

		// Per the YA-0030 migration: the deprecated Go-level alias
		// passwords.Argon2 is NOT separately registered. Callers using
		// Get("Argon2") must migrate to Get("Argon2id").
		_, err := Get("Argon2")
		if err == nil {
			t.Fatal("expected ErrAlgorithmNotSupported for legacy Argon2 name")
		}

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
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

		if len(list) < 5 {
			t.Fatalf("expected at least 5 methods, got %d", len(list))
		}
	})

	t.Run("contains Argon2id method", func(t *testing.T) {

		list := Supported()

		found := false
		for _, m := range list {
			if m.name == textCodecTestName {
				found = true
				break
			}
		}

		if !found {
			t.Fatalf("expected %s in supported list", textCodecTestName)
		}
	})

	t.Run("contains Argon2i method", func(t *testing.T) {

		list := Supported()

		found := false
		for _, m := range list {
			if m.name == "Argon2i" {
				found = true
				break
			}
		}

		if !found {
			t.Fatal("expected Argon2i in supported list")
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

	t.Run("routes new {argon2id} prefix to Argon2id", func(t *testing.T) {

		encoded, err := Argon2id.Encode("argon2id-prefix")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.HasPrefix(encoded, Argon2idPrefixKey) {
			t.Fatalf("expected new encode to carry %q, got %q", Argon2idPrefixKey, encoded)
		}

		got, err := ByPrefix(encoded)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Name() != textCodecTestName {
			t.Fatalf("expected %s, got %q", textCodecTestName, got.Name())
		}
	})

	t.Run("routes {argon2i} prefix to Argon2i", func(t *testing.T) {

		encoded, err := Argon2i.Encode("argon2i-prefix")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.HasPrefix(encoded, Argon2iPrefixKey) {
			t.Fatalf("expected encode to carry %q, got %q", Argon2iPrefixKey, encoded)
		}

		got, err := ByPrefix(encoded)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Name() != "Argon2i" {
			t.Fatalf("expected Argon2i, got %q", got.Name())
		}
	})

	t.Run("legacy {argon2} prefix dual-matches to Argon2id", func(t *testing.T) {

		// Simulate a hash produced by the pre-YA-0030 code by constructing a
		// Method that emits the legacy {argon2} prefix while still using the
		// underlying argon2id KDF (which is exactly what the old code did).
		legacy := NewMethod("LegacyArgon2", Argon2PrefixKey, WithArgon2Params(Argon2Iterations, Argon2Memory, Argon2Threads, Argon2SaltLength, Argon2KeyLength))
		encoded, err := legacy.Encode("legacy-argon2")
		if err != nil {
			t.Fatalf("unexpected encode error: %v", err)
		}
		if !strings.HasPrefix(encoded, Argon2PrefixKey) {
			t.Fatalf("expected legacy encode to carry %q, got %q", Argon2PrefixKey, encoded)
		}

		got, err := ByPrefix(encoded)
		if err != nil {
			t.Fatalf("unexpected ByPrefix error: %v", err)
		}
		if got.Name() != textCodecTestName {
			t.Fatalf("expected legacy {argon2} to route to %s, got %q", textCodecTestName, got.Name())
		}

		// And the resolved method MUST verify the legacy hash.
		ok, err := got.Verify(encoded, "legacy-argon2")
		if err != nil {
			t.Fatalf("unexpected verify error: %v", err)
		}
		if !ok {
			t.Fatal("expected legacy {argon2} hash to verify under Argon2id")
		}
	})
}
