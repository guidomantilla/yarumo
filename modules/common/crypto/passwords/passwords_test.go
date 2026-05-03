package passwords

import (
	"errors"
	"strings"
	"testing"
)

func TestNewMethod(t *testing.T) {
	t.Parallel()

	t.Run("creates method with default options", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("test", "{test}", WithBcryptParams(BcryptDefaultCost))

		if m == nil {
			t.Fatal("expected non-nil method")
		}
		if m.name != "test" {
			t.Fatalf("expected name 'test', got %q", m.name)
		}
		if m.prefix != "{test}" {
			t.Fatalf("expected prefix '{test}', got %q", m.prefix)
		}
	})

	t.Run("applies custom encode function via option", func(t *testing.T) {
		t.Parallel()

		called := false
		custom := func(_ *Method, _ string) (string, error) {
			called = true
			return "custom-encoded", nil
		}

		m := NewMethod("custom", "{custom}", WithBcryptParams(BcryptDefaultCost), WithEncodeFn(custom))

		result, err := m.Encode("password")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !called {
			t.Fatal("expected custom encodeFn to be called")
		}

		if result != "custom-encoded" {
			t.Fatalf("expected 'custom-encoded', got %q", result)
		}
	})
}

func TestMethod_Name(t *testing.T) {
	t.Parallel()

	t.Run("returns method name", func(t *testing.T) {
		t.Parallel()

		if Argon2.Name() != "Argon2" {
			t.Fatalf("expected 'Argon2', got %q", Argon2.Name())
		}
		if Bcrypt.Name() != "Bcrypt" {
			t.Fatalf("expected 'Bcrypt', got %q", Bcrypt.Name())
		}
	})
}

func TestMethod_Encode(t *testing.T) {
	t.Parallel()

	t.Run("bcrypt encodes successfully", func(t *testing.T) {
		t.Parallel()

		encoded, err := Bcrypt.Encode("test-password")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.HasPrefix(encoded, BcryptPrefixKey) {
			t.Fatalf("expected prefix %q, got %q", BcryptPrefixKey, encoded)
		}
	})

	t.Run("returns error for empty password", func(t *testing.T) {
		t.Parallel()

		_, err := Bcrypt.Encode("")
		if err == nil {
			t.Fatal("expected error")
		}

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}
	})

	t.Run("wraps encode function error", func(t *testing.T) {
		t.Parallel()

		fail := func(_ *Method, _ string) (string, error) {
			return "", errors.New("encode boom")
		}

		m := NewMethod("fail", "{fail}", WithBcryptParams(BcryptDefaultCost), WithEncodeFn(fail))

		_, err := m.Encode("password")
		if err == nil {
			t.Fatal("expected error")
		}

		if !strings.Contains(err.Error(), "encode boom") {
			t.Fatalf("expected 'encode boom' in error, got %q", err.Error())
		}
	})
}

func TestMethod_Verify(t *testing.T) {
	t.Parallel()

	t.Run("bcrypt verifies matching password", func(t *testing.T) {
		t.Parallel()

		encoded, err := Bcrypt.Encode("verify-me")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ok, err := Bcrypt.Verify(encoded, "verify-me")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatal("expected password to match")
		}
	})

	t.Run("bcrypt rejects wrong password", func(t *testing.T) {
		t.Parallel()

		encoded, err := Bcrypt.Encode("correct-password")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ok, err := Bcrypt.Verify(encoded, "wrong-password")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ok {
			t.Fatal("expected password not to match")
		}
	})

	t.Run("wraps verify function error", func(t *testing.T) {
		t.Parallel()

		fail := func(_ *Method, _ string, _ string) (bool, error) {
			return false, errors.New("verify boom")
		}

		m := NewMethod("fail", "{fail}", WithBcryptParams(BcryptDefaultCost), WithVerifyFn(fail))

		_, err := m.Verify("encoded", "raw")
		if err == nil {
			t.Fatal("expected error")
		}

		if !strings.Contains(err.Error(), "verify boom") {
			t.Fatalf("expected 'verify boom' in error, got %q", err.Error())
		}
	})
}

func TestMethod_UpgradeNeeded(t *testing.T) {
	t.Parallel()

	t.Run("bcrypt returns false for same cost", func(t *testing.T) {
		t.Parallel()

		encoded, err := Bcrypt.Encode("upgrade-test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		needed, err := Bcrypt.UpgradeNeeded(encoded)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if needed {
			t.Fatal("expected no upgrade needed for same cost")
		}
	})

	t.Run("wraps upgrade needed function error", func(t *testing.T) {
		t.Parallel()

		fail := func(_ *Method, _ string) (bool, error) {
			return false, errors.New("upgrade boom")
		}

		m := NewMethod("fail", "{fail}", WithBcryptParams(BcryptDefaultCost), WithUpgradeNeededFn(fail))

		_, err := m.UpgradeNeeded("encoded")
		if err == nil {
			t.Fatal("expected error")
		}

		if !strings.Contains(err.Error(), "upgrade boom") {
			t.Fatalf("expected 'upgrade boom' in error, got %q", err.Error())
		}
	})
}
