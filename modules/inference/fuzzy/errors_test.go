package fuzzy

import (
	"errors"
	"testing"
)

func TestError_implements_error(t *testing.T) {
	t.Parallel()

	err := ErrInfer()
	if err.Error() == "" {
		t.Fatal("expected non-empty error message")
	}
}

func TestErrInfer(t *testing.T) {
	t.Parallel()

	err := ErrInfer()
	if !errors.Is(err, ErrNoRules) {
		t.Fatal("expected ErrNoRules")
	}
}

func TestErrInfer_withCause(t *testing.T) {
	t.Parallel()

	cause := errors.New("no rules defined")
	err := ErrInfer(cause)

	if !errors.Is(err, ErrNoRules) {
		t.Fatal("expected ErrNoRules")
	}

	if !errors.Is(err, cause) {
		t.Fatal("expected cause error")
	}
}

func TestErrValidation(t *testing.T) {
	t.Parallel()

	err := ErrValidation()
	if !errors.Is(err, ErrVariableNotFound) {
		t.Fatal("expected ErrVariableNotFound")
	}
}

func TestErrValidation_withCause(t *testing.T) {
	t.Parallel()

	cause := errors.New("missing temperature")
	err := ErrValidation(cause)

	if !errors.Is(err, ErrVariableNotFound) {
		t.Fatal("expected ErrVariableNotFound")
	}

	if !errors.Is(err, cause) {
		t.Fatal("expected cause error")
	}
}

func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	t.Run("ErrVariableNotFound", func(t *testing.T) {
		t.Parallel()

		if ErrVariableNotFound.Error() != "variable not found" {
			t.Fatalf("unexpected: %s", ErrVariableNotFound.Error())
		}
	})

	t.Run("ErrTermNotFound", func(t *testing.T) {
		t.Parallel()

		if ErrTermNotFound.Error() != "term not found" {
			t.Fatalf("unexpected: %s", ErrTermNotFound.Error())
		}
	})

	t.Run("ErrNoRules", func(t *testing.T) {
		t.Parallel()

		if ErrNoRules.Error() != "no rules provided" {
			t.Fatalf("unexpected: %s", ErrNoRules.Error())
		}
	})

	t.Run("ErrNoInputs", func(t *testing.T) {
		t.Parallel()

		if ErrNoInputs.Error() != "no inputs provided" {
			t.Fatalf("unexpected: %s", ErrNoInputs.Error())
		}
	})

	t.Run("ErrInputOutOfRange", func(t *testing.T) {
		t.Parallel()

		if ErrInputOutOfRange.Error() != "input value out of variable range" {
			t.Fatalf("unexpected: %s", ErrInputOutOfRange.Error())
		}
	})
}

func TestError_type(t *testing.T) {
	t.Parallel()

	err := ErrInfer()

	var fuzzyErr *Error

	if !errors.As(err, &fuzzyErr) {
		t.Fatal("expected *Error type")
	}

	if fuzzyErr.Type != FuzzyType {
		t.Fatalf("expected type %s, got %s", FuzzyType, fuzzyErr.Type)
	}
}
