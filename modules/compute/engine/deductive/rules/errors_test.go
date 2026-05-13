package rules

import (
	"errors"
	"testing"
)

func TestErrValidation(t *testing.T) {
	t.Parallel()

	t.Run("wraps sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrValidation(ErrRuleInvalid)

		if !errors.Is(err, ErrRuleInvalid) {
			t.Fatal("expected ErrRuleInvalid in chain")
		}
	})

	t.Run("wraps additional cause", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("missing name")
		err := ErrValidation(cause, ErrRuleInvalid)

		if !errors.Is(err, cause) {
			t.Fatal("expected cause in chain")
		}

		if !errors.Is(err, ErrRuleInvalid) {
			t.Fatal("expected ErrRuleInvalid in chain")
		}
	})

	t.Run("error message contains type", func(t *testing.T) {
		t.Parallel()

		err := ErrValidation(ErrRuleInvalid)
		got := err.Error()

		if got == "" {
			t.Fatal("expected non-empty error message")
		}
	})

	t.Run("is Error type", func(t *testing.T) {
		t.Parallel()

		err := ErrValidation(ErrRuleInvalid)

		var ruleErr *Error

		ok := errors.As(err, &ruleErr)
		if !ok {
			t.Fatal("expected Error type")
		}

		if ruleErr.Type != RuleType {
			t.Fatalf("expected type %s, got %s", RuleType, ruleErr.Type)
		}
	})

	t.Run("zero args still wraps ErrRuleValidationFailed", func(t *testing.T) {
		t.Parallel()

		err := ErrValidation()
		if !errors.Is(err, ErrRuleValidationFailed) {
			t.Fatal("expected ErrRuleValidationFailed in chain")
		}
	})
}

func TestErrRuleValidationFailed(t *testing.T) {
	t.Parallel()

	if ErrRuleValidationFailed == nil {
		t.Fatal("expected non-nil error")
	}

	if ErrRuleValidationFailed.Error() != "rule validation failed" {
		t.Fatalf("unexpected message: %s", ErrRuleValidationFailed.Error())
	}
}
