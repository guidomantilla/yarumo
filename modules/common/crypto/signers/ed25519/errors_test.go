package ed25519

import (
	"errors"
	"strings"
	"testing"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("formats error with type and cause", func(t *testing.T) {
		t.Parallel()

		err := ErrKeyGeneration(errors.New("cause"))

		got := err.Error()
		if !strings.Contains(got, "ed25519") {
			t.Fatalf("expected 'ed25519' in error, got %q", got)
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
			t.Fatalf("expected algorithm name, got %q", err.Error())
		}
	})
}

func TestErrKeyGeneration(t *testing.T) {
	t.Parallel()

	t.Run("wraps with sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrKeyGeneration(errors.New("cause"))

		if !errors.Is(err, ErrKeyGenerationFailed) {
			t.Fatal("expected ErrKeyGenerationFailed")
		}
	})
}

func TestErrSigning(t *testing.T) {
	t.Parallel()

	t.Run("wraps with sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrSigning(errors.New("cause"))

		if !errors.Is(err, ErrSigningFailed) {
			t.Fatal("expected ErrSigningFailed")
		}
	})
}

func TestErrVerification(t *testing.T) {
	t.Parallel()

	t.Run("wraps with sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrVerification(errors.New("cause"))

		if !errors.Is(err, ErrVerificationFailed) {
			t.Fatal("expected ErrVerificationFailed")
		}
	})
}
