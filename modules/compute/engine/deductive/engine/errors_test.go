package engine

import (
	"errors"
	"testing"
)

func TestErrEngine(t *testing.T) {
	t.Parallel()

	t.Run("wraps sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrEngine(ErrMaxIterations)

		if !errors.Is(err, ErrMaxIterations) {
			t.Fatal("expected ErrMaxIterations in chain")
		}
	})

	t.Run("wraps additional cause", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("too many rules")
		err := ErrEngine(cause)

		if !errors.Is(err, cause) {
			t.Fatal("expected cause in chain")
		}
	})

	t.Run("wraps multiple sentinels", func(t *testing.T) {
		t.Parallel()

		err := ErrEngine(ErrNoRules, ErrMaxDepth)

		if !errors.Is(err, ErrNoRules) {
			t.Fatal("expected ErrNoRules in chain")
		}

		if !errors.Is(err, ErrMaxDepth) {
			t.Fatal("expected ErrMaxDepth in chain")
		}
	})

	t.Run("is Error type", func(t *testing.T) {
		t.Parallel()

		err := ErrEngine(ErrMaxIterations)

		var engineErr *Error

		ok := errors.As(err, &engineErr)
		if !ok {
			t.Fatal("expected Error type")
		}

		if engineErr.Type != EngineType {
			t.Fatalf("expected type %s, got %s", EngineType, engineErr.Type)
		}
	})

	t.Run("zero args still wraps ErrEngineFailed", func(t *testing.T) {
		t.Parallel()

		err := ErrEngine()
		if !errors.Is(err, ErrEngineFailed) {
			t.Fatal("expected ErrEngineFailed in chain")
		}
	})
}

func TestErrEngineFailed(t *testing.T) {
	t.Parallel()

	if ErrEngineFailed == nil {
		t.Fatal("expected non-nil error")
	}

	if ErrEngineFailed.Error() != "engine operation failed" {
		t.Fatalf("unexpected message: %s", ErrEngineFailed.Error())
	}
}
