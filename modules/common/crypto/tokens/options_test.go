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
		if opts.signingKey != nil {
			t.Fatalf("expected nil signing key, got %v", opts.signingKey)
		}
		if opts.verifyingKey != nil {
			t.Fatalf("expected nil verifying key, got %v", opts.verifyingKey)
		}
	})
}

func TestWithGeneratedKey(t *testing.T) {
	t.Parallel()

	t.Run("populates both keys with 64 bytes", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithGeneratedKey())

		signing, ok := opts.signingKey.([]byte)
		if !ok {
			t.Fatalf("expected signing key []byte, got %T", opts.signingKey)
		}
		verifying, ok := opts.verifyingKey.([]byte)
		if !ok {
			t.Fatalf("expected verifying key []byte, got %T", opts.verifyingKey)
		}
		if len(signing) != 64 {
			t.Fatalf("expected 64-byte signing key, got %d", len(signing))
		}
		if len(verifying) != 64 {
			t.Fatalf("expected 64-byte verifying key, got %d", len(verifying))
		}
		if string(signing) != string(verifying) {
			t.Fatal("expected signing and verifying keys to be identical")
		}
	})

	t.Run("entropy draw is fresh on each call", func(t *testing.T) {
		t.Parallel()

		a := NewOptions(WithGeneratedKey())
		b := NewOptions(WithGeneratedKey())

		aKey, ok := a.signingKey.([]byte)
		if !ok {
			t.Fatalf("expected a.signingKey []byte, got %T", a.signingKey)
		}
		bKey, ok := b.signingKey.([]byte)
		if !ok {
			t.Fatalf("expected b.signingKey []byte, got %T", b.signingKey)
		}
		if string(aKey) == string(bKey) {
			t.Fatal("expected distinct random keys across two calls")
		}
	})

	t.Run("later WithKey overrides generated key", func(t *testing.T) {
		t.Parallel()

		manual := []byte("manual-secret-key-1234567890")
		opts := NewOptions(WithGeneratedKey(), WithKey(manual))

		signing, ok := opts.signingKey.([]byte)
		if !ok {
			t.Fatalf("expected signing key []byte, got %T", opts.signingKey)
		}
		verifying, ok := opts.verifyingKey.([]byte)
		if !ok {
			t.Fatalf("expected verifying key []byte, got %T", opts.verifyingKey)
		}
		if string(signing) != string(manual) {
			t.Fatal("expected manual key to override generated signing key")
		}
		if string(verifying) != string(manual) {
			t.Fatal("expected manual key to override generated verifying key")
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

		signing, ok := opts.signingKey.([]byte)
		if !ok {
			t.Fatalf("expected signing key []byte, got %T", opts.signingKey)
		}
		verifying, ok := opts.verifyingKey.([]byte)
		if !ok {
			t.Fatalf("expected verifying key []byte, got %T", opts.verifyingKey)
		}
		if string(signing) != string(key) {
			t.Fatal("expected signing key to match")
		}
		if string(verifying) != string(key) {
			t.Fatal("expected verifying key to match")
		}
	})

	t.Run("ignores empty key", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithKey([]byte{}))

		if opts.signingKey != nil {
			t.Fatalf("expected nil signing key default preserved, got %v", opts.signingKey)
		}
		if opts.verifyingKey != nil {
			t.Fatalf("expected nil verifying key default preserved, got %v", opts.verifyingKey)
		}
	})
}

func TestWithSigningKey(t *testing.T) {
	t.Parallel()

	t.Run("sets signing key independently", func(t *testing.T) {
		t.Parallel()

		key := []byte("signing-key-0123456789")
		opts := NewOptions(WithSigningKey(key))

		signing, ok := opts.signingKey.([]byte)
		if !ok {
			t.Fatalf("expected signing key []byte, got %T", opts.signingKey)
		}
		if string(signing) != string(key) {
			t.Fatal("expected signing key to match")
		}
	})

	t.Run("ignores empty key", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithSigningKey([]byte{}))

		if opts.signingKey != nil {
			t.Fatalf("expected nil signing key default preserved, got %v", opts.signingKey)
		}
	})
}

func TestWithVerifyingKey(t *testing.T) {
	t.Parallel()

	t.Run("sets verifying key independently", func(t *testing.T) {
		t.Parallel()

		key := []byte("verifying-key-0123456789")
		opts := NewOptions(WithVerifyingKey(key))

		verifying, ok := opts.verifyingKey.([]byte)
		if !ok {
			t.Fatalf("expected verifying key []byte, got %T", opts.verifyingKey)
		}
		if string(verifying) != string(key) {
			t.Fatal("expected verifying key to match")
		}
	})

	t.Run("ignores empty key", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithVerifyingKey([]byte{}))

		if opts.verifyingKey != nil {
			t.Fatalf("expected nil verifying key default preserved, got %v", opts.verifyingKey)
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
