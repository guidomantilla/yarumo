package kdfs

import (
	"crypto"
	"errors"
	"testing"
)

func TestHkdfDerive(t *testing.T) {
	t.Parallel()

	t.Run("returns bare ErrMethodIsNil for nil method", func(t *testing.T) {
		t.Parallel()

		_, err := hkdfDerive(nil, []byte("ikm"), nil, nil, 32)
		if !errors.Is(err, ErrMethodIsNil) {
			t.Fatalf("expected ErrMethodIsNil, got %v", err)
		}
	})

	t.Run("returns bare ErrSecretIsNil for nil secret", func(t *testing.T) {
		t.Parallel()

		_, err := hkdfDerive(HKDF_with_SHA256, nil, nil, nil, 32)
		if !errors.Is(err, ErrSecretIsNil) {
			t.Fatalf("expected ErrSecretIsNil, got %v", err)
		}
	})

	t.Run("returns bare ErrLengthInvalid for length=0", func(t *testing.T) {
		t.Parallel()

		_, err := hkdfDerive(HKDF_with_SHA256, []byte("ikm"), nil, nil, 0)
		if !errors.Is(err, ErrLengthInvalid) {
			t.Fatalf("expected ErrLengthInvalid, got %v", err)
		}
	})

	t.Run("returns bare ErrHashNotAvailable for unavailable hash", func(t *testing.T) {
		t.Parallel()

		bad := NewMethod("bad-hkdf", 0)

		_, err := hkdfDerive(bad, []byte("ikm"), nil, nil, 32)
		if !errors.Is(err, ErrHashNotAvailable) {
			t.Fatalf("expected ErrHashNotAvailable, got %v", err)
		}
	})

	t.Run("succeeds with valid inputs", func(t *testing.T) {
		t.Parallel()

		out, err := hkdfDerive(HKDF_with_SHA256, []byte("ikm"), []byte("salt"), []byte("info"), 32)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(out) != 32 {
			t.Fatalf("expected 32 bytes, got %d", len(out))
		}
	})
}

func TestPbkdf2Derive(t *testing.T) {
	t.Parallel()

	t.Run("returns bare ErrMethodIsNil for nil method", func(t *testing.T) {
		t.Parallel()

		_, err := pbkdf2Derive(nil, []byte("password"), []byte("salt"), nil, 32)
		if !errors.Is(err, ErrMethodIsNil) {
			t.Fatalf("expected ErrMethodIsNil, got %v", err)
		}
	})

	t.Run("returns bare ErrSecretIsNil for nil secret", func(t *testing.T) {
		t.Parallel()

		_, err := pbkdf2Derive(PBKDF2_with_SHA256, nil, []byte("salt"), nil, 32)
		if !errors.Is(err, ErrSecretIsNil) {
			t.Fatalf("expected ErrSecretIsNil, got %v", err)
		}
	})

	t.Run("returns bare ErrSaltIsNil for nil salt", func(t *testing.T) {
		t.Parallel()

		_, err := pbkdf2Derive(PBKDF2_with_SHA256, []byte("password"), nil, nil, 32)
		if !errors.Is(err, ErrSaltIsNil) {
			t.Fatalf("expected ErrSaltIsNil, got %v", err)
		}
	})

	t.Run("returns bare ErrLengthInvalid for zero length", func(t *testing.T) {
		t.Parallel()

		_, err := pbkdf2Derive(PBKDF2_with_SHA256, []byte("password"), []byte("salt"), nil, 0)
		if !errors.Is(err, ErrLengthInvalid) {
			t.Fatalf("expected ErrLengthInvalid, got %v", err)
		}
	})

	t.Run("returns bare ErrParamsMissing when pbkdf2Params unset", func(t *testing.T) {
		t.Parallel()

		bare := &Method{name: "bare", kind: crypto.SHA256, deriveFn: pbkdf2Derive}

		_, err := pbkdf2Derive(bare, []byte("password"), []byte("salt"), nil, 32)
		if !errors.Is(err, ErrParamsMissing) {
			t.Fatalf("expected ErrParamsMissing, got %v", err)
		}
	})

	t.Run("returns bare ErrHashNotAvailable for unavailable hash", func(t *testing.T) {
		t.Parallel()

		bad := NewMethod("bad-pbkdf2", 0, WithPbkdf2Iterations(1))

		_, err := pbkdf2Derive(bad, []byte("password"), []byte("salt"), nil, 32)
		if !errors.Is(err, ErrHashNotAvailable) {
			t.Fatalf("expected ErrHashNotAvailable, got %v", err)
		}
	})

	t.Run("succeeds with valid inputs", func(t *testing.T) {
		t.Parallel()

		out, err := pbkdf2Derive(PBKDF2_with_SHA256, []byte("password"), []byte("salt"), nil, 32)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(out) != 32 {
			t.Fatalf("expected 32 bytes, got %d", len(out))
		}
	})
}

