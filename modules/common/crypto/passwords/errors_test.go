package passwords

import (
	"errors"
	"strings"
	"testing"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("formats error with type and cause", func(t *testing.T) {
		t.Parallel()

		err := ErrEncoding(ErrRawPasswordEmpty)

		got := err.Error()
		if !strings.Contains(got, "password") {
			t.Fatalf("expected 'password' in error, got %q", got)
		}
		if !strings.Contains(got, PasswordMethod) {
			t.Fatalf("expected type %q in error, got %q", PasswordMethod, got)
		}
	})
}

func TestErrEncoding(t *testing.T) {
	t.Parallel()

	t.Run("returns domain Error type", func(t *testing.T) {
		t.Parallel()

		err := ErrEncoding(ErrRawPasswordEmpty)

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}
	})

	t.Run("includes encoding failed sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrEncoding(ErrRawPasswordEmpty)

		if !strings.Contains(err.Error(), ErrEncodingFailed.Error()) {
			t.Fatalf("expected encoding failed in error, got %q", err.Error())
		}
	})
}

func TestErrVerification(t *testing.T) {
	t.Parallel()

	t.Run("returns domain Error type", func(t *testing.T) {
		t.Parallel()

		err := ErrVerification(ErrRawPasswordEmpty)

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}
	})

	t.Run("includes verification failed sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrVerification(ErrRawPasswordEmpty)

		if !strings.Contains(err.Error(), ErrVerificationFailed.Error()) {
			t.Fatalf("expected verification failed in error, got %q", err.Error())
		}
	})
}

func TestErrUpgradeCheck(t *testing.T) {
	t.Parallel()

	t.Run("returns domain Error type", func(t *testing.T) {
		t.Parallel()

		err := ErrUpgradeCheck(ErrEncodedPasswordEmpty)

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}
	})

	t.Run("includes upgrade check failed sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrUpgradeCheck(ErrEncodedPasswordEmpty)

		if !strings.Contains(err.Error(), ErrUpgradeCheckFailed.Error()) {
			t.Fatalf("expected upgrade check failed in error, got %q", err.Error())
		}
	})
}

func TestErrAlgorithmNotSupported(t *testing.T) {
	t.Parallel()

	t.Run("returns domain Error type", func(t *testing.T) {
		t.Parallel()

		err := ErrAlgorithmNotSupported("unknown")

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}
	})

	t.Run("includes algorithm name in message", func(t *testing.T) {
		t.Parallel()

		err := ErrAlgorithmNotSupported("unknown")

		if !strings.Contains(err.Error(), "unknown") {
			t.Fatalf("expected 'unknown' in error, got %q", err.Error())
		}
	})
}

func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	t.Run("ErrRawPasswordEmpty is not nil", func(t *testing.T) {
		t.Parallel()
		if ErrRawPasswordEmpty == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrEncodedPasswordEmpty is not nil", func(t *testing.T) {
		t.Parallel()
		if ErrEncodedPasswordEmpty == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrEncodedPasswordFormat is not nil", func(t *testing.T) {
		t.Parallel()
		if ErrEncodedPasswordFormat == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrSaltGenerationFailed is not nil", func(t *testing.T) {
		t.Parallel()
		if ErrSaltGenerationFailed == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrMethodConfigMissing is not nil", func(t *testing.T) {
		t.Parallel()
		if ErrMethodConfigMissing == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrBcryptCostNotAllowed is not nil", func(t *testing.T) {
		t.Parallel()
		if ErrBcryptCostNotAllowed == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrEncodingFailed is not nil", func(t *testing.T) {
		t.Parallel()
		if ErrEncodingFailed == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrVerificationFailed is not nil", func(t *testing.T) {
		t.Parallel()
		if ErrVerificationFailed == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrUpgradeCheckFailed is not nil", func(t *testing.T) {
		t.Parallel()
		if ErrUpgradeCheckFailed == nil {
			t.Fatal("expected non-nil")
		}
	})
}
