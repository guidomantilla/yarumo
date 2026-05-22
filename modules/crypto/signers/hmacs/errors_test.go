package hmacs

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
		if !strings.Contains(got, "hmac") {
			t.Fatalf("expected 'hmac' in error, got %q", got)
		}

		if !strings.Contains(got, HmacMethod) {
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
			t.Fatalf("expected algorithm name, got %q", err.Error())
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

func TestErrDigest(t *testing.T) {
	t.Parallel()

	t.Run("wraps with ErrDigestFailed sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrDigest(errors.New("cause"))

		if !errors.Is(err, ErrDigestFailed) {
			t.Fatal("expected ErrDigestFailed")
		}
	})
}

func TestErrValidation(t *testing.T) {
	t.Parallel()

	t.Run("wraps with ErrValidationFailed sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrValidation(errors.New("cause"))

		if !errors.Is(err, ErrValidationFailed) {
			t.Fatal("expected ErrValidationFailed")
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

	t.Run("ErrHashNotAvailable is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrHashNotAvailable == nil {
			t.Fatal("expected non-nil")
		}
	})
}
