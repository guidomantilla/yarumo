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

func TestWithSecureDefaults(t *testing.T) {
	t.Parallel()

	t.Run("argon2id prefix populates argon2id params", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("SecureArgon2id", Argon2idPrefixKey, WithSecureDefaults())

		if m.argon2Params == nil {
			t.Fatal("expected argon2Params to be set")
		}
		if m.argon2Params.iterations != SecureArgon2Iterations {
			t.Fatalf("expected iterations %d, got %d", SecureArgon2Iterations, m.argon2Params.iterations)
		}
		if m.argon2Params.memory != SecureArgon2Memory {
			t.Fatalf("expected memory %d, got %d", SecureArgon2Memory, m.argon2Params.memory)
		}
		if m.argon2Params.threads != SecureArgon2Threads {
			t.Fatalf("expected threads %d, got %d", SecureArgon2Threads, m.argon2Params.threads)
		}
		if m.argon2Params.useArgon2i {
			t.Fatal("expected useArgon2i to be false for argon2id prefix")
		}
	})

	t.Run("legacy argon2 prefix routes to argon2id profile", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("LegacyArgon2", Argon2PrefixKey, WithSecureDefaults())

		if m.argon2Params == nil {
			t.Fatal("expected argon2Params to be set")
		}
		if m.argon2Params.useArgon2i {
			t.Fatal("expected useArgon2i to be false for legacy argon2 prefix")
		}
	})

	t.Run("argon2i prefix selects useArgon2i variant", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("SecureArgon2i", Argon2iPrefixKey, WithSecureDefaults())

		if m.argon2Params == nil {
			t.Fatal("expected argon2Params to be set")
		}
		if !m.argon2Params.useArgon2i {
			t.Fatal("expected useArgon2i to be true for argon2i prefix")
		}
	})

	t.Run("bcrypt prefix populates bcrypt params", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("SecureBcrypt", BcryptPrefixKey, WithSecureDefaults())

		if m.bcryptParams == nil {
			t.Fatal("expected bcryptParams to be set")
		}
		if m.bcryptParams.cost != SecureBcryptCost {
			t.Fatalf("expected cost %d, got %d", SecureBcryptCost, m.bcryptParams.cost)
		}
	})

	t.Run("pbkdf2 prefix populates pbkdf2 params", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("SecurePbkdf2", Pbkdf2PrefixKey, WithSecureDefaults())

		if m.pbkdf2Params == nil {
			t.Fatal("expected pbkdf2Params to be set")
		}
		if m.pbkdf2Params.iterations != SecurePbkdf2Iterations {
			t.Fatalf("expected iterations %d, got %d", SecurePbkdf2Iterations, m.pbkdf2Params.iterations)
		}
		if m.pbkdf2Params.hashFunc == nil {
			t.Fatal("expected hashFunc to be set")
		}
	})

	t.Run("scrypt prefix populates scrypt params", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("SecureScrypt", ScryptPrefixKey, WithSecureDefaults())

		if m.scryptParams == nil {
			t.Fatal("expected scryptParams to be set")
		}
		if m.scryptParams.n != SecureScryptN {
			t.Fatalf("expected n %d, got %d", SecureScryptN, m.scryptParams.n)
		}
		if m.scryptParams.r != SecureScryptR {
			t.Fatalf("expected r %d, got %d", SecureScryptR, m.scryptParams.r)
		}
		if m.scryptParams.p != SecureScryptP {
			t.Fatalf("expected p %d, got %d", SecureScryptP, m.scryptParams.p)
		}
	})

	t.Run("unknown prefix is a no-op", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithSecureDefaults())

		if opts.argon2Params != nil {
			t.Fatal("expected argon2Params to remain nil")
		}
		if opts.bcryptParams != nil {
			t.Fatal("expected bcryptParams to remain nil")
		}
		if opts.pbkdf2Params != nil {
			t.Fatal("expected pbkdf2Params to remain nil")
		}
		if opts.scryptParams != nil {
			t.Fatal("expected scryptParams to remain nil")
		}
	})

	t.Run("custom prefix is a no-op", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("Custom", "{custom}", WithSecureDefaults(), WithBcryptParams(BcryptDefaultCost))

		if m.argon2Params != nil {
			t.Fatal("expected argon2Params to remain nil for unknown prefix")
		}
		if m.bcryptParams == nil {
			t.Fatal("expected bcryptParams set by the trailing WithBcryptParams")
		}
	})
}

func TestWithSecureDefaults_EncodeAndUpgradeNeeded(t *testing.T) {
	t.Parallel()

	t.Run("argon2id encodes and reports no upgrade needed", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("SecureArgon2id_E2E", Argon2idPrefixKey, WithSecureDefaults())

		encoded, err := m.Encode("secure-pwd")
		if err != nil {
			t.Fatalf("unexpected encode error: %v", err)
		}
		if encoded == "" {
			t.Fatal("expected non-empty encoded output")
		}

		needed, err := m.UpgradeNeeded(encoded)
		if err != nil {
			t.Fatalf("unexpected upgrade check error: %v", err)
		}
		if needed {
			t.Fatal("expected UpgradeNeeded=false immediately after encode")
		}
	})

	t.Run("argon2i encodes and reports no upgrade needed", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("SecureArgon2i_E2E", Argon2iPrefixKey, WithSecureDefaults())

		encoded, err := m.Encode("secure-pwd")
		if err != nil {
			t.Fatalf("unexpected encode error: %v", err)
		}

		needed, err := m.UpgradeNeeded(encoded)
		if err != nil {
			t.Fatalf("unexpected upgrade check error: %v", err)
		}
		if needed {
			t.Fatal("expected UpgradeNeeded=false immediately after encode")
		}
	})

	t.Run("bcrypt encodes and reports no upgrade needed", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("SecureBcrypt_E2E", BcryptPrefixKey, WithSecureDefaults())

		encoded, err := m.Encode("secure-pwd")
		if err != nil {
			t.Fatalf("unexpected encode error: %v", err)
		}

		needed, err := m.UpgradeNeeded(encoded)
		if err != nil {
			t.Fatalf("unexpected upgrade check error: %v", err)
		}
		if needed {
			t.Fatal("expected UpgradeNeeded=false immediately after encode")
		}
	})

	t.Run("pbkdf2 encodes and reports no upgrade needed", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("SecurePbkdf2_E2E", Pbkdf2PrefixKey, WithSecureDefaults())

		encoded, err := m.Encode("secure-pwd")
		if err != nil {
			t.Fatalf("unexpected encode error: %v", err)
		}

		needed, err := m.UpgradeNeeded(encoded)
		if err != nil {
			t.Fatalf("unexpected upgrade check error: %v", err)
		}
		if needed {
			t.Fatal("expected UpgradeNeeded=false immediately after encode")
		}
	})

	t.Run("scrypt encodes and reports no upgrade needed", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("SecureScrypt_E2E", ScryptPrefixKey, WithSecureDefaults())

		encoded, err := m.Encode("secure-pwd")
		if err != nil {
			t.Fatalf("unexpected encode error: %v", err)
		}

		needed, err := m.UpgradeNeeded(encoded)
		if err != nil {
			t.Fatalf("unexpected upgrade check error: %v", err)
		}
		if needed {
			t.Fatal("expected UpgradeNeeded=false immediately after encode")
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
