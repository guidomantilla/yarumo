package passwords

import (
	"crypto/sha512"
	"testing"
)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("returns defaults", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()

		if opts.encodeFn == nil {
			t.Fatal("expected default encodeFn")
		}
		if opts.verifyFn == nil {
			t.Fatal("expected default verifyFn")
		}
		if opts.upgradeNeededFn == nil {
			t.Fatal("expected default upgradeNeededFn")
		}
		if opts.argon2Params != nil {
			t.Fatal("expected nil argon2Params by default")
		}
	})
}

func TestWithEncodeFn(t *testing.T) {
	t.Parallel()

	t.Run("sets custom encode function", func(t *testing.T) {
		t.Parallel()

		called := false
		custom := func(_ *Method, _ string) (string, error) {
			called = true
			return "encoded", nil
		}

		opts := NewOptions(WithEncodeFn(custom))
		_, _ = opts.encodeFn(nil, "test")

		if !called {
			t.Fatal("expected custom encodeFn to be called")
		}
	})

	t.Run("ignores nil", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithEncodeFn(nil))

		if opts.encodeFn == nil {
			t.Fatal("expected default encodeFn preserved")
		}
	})
}

func TestWithVerifyFn(t *testing.T) {
	t.Parallel()

	t.Run("sets custom verify function", func(t *testing.T) {
		t.Parallel()

		called := false
		custom := func(_ *Method, _ string, _ string) (bool, error) {
			called = true
			return true, nil
		}

		opts := NewOptions(WithVerifyFn(custom))
		_, _ = opts.verifyFn(nil, "enc", "raw")

		if !called {
			t.Fatal("expected custom verifyFn to be called")
		}
	})

	t.Run("ignores nil", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithVerifyFn(nil))

		if opts.verifyFn == nil {
			t.Fatal("expected default verifyFn preserved")
		}
	})
}

func TestWithUpgradeNeededFn(t *testing.T) {
	t.Parallel()

	t.Run("sets custom upgrade needed function", func(t *testing.T) {
		t.Parallel()

		called := false
		custom := func(_ *Method, _ string) (bool, error) {
			called = true
			return false, nil
		}

		opts := NewOptions(WithUpgradeNeededFn(custom))
		_, _ = opts.upgradeNeededFn(nil, "enc")

		if !called {
			t.Fatal("expected custom upgradeNeededFn to be called")
		}
	})

	t.Run("ignores nil", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithUpgradeNeededFn(nil))

		if opts.upgradeNeededFn == nil {
			t.Fatal("expected default upgradeNeededFn preserved")
		}
	})
}

func TestWithArgon2Params(t *testing.T) {
	t.Parallel()

	t.Run("sets argon2 params with valid values", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithArgon2Params(Argon2Iterations, Argon2Memory, Argon2Threads, Argon2SaltLength, Argon2KeyLength))

		if opts.argon2Params == nil {
			t.Fatal("expected argon2Params to be set")
		}
		if opts.argon2Params.iterations != Argon2Iterations {
			t.Fatalf("expected iterations %d, got %d", Argon2Iterations, opts.argon2Params.iterations)
		}
	})

	t.Run("ignores invalid values", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithArgon2Params(0, 0, 0, 0, 0))

		if opts.argon2Params != nil {
			t.Fatal("expected argon2Params to be nil for invalid values")
		}
	})
}

func TestWithBcryptParams(t *testing.T) {
	t.Parallel()

	t.Run("sets bcrypt params with valid cost", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithBcryptParams(BcryptDefaultCost))

		if opts.bcryptParams == nil {
			t.Fatal("expected bcryptParams to be set")
		}
		if opts.bcryptParams.cost != BcryptDefaultCost {
			t.Fatalf("expected cost %d, got %d", BcryptDefaultCost, opts.bcryptParams.cost)
		}
	})

	t.Run("ignores invalid cost", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithBcryptParams(0))

		if opts.bcryptParams != nil {
			t.Fatal("expected bcryptParams to be nil for invalid cost")
		}
	})
}

func TestWithPbkdf2Params(t *testing.T) {
	t.Parallel()

	t.Run("sets pbkdf2 params with valid values", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithPbkdf2Params(Pbkdf2Iterations, Pbkdf2SaltLength, Pbkdf2KeyLength, sha512.New))

		if opts.pbkdf2Params == nil {
			t.Fatal("expected pbkdf2Params to be set")
		}
	})

	t.Run("ignores nil hash func", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithPbkdf2Params(Pbkdf2Iterations, Pbkdf2SaltLength, Pbkdf2KeyLength, nil))

		if opts.pbkdf2Params != nil {
			t.Fatal("expected pbkdf2Params to be nil for nil hashFunc")
		}
	})

	t.Run("ignores invalid iterations", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithPbkdf2Params(0, Pbkdf2SaltLength, Pbkdf2KeyLength, sha512.New))

		if opts.pbkdf2Params != nil {
			t.Fatal("expected pbkdf2Params to be nil for invalid iterations")
		}
	})
}

func TestWithScryptParams(t *testing.T) {
	t.Parallel()

	t.Run("sets scrypt params with valid values", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithScryptParams(ScryptN, ScryptR, ScryptP, ScryptSaltLength, ScryptKeyLength))

		if opts.scryptParams == nil {
			t.Fatal("expected scryptParams to be set")
		}
	})

	t.Run("ignores invalid values", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithScryptParams(0, 0, 0, 0, 0))

		if opts.scryptParams != nil {
			t.Fatal("expected scryptParams to be nil for invalid values")
		}
	})
}
