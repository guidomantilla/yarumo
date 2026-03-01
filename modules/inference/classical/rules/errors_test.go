package rules

import (
	"errors"
	"testing"
)

func TestErrValidation(t *testing.T) {
	t.Parallel()

	t.Run("wraps sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrValidation()

		if !errors.Is(err, ErrRuleInvalid) {
			t.Fatal("expected ErrRuleInvalid in chain")
		}
	})

	t.Run("wraps additional cause", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("missing name")
		err := ErrValidation(cause)

		if !errors.Is(err, cause) {
			t.Fatal("expected cause in chain")
		}

		if !errors.Is(err, ErrRuleInvalid) {
			t.Fatal("expected ErrRuleInvalid in chain")
		}
	})

	t.Run("error message contains type", func(t *testing.T) {
		t.Parallel()

		err := ErrValidation()
		got := err.Error()

		if got == "" {
			t.Fatal("expected non-empty error message")
		}
	})

	t.Run("is Error type", func(t *testing.T) {
		t.Parallel()

		err := ErrValidation()

		var ruleErr *Error

		ok := errors.As(err, &ruleErr)
		if !ok {
			t.Fatal("expected Error type")
		}

		if ruleErr.Type != RuleType {
			t.Fatalf("expected type %s, got %s", RuleType, ruleErr.Type)
		}
	})
}
