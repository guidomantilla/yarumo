package fuzzy

import (
	"errors"
	"testing"
)

func TestErrFuzzy(t *testing.T) {
	t.Parallel()

	t.Run("wraps sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrFuzzy(ErrNoRules)
		if !errors.Is(err, ErrNoRules) {
			t.Fatal("expected ErrNoRules")
		}
	})

	t.Run("wraps additional cause", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("no rules defined")
		err := ErrFuzzy(ErrNoRules, cause)

		if !errors.Is(err, ErrNoRules) {
			t.Fatal("expected ErrNoRules")
		}

		if !errors.Is(err, cause) {
			t.Fatal("expected cause error")
		}
	})

	t.Run("wraps multiple sentinels", func(t *testing.T) {
		t.Parallel()

		err := ErrFuzzy(ErrVariableNotFound, ErrNoRules)

		if !errors.Is(err, ErrVariableNotFound) {
			t.Fatal("expected ErrVariableNotFound")
		}

		if !errors.Is(err, ErrNoRules) {
			t.Fatal("expected ErrNoRules")
		}
	})

	t.Run("is Error type", func(t *testing.T) {
		t.Parallel()

		err := ErrFuzzy(ErrNoRules)

		var fuzzyErr *Error

		if !errors.As(err, &fuzzyErr) {
			t.Fatal("expected *Error type")
		}

		if fuzzyErr.Type != FuzzyType {
			t.Fatalf("expected type %s, got %s", FuzzyType, fuzzyErr.Type)
		}
	})
}

func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	t.Run("ErrVariableNotFound", func(t *testing.T) {
		t.Parallel()

		if ErrVariableNotFound.Error() != "variable not found" {
			t.Fatalf("unexpected: %s", ErrVariableNotFound.Error())
		}
	})

	t.Run("ErrNoRules", func(t *testing.T) {
		t.Parallel()

		if ErrNoRules.Error() != "no rules provided" {
			t.Fatalf("unexpected: %s", ErrNoRules.Error())
		}
	})
}
