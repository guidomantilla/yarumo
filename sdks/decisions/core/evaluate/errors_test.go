package evaluate

import (
	"errors"
	"testing"
)

func TestErrExecute(t *testing.T) {
	t.Parallel()

	t.Run("wraps sentinel", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("test cause")
		err := ErrExecute(cause)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, ErrExecuteFailed) {
			t.Fatal("expected error to wrap ErrExecuteFailed")
		}

		if !errors.Is(err, cause) {
			t.Fatal("expected error to wrap cause")
		}

		var typed *Error
		ok := errors.As(err, &typed)

		if !ok {
			t.Fatal("expected error to be *Error")
		}

		if typed.Type != EvaluateType {
			t.Fatalf("expected type %s, got %s", EvaluateType, typed.Type)
		}
	})
}

func TestErrExplain(t *testing.T) {
	t.Parallel()

	t.Run("wraps sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrExplain(errors.New("bad"))

		if !errors.Is(err, ErrExplainFailed) {
			t.Fatal("expected error to wrap ErrExplainFailed")
		}
	})
}

func TestErrAudit(t *testing.T) {
	t.Parallel()

	t.Run("wraps sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrAudit(errors.New("bad"))

		if !errors.Is(err, ErrAuditFailed) {
			t.Fatal("expected error to wrap ErrAuditFailed")
		}
	})
}

func TestErrCascade(t *testing.T) {
	t.Parallel()

	t.Run("wraps sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrCascade(errors.New("bad"))

		if !errors.Is(err, ErrCascadeFailed) {
			t.Fatal("expected error to wrap ErrCascadeFailed")
		}
	})
}
