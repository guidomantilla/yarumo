package adapters

import (
	"errors"
	"testing"
)

func TestErrAdaptRules(t *testing.T) {
	t.Parallel()

	t.Run("wraps sentinel", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("test cause")
		err := ErrAdaptRules(cause)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, ErrAdaptRulesFailed) {
			t.Fatal("expected error to wrap ErrAdaptRulesFailed")
		}

		if !errors.Is(err, cause) {
			t.Fatal("expected error to wrap cause")
		}

		var typed *Error
		ok := errors.As(err, &typed)

		if !ok {
			t.Fatal("expected error to be *Error")
		}

		if typed.Type != AdapterType {
			t.Fatalf("expected type %s, got %s", AdapterType, typed.Type)
		}
	})
}

func TestErrAdaptNetwork(t *testing.T) {
	t.Parallel()

	t.Run("wraps sentinel", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("test cause")
		err := ErrAdaptNetwork(cause)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, ErrAdaptNetworkFailed) {
			t.Fatal("expected error to wrap ErrAdaptNetworkFailed")
		}

		if !errors.Is(err, cause) {
			t.Fatal("expected error to wrap cause")
		}

		var typed *Error
		ok := errors.As(err, &typed)

		if !ok {
			t.Fatal("expected error to be *Error")
		}

		if typed.Type != AdapterType {
			t.Fatalf("expected type %s, got %s", AdapterType, typed.Type)
		}
	})
}

func TestErrAdaptVariables(t *testing.T) {
	t.Parallel()

	t.Run("wraps sentinel", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("test cause")
		err := ErrAdaptVariables(cause)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, ErrAdaptVariablesFailed) {
			t.Fatal("expected error to wrap ErrAdaptVariablesFailed")
		}

		if !errors.Is(err, cause) {
			t.Fatal("expected error to wrap cause")
		}

		var typed *Error
		ok := errors.As(err, &typed)

		if !ok {
			t.Fatal("expected error to be *Error")
		}

		if typed.Type != AdapterType {
			t.Fatalf("expected type %s, got %s", AdapterType, typed.Type)
		}
	})
}

func TestErrAdaptMembership(t *testing.T) {
	t.Parallel()

	t.Run("wraps sentinel", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("test cause")
		err := ErrAdaptMembership(cause)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, ErrAdaptMembershipFailed) {
			t.Fatal("expected error to wrap ErrAdaptMembershipFailed")
		}

		if !errors.Is(err, cause) {
			t.Fatal("expected error to wrap cause")
		}

		var typed *Error
		ok := errors.As(err, &typed)

		if !ok {
			t.Fatal("expected error to be *Error")
		}

		if typed.Type != AdapterType {
			t.Fatalf("expected type %s, got %s", AdapterType, typed.Type)
		}
	})
}
