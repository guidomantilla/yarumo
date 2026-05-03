package causal

import (
	"errors"
	"testing"
)

func TestErrCausal(t *testing.T) {
	t.Parallel()

	t.Run("wraps sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrCausal(ErrCyclicModel)
		if !errors.Is(err, ErrCyclicModel) {
			t.Fatal("expected ErrCyclicModel")
		}
	})

	t.Run("wraps additional cause", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("graph has back edge")
		err := ErrCausal(ErrCyclicModel, cause)

		if !errors.Is(err, ErrCyclicModel) {
			t.Fatal("expected ErrCyclicModel")
		}

		if !errors.Is(err, cause) {
			t.Fatal("expected cause error")
		}
	})

	t.Run("wraps multiple sentinels", func(t *testing.T) {
		t.Parallel()

		err := ErrCausal(ErrVariableNotFound, ErrNilEquation)

		if !errors.Is(err, ErrVariableNotFound) {
			t.Fatal("expected ErrVariableNotFound")
		}

		if !errors.Is(err, ErrNilEquation) {
			t.Fatal("expected ErrNilEquation")
		}
	})

	t.Run("is Error type", func(t *testing.T) {
		t.Parallel()

		err := ErrCausal(ErrCyclicModel)

		var causalErr *Error

		if !errors.As(err, &causalErr) {
			t.Fatal("expected *Error type")
		}

		if causalErr.Type != CausalType {
			t.Fatalf("expected type %s, got %s", CausalType, causalErr.Type)
		}
	})
}

func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	t.Run("ErrCyclicModel", func(t *testing.T) {
		t.Parallel()

		if ErrCyclicModel.Error() != "model contains a cycle" {
			t.Fatalf("unexpected: %s", ErrCyclicModel.Error())
		}
	})

	t.Run("ErrVariableNotFound", func(t *testing.T) {
		t.Parallel()

		if ErrVariableNotFound.Error() != "variable not found in model" {
			t.Fatalf("unexpected: %s", ErrVariableNotFound.Error())
		}
	})

	t.Run("ErrDuplicateVariable", func(t *testing.T) {
		t.Parallel()

		if ErrDuplicateVariable.Error() != "duplicate variable" {
			t.Fatalf("unexpected: %s", ErrDuplicateVariable.Error())
		}
	})

	t.Run("ErrNilEquation", func(t *testing.T) {
		t.Parallel()

		if ErrNilEquation.Error() != "equation function is nil" {
			t.Fatalf("unexpected: %s", ErrNilEquation.Error())
		}
	})

	t.Run("ErrParentNotFound", func(t *testing.T) {
		t.Parallel()

		if ErrParentNotFound.Error() != "parent variable not found" {
			t.Fatalf("unexpected: %s", ErrParentNotFound.Error())
		}
	})
}
