package hybrid

import (
	"errors"
	"strings"
	"testing"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("formats error with type and cause", func(t *testing.T) {
		t.Parallel()

		err := ErrEncrypt(errors.New("cause"))

		got := err.Error()
		if !strings.Contains(got, "hybrid") {
			t.Fatalf("expected 'hybrid' in error, got %q", got)
		}

		if !strings.Contains(got, HybridMethod) {
			t.Fatalf("expected type in error, got %q", got)
		}
	})
}

func TestErrAlgorithmNotSupported(t *testing.T) {
	t.Parallel()

	t.Run("returns domain Error with algorithm name", func(t *testing.T) {
		t.Parallel()

		err := ErrAlgorithmNotSupported("UNKNOWN")

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}

		if !strings.Contains(err.Error(), "UNKNOWN") {
			t.Fatalf("expected algorithm name in error, got %q", err.Error())
		}
	})
}

func TestErrKeyGeneration(t *testing.T) {
	t.Parallel()

	t.Run("wraps with ErrKeyGenerationFailed sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrKeyGeneration(errors.New("cause"))

		if !errors.Is(err, ErrKeyGenerationFailed) {
			t.Fatal("expected ErrKeyGenerationFailed")
		}
	})
}

func TestErrEncrypt(t *testing.T) {
	t.Parallel()

	t.Run("wraps with ErrEncryptionFailed sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrEncrypt(errors.New("cause"))

		if !errors.Is(err, ErrEncryptionFailed) {
			t.Fatal("expected ErrEncryptionFailed")
		}
	})
}

func TestErrDecrypt(t *testing.T) {
	t.Parallel()

	t.Run("wraps with ErrDecryptionFailed sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrDecrypt(errors.New("cause"))

		if !errors.Is(err, ErrDecryptionFailed) {
			t.Fatal("expected ErrDecryptionFailed")
		}
	})
}

func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	t.Run("ErrMethodIsNil is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrMethodIsNil == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrPublicKeyIsNil is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrPublicKeyIsNil == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrPrivateKeyIsNil is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrPrivateKeyIsNil == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrSuiteSetupFailed is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrSuiteSetupFailed == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrEncapsulationFailed is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrEncapsulationFailed == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrDecapsulationFailed is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrDecapsulationFailed == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrCiphertextTooShort is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrCiphertextTooShort == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrKeyTypeMismatch is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrKeyTypeMismatch == nil {
			t.Fatal("expected non-nil")
		}
	})
}
