package rbac

import (
	"errors"
	"strings"
	"testing"
)

func TestErrRBAC(t *testing.T) {
	t.Parallel()

	t.Run("wraps cause", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("boom")
		err := ErrRBAC(cause)

		if err == nil {
			t.Fatal("expected non-nil error")
		}

		if !errors.Is(err, ErrRBACFailed) {
			t.Fatal("expected errors.Is(err, ErrRBACFailed) to be true")
		}

		if !errors.Is(err, cause) {
			t.Fatal("expected errors.Is(err, cause) to be true")
		}
	})

	t.Run("error string includes type", func(t *testing.T) {
		t.Parallel()

		err := ErrRBAC(errors.New("x"))

		if !strings.Contains(err.Error(), RBACType) {
			t.Fatalf("expected error string to include %q, got %q", RBACType, err.Error())
		}
	})
}

func TestErrRBAC_AsDomainError(t *testing.T) {
	t.Parallel()

	err := ErrRBAC(ErrInheritanceCycle)

	var domain *Error
	if !errors.As(err, &domain) {
		t.Fatal("expected errors.As(err, *Error) to succeed")
	}

	if domain.Type != RBACType {
		t.Fatalf("expected type %q, got %q", RBACType, domain.Type)
	}
}
