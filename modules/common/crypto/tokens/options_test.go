package tokens

import (
	"testing"
	"time"
)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("returns defaults", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()

		if opts.generateFn == nil {
			t.Fatal("expected default generateFn")
		}
		if opts.validateFn == nil {
			t.Fatal("expected default validateFn")
		}
		if opts.timeout != 24*time.Hour {
			t.Fatalf("expected 24h timeout, got %v", opts.timeout)
		}
		if len(opts.signingKey) != 64 {
			t.Fatalf("expected 64-byte signing key, got %d", len(opts.signingKey))
		}
		if len(opts.verifyingKey) != 64 {
			t.Fatalf("expected 64-byte verifying key, got %d", len(opts.verifyingKey))
		}
	})
}

func TestWithIssuer(t *testing.T) {
	t.Parallel()

	t.Run("sets issuer", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithIssuer("test-issuer"))

		if opts.issuer != "test-issuer" {
			t.Fatalf("expected 'test-issuer', got %q", opts.issuer)
		}
	})

	t.Run("ignores empty issuer", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithIssuer("valid-issuer"), WithIssuer(""))

		if opts.issuer != "valid-issuer" {
			t.Fatalf("expected 'valid-issuer' preserved, got %q", opts.issuer)
		}
	})
}

func TestWithTimeout(t *testing.T) {
	t.Parallel()

	t.Run("sets positive timeout", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithTimeout(48 * time.Hour))

		if opts.timeout != 48*time.Hour {
			t.Fatalf("expected 48h, got %v", opts.timeout)
		}
	})

	t.Run("ignores zero timeout", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithTimeout(0))

		if opts.timeout != 24*time.Hour {
			t.Fatalf("expected default 24h, got %v", opts.timeout)
		}
	})

	t.Run("ignores negative timeout", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithTimeout(-1 * time.Hour))

		if opts.timeout != 24*time.Hour {
			t.Fatalf("expected default 24h, got %v", opts.timeout)
		}
	})
}

func TestWithKey(t *testing.T) {
	t.Parallel()

	t.Run("sets both keys", func(t *testing.T) {
		t.Parallel()

		key := []byte("test-key-0123456789")
		opts := NewOptions(WithKey(key))

		if string(opts.signingKey) != string(key) {
			t.Fatal("expected signing key to match")
		}
		if string(opts.verifyingKey) != string(key) {
			t.Fatal("expected verifying key to match")
		}
	})

	t.Run("ignores empty key", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithKey([]byte{}))

		if len(opts.signingKey) != 64 {
			t.Fatalf("expected default 64-byte key, got %d", len(opts.signingKey))
		}
	})
}

func TestWithSigningKey(t *testing.T) {
	t.Parallel()

	t.Run("sets signing key independently", func(t *testing.T) {
		t.Parallel()

		key := []byte("signing-key-0123456789")
		opts := NewOptions(WithSigningKey(key))

		if string(opts.signingKey) != string(key) {
			t.Fatal("expected signing key to match")
		}
	})

	t.Run("ignores empty key", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithSigningKey([]byte{}))

		if len(opts.signingKey) != 64 {
			t.Fatalf("expected default key preserved, got %d bytes", len(opts.signingKey))
		}
	})
}

func TestWithVerifyingKey(t *testing.T) {
	t.Parallel()

	t.Run("sets verifying key independently", func(t *testing.T) {
		t.Parallel()

		key := []byte("verifying-key-0123456789")
		opts := NewOptions(WithVerifyingKey(key))

		if string(opts.verifyingKey) != string(key) {
			t.Fatal("expected verifying key to match")
		}
	})

	t.Run("ignores empty key", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithVerifyingKey([]byte{}))

		if len(opts.verifyingKey) != 64 {
			t.Fatalf("expected default key preserved, got %d bytes", len(opts.verifyingKey))
		}
	})
}

func TestWithGenerateFn(t *testing.T) {
	t.Parallel()

	t.Run("sets custom generate function", func(t *testing.T) {
		t.Parallel()

		called := false
		custom := func(_ *Method, _ string, _ Payload) (string, error) {
			called = true
			return "custom-token", nil
		}

		opts := NewOptions(WithGenerateFn(custom))
		_, _ = opts.generateFn(nil, "sub", Payload{})

		if !called {
			t.Fatal("expected custom generateFn to be called")
		}
	})

	t.Run("ignores nil", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithGenerateFn(nil))

		if opts.generateFn == nil {
			t.Fatal("expected default generateFn preserved")
		}
	})
}

func TestWithValidateFn(t *testing.T) {
	t.Parallel()

	t.Run("sets custom validate function", func(t *testing.T) {
		t.Parallel()

		called := false
		custom := func(_ *Method, _ string) (Payload, error) {
			called = true
			return Payload{}, nil
		}

		opts := NewOptions(WithValidateFn(custom))
		_, _ = opts.validateFn(nil, "token")

		if !called {
			t.Fatal("expected custom validateFn to be called")
		}
	})

	t.Run("ignores nil", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithValidateFn(nil))

		if opts.validateFn == nil {
			t.Fatal("expected default validateFn preserved")
		}
	})
}
