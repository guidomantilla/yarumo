package tokens

import (
	"errors"
	"strings"
	"testing"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("formats error with type and cause", func(t *testing.T) {
		t.Parallel()

		err := ErrGeneration(ErrSubjectEmpty)

		got := err.Error()
		if !strings.Contains(got, "token") {
			t.Fatalf("expected 'token' in error, got %q", got)
		}
		if !strings.Contains(got, TokenMethod) {
			t.Fatalf("expected type %q in error, got %q", TokenMethod, got)
		}
	})
}

func TestErrGeneration(t *testing.T) {
	t.Parallel()

	t.Run("returns domain Error type", func(t *testing.T) {
		t.Parallel()

		err := ErrGeneration(ErrSubjectEmpty)

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}
	})

	t.Run("includes generation failed sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrGeneration(ErrSubjectEmpty)

		if !strings.Contains(err.Error(), ErrGenerationFailed.Error()) {
			t.Fatalf("expected generation failed in error, got %q", err.Error())
		}
	})

	t.Run("wraps additional cause errors", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("signing error")
		err := ErrGeneration(cause)

		if !strings.Contains(err.Error(), cause.Error()) {
			t.Fatalf("expected cause in error, got %q", err.Error())
		}
	})
}

func TestErrValidation(t *testing.T) {
	t.Parallel()

	t.Run("returns domain Error type", func(t *testing.T) {
		t.Parallel()

		err := ErrValidation(ErrTokenEmpty)

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}
	})

	t.Run("includes validation failed sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrValidation(ErrTokenEmpty)

		if !strings.Contains(err.Error(), ErrValidationFailed.Error()) {
			t.Fatalf("expected validation failed in error, got %q", err.Error())
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

	t.Run("ErrSubjectEmpty is not nil", func(t *testing.T) {
		t.Parallel()
		if ErrSubjectEmpty == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrPayloadNil is not nil", func(t *testing.T) {
		t.Parallel()
		if ErrPayloadNil == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrTokenEmpty is not nil", func(t *testing.T) {
		t.Parallel()
		if ErrTokenEmpty == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrSigningKeyNil is not nil", func(t *testing.T) {
		t.Parallel()
		if ErrSigningKeyNil == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrSigningMethodNil is not nil", func(t *testing.T) {
		t.Parallel()
		if ErrSigningMethodNil == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrTokenParseFailed is not nil", func(t *testing.T) {
		t.Parallel()
		if ErrTokenParseFailed == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrTokenPayloadEmpty is not nil", func(t *testing.T) {
		t.Parallel()
		if ErrTokenPayloadEmpty == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrGenerationFailed is not nil", func(t *testing.T) {
		t.Parallel()
		if ErrGenerationFailed == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrValidationFailed is not nil", func(t *testing.T) {
		t.Parallel()
		if ErrValidationFailed == nil {
			t.Fatal("expected non-nil")
		}
	})
}
