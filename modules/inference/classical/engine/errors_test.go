package engine

import (
	"errors"
	"testing"
)

func TestErrForward(t *testing.T) {
	t.Parallel()

	t.Run("wraps sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrForward()

		if !errors.Is(err, ErrMaxIterations) {
			t.Fatal("expected ErrMaxIterations in chain")
		}
	})

	t.Run("wraps additional cause", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("too many rules")
		err := ErrForward(cause)

		if !errors.Is(err, cause) {
			t.Fatal("expected cause in chain")
		}
	})

	t.Run("is Error type", func(t *testing.T) {
		t.Parallel()

		err := ErrForward()

		var engineErr *Error

		ok := errors.As(err, &engineErr)
		if !ok {
			t.Fatal("expected Error type")
		}

		if engineErr.Type != EngineType {
			t.Fatalf("expected type %s, got %s", EngineType, engineErr.Type)
		}
	})
}

func TestErrBackward(t *testing.T) {
	t.Parallel()

	t.Run("wraps sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrBackward()

		if !errors.Is(err, ErrNoRules) {
			t.Fatal("expected ErrNoRules in chain")
		}
	})

	t.Run("wraps additional cause", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("goal missing")
		err := ErrBackward(cause)

		if !errors.Is(err, cause) {
			t.Fatal("expected cause in chain")
		}
	})

	t.Run("is Error type", func(t *testing.T) {
		t.Parallel()

		err := ErrBackward()

		var engineErr *Error

		ok := errors.As(err, &engineErr)
		if !ok {
			t.Fatal("expected Error type")
		}

		if engineErr.Type != EngineType {
			t.Fatalf("expected type %s, got %s", EngineType, engineErr.Type)
		}
	})
}