func TestScryptDerive(t *testing.T) {
	t.Parallel()

	t.Run("returns bare ErrMethodIsNil for nil method", func(t *testing.T) {
		t.Parallel()

		_, err := scryptDerive(nil, []byte("password"), []byte("salt"), nil, 32)
		if !errors.Is(err, ErrMethodIsNil) {
			t.Fatalf("expected ErrMethodIsNil, got %v", err)
		}
	})

	t.Run("returns bare ErrSecretIsNil for nil secret", func(t *testing.T) {
		t.Parallel()

		_, err := scryptDerive(Scrypt_KDF, nil, []byte("salt"), nil, 32)
		if !errors.Is(err, ErrSecretIsNil) {
			t.Fatalf("expected ErrSecretIsNil, got %v", err)
		}
	})

	t.Run("returns bare ErrSaltIsNil for nil salt", func(t *testing.T) {
		t.Parallel()

		_, err := scryptDerive(Scrypt_KDF, []byte("password"), nil, nil, 32)
		if !errors.Is(err, ErrSaltIsNil) {
			t.Fatalf("expected ErrSaltIsNil, got %v", err)
		}
	})

	t.Run("returns bare ErrLengthInvalid for zero length", func(t *testing.T) {
		t.Parallel()

		_, err := scryptDerive(Scrypt_KDF, []byte("password"), []byte("salt"), nil, 0)
		if !errors.Is(err, ErrLengthInvalid) {
			t.Fatalf("expected ErrLengthInvalid, got %v", err)
		}
	})

	t.Run("returns bare ErrParamsMissing when scryptParams unset", func(t *testing.T) {
		t.Parallel()

		bare := &Method{name: "bare-scrypt", deriveFn: scryptDerive}

		_, err := scryptDerive(bare, []byte("password"), []byte("salt"), nil, 32)
		if !errors.Is(err, ErrParamsMissing) {
			t.Fatalf("expected ErrParamsMissing, got %v", err)
		}
	})

	t.Run("wraps underlying scrypt error via ErrDeriveFailed", func(t *testing.T) {
		t.Parallel()

		// Invalid scrypt params (N=3 is not a power of 2 > 1): NewMethod
		// won't install them, so build a Method directly to force the
		// scrypt.Key call to fail.
		bad := &Method{
			name:         "bad-scrypt",
			deriveFn:     scryptDerive,
			scryptParams: &scryptConfig{n: 3, r: 8, p: 1},
		}

		_, err := scryptDerive(bad, []byte("password"), []byte("salt"), nil, 32)
		if err == nil {
			t.Fatal("expected error from underlying scrypt.Key")
		}

		if !errors.Is(err, ErrDeriveFailed) {
			t.Fatalf("expected ErrDeriveFailed chain, got %v", err)
		}
	})

	t.Run("succeeds with small valid params", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("scrypt-small", 0, WithScryptParams(16, 1, 1))

		out, err := scryptDerive(m, []byte("password"), []byte("salt"), nil, 32)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(out) != 32 {
			t.Fatalf("expected 32 bytes, got %d", len(out))
		}
	})
}
