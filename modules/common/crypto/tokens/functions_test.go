package tokens

import (
	"errors"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

func TestGenerate(t *testing.T) {
	t.Parallel()

	t.Run("returns error for empty subject", func(t *testing.T) {
		t.Parallel()

		key := []byte("test-secret-key-for-testing-1234567890")
		m := NewMethod("test", jwt.SigningMethodHS256, WithKey(key))

		_, err := m.Generate("", Payload{"role": "admin"})
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrSubjectEmpty) {
			t.Fatalf("expected ErrSubjectEmpty, got %v", err)
		}
	})

	t.Run("returns error for nil payload", func(t *testing.T) {
		t.Parallel()

		key := []byte("test-secret-key-for-testing-1234567890")
		m := NewMethod("test", jwt.SigningMethodHS256, WithKey(key))

		_, err := m.Generate("user@test.com", nil)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrPayloadNil) {
			t.Fatalf("expected ErrPayloadNil, got %v", err)
		}
	})

	t.Run("generates valid token", func(t *testing.T) {
		t.Parallel()

		key := []byte("test-secret-key-for-testing-1234567890")
		m := NewMethod("test", jwt.SigningMethodHS256, WithKey(key))

		token, err := m.Generate("user@test.com", Payload{"role": "admin"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if token == "" {
			t.Fatal("expected non-empty token")
		}
	})

	t.Run("generates token with issuer", func(t *testing.T) {
		t.Parallel()

		key := []byte("test-secret-key-for-testing-1234567890")
		m := NewMethod("test", jwt.SigningMethodHS256, WithKey(key), WithIssuer("my-app"))

		token, err := m.Generate("user@test.com", Payload{"role": "admin"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		payload, err := m.Validate(token)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if payload["role"] != "admin" {
			t.Fatalf("expected role 'admin', got %v", payload["role"])
		}
	})

	t.Run("returns error for nil signing key", func(t *testing.T) {
		t.Parallel()

		m := &Method{
			name:          "nil-key",
			signingMethod: jwt.SigningMethodHS256,
			signingKey:    nil,
			generateFn:    generate,
		}

		_, err := m.Generate("user@test.com", Payload{"role": "admin"})
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrSigningKeyNil) {
			t.Fatalf("expected ErrSigningKeyNil, got %v", err)
		}
	})

	t.Run("returns error for nil signing method", func(t *testing.T) {
		t.Parallel()

		m := &Method{
			name:          "nil-method",
			signingMethod: nil,
			signingKey:    []byte("key"),
			generateFn:    generate,
		}

		_, err := m.Generate("user@test.com", Payload{"role": "admin"})
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrSigningMethodNil) {
			t.Fatalf("expected ErrSigningMethodNil, got %v", err)
		}
	})
}

func TestValidate(t *testing.T) {
	t.Parallel()

	t.Run("returns error for empty token", func(t *testing.T) {
		t.Parallel()

		key := []byte("test-secret-key-for-testing-1234567890")
		m := NewMethod("test", jwt.SigningMethodHS256, WithKey(key))

		_, err := m.Validate("")
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrTokenEmpty) {
			t.Fatalf("expected ErrTokenEmpty, got %v", err)
		}
	})

	t.Run("returns error for invalid token", func(t *testing.T) {
		t.Parallel()

		key := []byte("test-secret-key-for-testing-1234567890")
		m := NewMethod("test", jwt.SigningMethodHS256, WithKey(key))

		_, err := m.Validate("invalid.token.string")
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrTokenParseFailed) {
			t.Fatalf("expected ErrTokenParseFailed, got %v", err)
		}
	})

	t.Run("validates generated token roundtrip", func(t *testing.T) {
		t.Parallel()

		key := []byte("test-secret-key-for-testing-1234567890")
		m := NewMethod("test", jwt.SigningMethodHS256, WithKey(key))

		token, err := m.Generate("user@test.com", Payload{"role": "admin", "id": float64(42)})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		payload, err := m.Validate(token)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if payload["role"] != "admin" {
			t.Fatalf("expected role 'admin', got %v", payload["role"])
		}

		if payload["id"] != float64(42) {
			t.Fatalf("expected id 42, got %v", payload["id"])
		}
	})

	t.Run("returns error for token signed with different key", func(t *testing.T) {
		t.Parallel()

		key1 := []byte("key-one-1234567890123456789012345678")
		key2 := []byte("key-two-1234567890123456789012345678")

		m1 := NewMethod("gen", jwt.SigningMethodHS256, WithKey(key1))
		m2 := NewMethod("val", jwt.SigningMethodHS256, WithKey(key2))

		token, err := m1.Generate("user@test.com", Payload{"role": "admin"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err = m2.Validate(token)
		if err == nil {
			t.Fatal("expected error for wrong key")
		}
	})

	t.Run("returns error for expired token", func(t *testing.T) {
		t.Parallel()

		key := []byte("test-secret-key-for-testing-1234567890")

		claims := &Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   "user@test.com",
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
				NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			},
			Payload: Payload{"role": "admin"},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := token.SignedString(key)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := NewMethod("test", jwt.SigningMethodHS256, WithKey(key))

		_, err = m.Validate(signed)
		if err == nil {
			t.Fatal("expected error for expired token")
		}
	})

	t.Run("returns error for token with nil payload", func(t *testing.T) {
		t.Parallel()

		key := []byte("test-secret-key-for-testing-1234567890")

		claims := &Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   "user@test.com",
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				NotBefore: jwt.NewNumericDate(time.Now()),
			},
			Payload: nil,
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := token.SignedString(key)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := NewMethod("test", jwt.SigningMethodHS256, WithKey(key))

		_, err = m.Validate(signed)
		if err == nil {
			t.Fatal("expected error for nil payload")
		}

		if !errors.Is(err, ErrTokenPayloadEmpty) {
			t.Fatalf("expected ErrTokenPayloadEmpty, got %v", err)
		}
	})

	t.Run("wraps generate error", func(t *testing.T) {
		t.Parallel()

		fail := func(_ *Method, _ string, _ Payload) (string, error) {
			return "", errors.New("gen boom")
		}

		m := NewMethod("fail", jwt.SigningMethodHS256, WithGenerateFn(fail))

		_, err := m.Generate("user", Payload{"x": "y"})
		if err == nil {
			t.Fatal("expected error")
		}

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}
	})

	t.Run("wraps validate error", func(t *testing.T) {
		t.Parallel()

		fail := func(_ *Method, _ string) (Payload, error) {
			return nil, errors.New("val boom")
		}

		m := NewMethod("fail", jwt.SigningMethodHS256, WithValidateFn(fail))

		_, err := m.Validate("some-token")
		if err == nil {
			t.Fatal("expected error")
		}

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}
	})
}
