package authz

import (
	"errors"
	"strings"
	"testing"
)

func TestErrAuthz(t *testing.T) {
	t.Parallel()

	t.Run("wraps cause", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("boom")
		err := ErrAuthz(cause)

		if err == nil {
			t.Fatal("expected non-nil error")
		}

		if !errors.Is(err, ErrAuthzFailed) {
			t.Fatal("expected errors.Is(err, ErrAuthzFailed) to be true")
		}

		if !errors.Is(err, cause) {
			t.Fatal("expected errors.Is(err, cause) to be true")
		}
	})

	t.Run("no causes still wraps sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrAuthz()
		if err == nil {
			t.Fatal("expected non-nil error")
		}

		if !errors.Is(err, ErrAuthzFailed) {
			t.Fatal("expected errors.Is(err, ErrAuthzFailed) to be true")
		}
	})

	t.Run("error string includes type", func(t *testing.T) {
		t.Parallel()

		err := ErrAuthz(errors.New("nope"))

		if !strings.Contains(err.Error(), AuthzType) {
			t.Fatalf("expected error string to include %q, got %q", AuthzType, err.Error())
		}
	})
}

func TestErrAuthz_AsDomainError(t *testing.T) {
	t.Parallel()

	err := ErrAuthz(ErrDenied)

	var domain *Error
	if !errors.As(err, &domain) {
		t.Fatal("expected errors.As(err, *Error) to succeed")
	}

	if domain.Type != AuthzType {
		t.Fatalf("expected type %q, got %q", AuthzType, domain.Type)
	}
}
