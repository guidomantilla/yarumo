package hmacs

import (
	"crypto"
	"errors"
	"testing"
)

func TestRegister(t *testing.T) {
	t.Parallel()

	t.Run("registers a new method", func(t *testing.T) {
		t.Parallel()

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
	t.Parallel()

	t.Run("retrieves predefined HMAC_with_SHA256", func(t *testing.T) {
		t.Parallel()

		got, err := Get("HMAC_with_SHA256")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "HMAC_with_SHA256" {
			t.Fatalf("expected 'HMAC_with_SHA256', got %q", got.Name())
		}
	})

	t.Run("retrieves predefined HMAC_with_SHA384", func(t *testing.T) {
		t.Parallel()

		got, err := Get("HMAC_with_SHA384")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Name() != "HMAC_with_SHA384" {
			t.Fatalf("expected 'HMAC_with_SHA384', got %q", got.Name())
		}
	})

	t.Run("returns error for unknown method", func(t *testing.T) {
		t.Parallel()

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
	t.Parallel()

	t.Run("returns at least the predefined methods", func(t *testing.T) {
		t.Parallel()

		list := Supported()

		if len(list) < 3 {
			t.Fatalf("expected at least 3 predefined methods, got %d", len(list))
		}

		names := make(map[string]bool, len(list))
		for _, m := range list {
			names[m.Name()] = true
		}

		for _, want := range []string{"HMAC_with_SHA256", "HMAC_with_SHA384", "HMAC_with_SHA512"} {
			if !names[want] {
				t.Fatalf("expected Supported() to include %q", want)
			}
		}
	})
}

func TestHMAC_with_SHA384_RoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("generate key, digest, and validate", func(t *testing.T) {
		t.Parallel()

		method, err := Get("HMAC_with_SHA384")
		if err != nil {
			t.Fatalf("unexpected error retrieving method: %v", err)
		}

		key, err := method.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error generating key: %v", err)
		}

		if len(key) != 48 {
			t.Fatalf("expected 48-byte key, got %d", len(key))
		}

		digest, err := method.Digest(key, []byte("hello-sha384"))
		if err != nil {
			t.Fatalf("unexpected error computing digest: %v", err)
		}

		if len(digest) != 48 {
			t.Fatalf("expected 48-byte digest, got %d", len(digest))
		}

		ok, err := method.Validate(key, digest, []byte("hello-sha384"))
		if err != nil {
			t.Fatalf("unexpected error validating digest: %v", err)
		}

		if !ok {
			t.Fatal("expected validation to succeed")
		}
	})
}
