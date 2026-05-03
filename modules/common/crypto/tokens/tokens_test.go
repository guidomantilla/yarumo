package tokens

import (
	"testing"

	jwt "github.com/golang-jwt/jwt/v5"
)

func TestNewMethod(t *testing.T) {
	t.Parallel()

	t.Run("creates method with default options", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("test", jwt.SigningMethodHS256)

		if m == nil {
			t.Fatal("expected non-nil method")
		}
		if m.name != "test" {
			t.Fatalf("expected name 'test', got %q", m.name)
		}
		if m.signingMethod != jwt.SigningMethodHS256 {
			t.Fatalf("expected HS256, got %v", m.signingMethod)
		}
	})

	t.Run("creates method with custom key", func(t *testing.T) {
		t.Parallel()

		key := []byte("custom-key-12345")
		m := NewMethod("custom", jwt.SigningMethodHS512, WithKey(key))

		if string(m.signingKey) != string(key) {
			t.Fatal("expected custom signing key")
		}
	})

	t.Run("creates method with custom issuer and timeout", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("issuer-test", jwt.SigningMethodHS256, WithIssuer("my-app"))

		if m.issuer != "my-app" {
			t.Fatalf("expected issuer 'my-app', got %q", m.issuer)
		}
	})
}

func TestMethod_Name(t *testing.T) {
	t.Parallel()

	t.Run("returns method name for HS256", func(t *testing.T) {
		t.Parallel()

		if JWT_HS256.Name() != JWT_HS256.name {
			t.Fatalf("expected %q, got %q", JWT_HS256.name, JWT_HS256.Name())
		}
	})

	t.Run("returns method name for HS384", func(t *testing.T) {
		t.Parallel()

		if JWT_HS384.Name() != JWT_HS384.name {
			t.Fatalf("expected %q, got %q", JWT_HS384.name, JWT_HS384.Name())
		}
	})

	t.Run("returns method name for HS512", func(t *testing.T) {
		t.Parallel()

		if JWT_HS512.Name() != JWT_HS512.name {
			t.Fatalf("expected %q, got %q", JWT_HS512.name, JWT_HS512.Name())
		}
	})
}

func TestMethod_Generate_Validate_Roundtrip(t *testing.T) {
	t.Parallel()

	t.Run("HS256 roundtrip", func(t *testing.T) {
		t.Parallel()

		key := []byte("hs256-secret-key-for-testing-1234567890")
		m := NewMethod("hs256", jwt.SigningMethodHS256, WithKey(key))

		expected := "hs256-val"
		token, err := m.Generate("subject", Payload{"key": expected})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		payload, err := m.Validate(token)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if payload["key"] != expected {
			t.Fatalf("expected key %q, got %v", expected, payload["key"])
		}
	})

	t.Run("HS384 roundtrip", func(t *testing.T) {
		t.Parallel()

		key := []byte("hs384-secret-key-for-testing-123456789012345678901234")
		m := NewMethod("hs384", jwt.SigningMethodHS384, WithKey(key))

		expected := "hs384-val"
		token, err := m.Generate("subject", Payload{"key": expected})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		payload, err := m.Validate(token)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if payload["key"] != expected {
			t.Fatalf("expected key %q, got %v", expected, payload["key"])
		}
	})

	t.Run("HS512 roundtrip", func(t *testing.T) {
		t.Parallel()

		key := []byte("hs512-secret-key-for-testing-12345678901234567890123456789012345678901234")
		m := NewMethod("hs512", jwt.SigningMethodHS512, WithKey(key))

		expected := "hs512-val"
		token, err := m.Generate("subject", Payload{"key": expected})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		payload, err := m.Validate(token)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if payload["key"] != expected {
			t.Fatalf("expected key %q, got %v", expected, payload["key"])
		}
	})
}
