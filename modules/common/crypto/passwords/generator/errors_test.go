package generator

import (
	"errors"
	"strings"
	"testing"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("includes type and inner error", func(t *testing.T) {
		t.Parallel()

		err := ErrValidation(ErrPasswordLength)

		msg := err.Error()
		if !strings.Contains(msg, PasswordGenerator) {
			t.Fatalf("error message %q missing type %q", msg, PasswordGenerator)
		}
		if !strings.Contains(msg, ErrPasswordLength.Error()) {
			t.Fatalf("error message %q missing inner sentinel", msg)
		}
		if !strings.Contains(msg, ErrValidationFailed.Error()) {
			t.Fatalf("error message %q missing validation failed sentinel", msg)
		}
	})
}

func TestErrConfiguration(t *testing.T) {
	t.Parallel()

	t.Run("wraps invalid option sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrConfiguration(ErrConstraintsExceedLength)

		if !errors.Is(err, ErrInvalidOption) {
			t.Fatalf("expected ErrInvalidOption in chain, got %v", err)
		}
		if !errors.Is(err, ErrConstraintsExceedLength) {
			t.Fatalf("expected ErrConstraintsExceedLength in chain, got %v", err)
		}
	})

	t.Run("unwraps to typed error", func(t *testing.T) {
		t.Parallel()

		err := ErrConfiguration(ErrConstraintsExceedLength)

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}
		if domErr.ErrorType() != PasswordGenerator {
			t.Fatalf("unexpected type: %q", domErr.ErrorType())
		}
	})
}

func TestErrGeneration(t *testing.T) {
	t.Parallel()

	t.Run("wraps generation failed sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrGeneration(ErrShuffleFailed)

		if !errors.Is(err, ErrGenerationFailed) {
			t.Fatalf("expected ErrGenerationFailed in chain, got %v", err)
		}
		if !errors.Is(err, ErrShuffleFailed) {
			t.Fatalf("expected ErrShuffleFailed in chain, got %v", err)
		}
	})
}

func TestErrValidation(t *testing.T) {
	t.Parallel()

	t.Run("wraps validation failed sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrValidation(ErrPasswordLength)

		if !errors.Is(err, ErrValidationFailed) {
			t.Fatalf("expected ErrValidationFailed in chain, got %v", err)
		}
		if !errors.Is(err, ErrPasswordLength) {
			t.Fatalf("expected ErrPasswordLength in chain, got %v", err)
		}
	})

	t.Run("returns *Error type", func(t *testing.T) {
		t.Parallel()

		err := ErrValidation(ErrPasswordNumbers)

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}
	})
}
