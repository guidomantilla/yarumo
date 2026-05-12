package tokens

import (
	"errors"
	"testing"

	jwt "github.com/golang-jwt/jwt/v5"
)

const roleAdmin = "admin"

func TestNewMethod(t *testing.T) {
	t.Parallel()

	t.Run("creates method with default options", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("test", AlgorithmHS256)

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
		m := NewMethod("custom", AlgorithmHS512, WithKey(key))

		signing, ok := m.signingKey.([]byte)
		if !ok {
			t.Fatalf("expected signing key []byte, got %T", m.signingKey)
		}
		if string(signing) != string(key) {
			t.Fatal("expected custom signing key")
		}
		if m.signingMethod != jwt.SigningMethodHS512 {
			t.Fatalf("expected HS512, got %v", m.signingMethod)
		}
	})

	t.Run("creates method with custom issuer and timeout", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("issuer-test", AlgorithmHS256, WithIssuer("my-app"))

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
		m := NewMethod("hs256", AlgorithmHS256, WithKey(key))

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
		m := NewMethod("hs384", AlgorithmHS384, WithKey(key))

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
		m := NewMethod("hs512", AlgorithmHS512, WithKey(key))

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

// TestYA0008_KeyManagementPaths exercises the three caller paths after the
// YA-0008 behavior change. See modules/common/crypto/tokens documentation in
// types.go for the design decision.
func TestYA0008_KeyManagementPaths(t *testing.T) {
	t.Parallel()

	t.Run("default NewOptions then Generate returns ErrSigningKeyNil", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("no-key", AlgorithmHS256)

		_, err := m.Generate("user@test.com", Payload{"role": roleAdmin})
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, ErrSigningKeyNil) {
			t.Fatalf("expected ErrSigningKeyNil, got %v", err)
		}
	})

	t.Run("default NewOptions then Validate returns ErrVerifyingKeyNil", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("no-key", AlgorithmHS256)

		_, err := m.Validate("some.jwt.token")
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, ErrVerifyingKeyNil) {
			t.Fatalf("expected ErrVerifyingKeyNil, got %v", err)
		}
	})

	t.Run("WithGeneratedKey enables Generate/Validate roundtrip", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("gen-key", AlgorithmHS256, WithGeneratedKey())

		token, err := m.Generate("user@test.com", Payload{"role": roleAdmin})
		if err != nil {
			t.Fatalf("unexpected generate error: %v", err)
		}
		if token == "" {
			t.Fatal("expected non-empty token")
		}

		payload, err := m.Validate(token)
		if err != nil {
			t.Fatalf("unexpected validate error: %v", err)
		}
		if payload["role"] != roleAdmin {
			t.Fatalf("expected role 'admin', got %v", payload["role"])
		}
	})

	t.Run("WithSigningKey and WithVerifyingKey enable roundtrip", func(t *testing.T) {
		t.Parallel()

		key := []byte("explicit-key-1234567890-abcdefghij")
		m := NewMethod("split-keys", AlgorithmHS256,
			WithSigningKey(key),
			WithVerifyingKey(key),
		)

		token, err := m.Generate("user@test.com", Payload{"scope": "read"})
		if err != nil {
			t.Fatalf("unexpected generate error: %v", err)
		}

		payload, err := m.Validate(token)
		if err != nil {
			t.Fatalf("unexpected validate error: %v", err)
		}
		if payload["scope"] != "read" {
			t.Fatalf("expected scope 'read', got %v", payload["scope"])
		}
	})
}

// TestYA0009_AlgorithmEnum verifies the Algorithm enum maps to the expected
// jwt.SigningMethod internally and that unknown values surface as
// ErrAlgorithmInvalid / ErrAlgorithmUnknown via the package's assertion path.
func TestYA0009_AlgorithmEnum(t *testing.T) {
	t.Parallel()

	t.Run("HS256 maps to jwt.SigningMethodHS256", func(t *testing.T) {
		t.Parallel()

		if got := signingMethodFor(AlgorithmHS256); got != jwt.SigningMethodHS256 {
			t.Fatalf("expected jwt.SigningMethodHS256, got %v", got)
		}
	})

	t.Run("HS384 maps to jwt.SigningMethodHS384", func(t *testing.T) {
		t.Parallel()

		if got := signingMethodFor(AlgorithmHS384); got != jwt.SigningMethodHS384 {
			t.Fatalf("expected jwt.SigningMethodHS384, got %v", got)
		}
	})

	t.Run("HS512 maps to jwt.SigningMethodHS512", func(t *testing.T) {
		t.Parallel()

		if got := signingMethodFor(AlgorithmHS512); got != jwt.SigningMethodHS512 {
			t.Fatalf("expected jwt.SigningMethodHS512, got %v", got)
		}
	})

	t.Run("unknown Algorithm returns nil from signingMethodFor", func(t *testing.T) {
		t.Parallel()

		if got := signingMethodFor(Algorithm("XX999")); got != nil {
			t.Fatalf("expected nil for unsupported algorithm, got %v", got)
		}
	})

	t.Run("ErrAlgorithmInvalid wraps ErrAlgorithmUnknown for unknown values", func(t *testing.T) {
		t.Parallel()

		err := ErrAlgorithmInvalid(Algorithm("XX999"))
		if err == nil {
			t.Fatal("expected non-nil error")
		}
		if !errors.Is(err, ErrAlgorithmUnknown) {
			t.Fatalf("expected ErrAlgorithmUnknown in chain, got %v", err)
		}

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}
	})
}
