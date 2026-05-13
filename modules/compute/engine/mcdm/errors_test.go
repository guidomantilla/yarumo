package mcdm

import (
	"errors"
	"testing"
)

func TestError_implements_error(t *testing.T) {
	t.Parallel()

	err := ErrMCDM()
	if err.Error() == "" {
		t.Fatal("expected non-empty error message")
	}
}

func TestErrMCDM(t *testing.T) {
	t.Parallel()

	err := ErrMCDM(ErrEmptyMatrix)
	if !errors.Is(err, ErrEmptyMatrix) {
		t.Fatal("expected ErrEmptyMatrix")
	}
}

func TestErrMCDM_withCause(t *testing.T) {
	t.Parallel()

	cause := errors.New("custom cause")
	err := ErrMCDM(ErrInvalidMatrix, cause)

	if !errors.Is(err, ErrInvalidMatrix) {
		t.Fatal("expected ErrInvalidMatrix")
	}

	if !errors.Is(err, cause) {
		t.Fatal("expected cause error")
	}
}

func TestErrMCDM_multipleSentinels(t *testing.T) {
	t.Parallel()

	err := ErrMCDM(ErrNotSquareMatrix, ErrDimensionMismatch)

	if !errors.Is(err, ErrNotSquareMatrix) {
		t.Fatal("expected ErrNotSquareMatrix")
	}

	if !errors.Is(err, ErrDimensionMismatch) {
		t.Fatal("expected ErrDimensionMismatch")
	}
}

func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	t.Run("ErrInvalidMatrix", func(t *testing.T) {
		t.Parallel()

		if ErrInvalidMatrix.Error() != "invalid matrix dimensions" {
			t.Fatalf("unexpected: %s", ErrInvalidMatrix.Error())
		}
	})

	t.Run("ErrNotSquareMatrix", func(t *testing.T) {
		t.Parallel()

		if ErrNotSquareMatrix.Error() != "matrix must be square" {
			t.Fatalf("unexpected: %s", ErrNotSquareMatrix.Error())
		}
	})

	t.Run("ErrInconsistentMatrix", func(t *testing.T) {
		t.Parallel()

		if ErrInconsistentMatrix.Error() != "pairwise matrix is inconsistent" {
			t.Fatalf("unexpected: %s", ErrInconsistentMatrix.Error())
		}
	})

	t.Run("ErrEmptyMatrix", func(t *testing.T) {
		t.Parallel()

		if ErrEmptyMatrix.Error() != "matrix is empty" {
			t.Fatalf("unexpected: %s", ErrEmptyMatrix.Error())
		}
	})

	t.Run("ErrInvalidWeight", func(t *testing.T) {
		t.Parallel()

		if ErrInvalidWeight.Error() != "weights must be positive" {
			t.Fatalf("unexpected: %s", ErrInvalidWeight.Error())
		}
	})

	t.Run("ErrDimensionMismatch", func(t *testing.T) {
		t.Parallel()

		if ErrDimensionMismatch.Error() != "dimensions do not match" {
			t.Fatalf("unexpected: %s", ErrDimensionMismatch.Error())
		}
	})

	t.Run("ErrEmptyInput", func(t *testing.T) {
		t.Parallel()

		if ErrEmptyInput.Error() != "empty input" {
			t.Fatalf("unexpected: %s", ErrEmptyInput.Error())
		}
	})

	t.Run("ErrMCDMFailed", func(t *testing.T) {
		t.Parallel()

		if ErrMCDMFailed.Error() != "mcdm operation failed" {
			t.Fatalf("unexpected: %s", ErrMCDMFailed.Error())
		}
	})
}

func TestError_type(t *testing.T) {
	t.Parallel()

	err := ErrMCDM(ErrEmptyMatrix)

	var mcdmErr *Error

	if !errors.As(err, &mcdmErr) {
		t.Fatal("expected *Error type")
	}

	if mcdmErr.Type != MCDMType {
		t.Fatalf("expected type %s, got %s", MCDMType, mcdmErr.Type)
	}
}

func TestErrMCDM_zeroArgs(t *testing.T) {
	t.Parallel()

	err := ErrMCDM()
	if !errors.Is(err, ErrMCDMFailed) {
		t.Fatal("expected ErrMCDMFailed in chain")
	}
}
