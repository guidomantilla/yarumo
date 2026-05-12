package kdfs

import (
	"errors"
	"strings"
	"testing"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("formats error with type and cause", func(t *testing.T) {
		t.Parallel()

		err := ErrDerive(errors.New("cause"))

		got := err.Error()
		if !strings.Contains(got, "kdf") {
			t.Fatalf("expected 'kdf' in error, got %q", got)
		}

		if !strings.Contains(got, KdfMethod) {
			t.Fatalf("expected type %q in error, got %q", KdfMethod, got)
		}
	})
}

func TestErrAlgorithmNotSupported(t *testing.T) {
	t.Parallel()

	t.Run("returns domain Error with algorithm name", func(t *testing.T) {
		t.Parallel()

		err := ErrAlgorithmNotSupported("UNKNOWN_KDF")

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}

		if !strings.Contains(err.Error(), "UNKNOWN_KDF") {
			t.Fatalf("expected algorithm name, got %q", err.Error())
		}
	})
}

func TestErrDerive(t *testing.T) {
	t.Parallel()

	t.Run("wraps with ErrDeriveFailed sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrDerive(errors.New("cause"))

		if !errors.Is(err, ErrDeriveFailed) {
			t.Fatal("expected ErrDeriveFailed in chain")
		}
	})

	t.Run("wraps multiple causes", func(t *testing.T) {
		t.Parallel()

		a := errors.New("a")
		b := errors.New("b")

		err := ErrDerive(a, b)

		if !errors.Is(err, a) {
			t.Fatal("expected error to contain a")
		}

		if !errors.Is(err, b) {
			t.Fatal("expected error to contain b")
		}
	})
}

func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	t.Run("ErrMethodIsNil is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrMethodIsNil == nil {
			t.Fatal("expected non-nil sentinel")
		}
	})

	t.Run("ErrSecretIsNil is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrSecretIsNil == nil {
			t.Fatal("expected non-nil sentinel")
		}
	})

	t.Run("ErrSaltIsNil is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrSaltIsNil == nil {
			t.Fatal("expected non-nil sentinel")
		}
	})

	t.Run("ErrLengthInvalid is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrLengthInvalid == nil {
			t.Fatal("expected non-nil sentinel")
		}
	})

	t.Run("ErrHashNotAvailable is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrHashNotAvailable == nil {
			t.Fatal("expected non-nil sentinel")
		}
	})

	t.Run("ErrParamsMissing is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrParamsMissing == nil {
			t.Fatal("expected non-nil sentinel")
		}
	})
}
