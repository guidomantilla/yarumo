package facts

import (
	"errors"
	"testing"
)

func TestErrQuery(t *testing.T) {
	t.Parallel()

	t.Run("wraps sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrQuery(ErrNotFound)

		if !errors.Is(err, ErrNotFound) {
			t.Fatal("expected ErrNotFound in chain")
		}
	})

	t.Run("wraps additional cause", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("var is empty")
		err := ErrQuery(cause, ErrNotFound)

		if !errors.Is(err, cause) {
			t.Fatal("expected cause in chain")
		}

		if !errors.Is(err, ErrNotFound) {
			t.Fatal("expected ErrNotFound in chain")
		}
	})

	t.Run("error message contains type", func(t *testing.T) {
		t.Parallel()

		err := ErrQuery(ErrNotFound)
		got := err.Error()

		if got == "" {
			t.Fatal("expected non-empty error message")
		}
	})

	t.Run("is Error type", func(t *testing.T) {
		t.Parallel()

		err := ErrQuery(ErrNotFound)

		var factErr *Error

		ok := errors.As(err, &factErr)
		if !ok {
			t.Fatal("expected Error type")
		}

		if factErr.Type != FactType {
			t.Fatalf("expected type %s, got %s", FactType, factErr.Type)
		}
	})

	t.Run("zero args still wraps ErrFactQueryFailed", func(t *testing.T) {
		t.Parallel()

		err := ErrQuery()
		if !errors.Is(err, ErrFactQueryFailed) {
			t.Fatal("expected ErrFactQueryFailed in chain")
		}
	})
}

func TestErrFactQueryFailed(t *testing.T) {
	t.Parallel()

	if ErrFactQueryFailed == nil {
		t.Fatal("expected non-nil error")
	}

	if ErrFactQueryFailed.Error() != "fact query failed" {
		t.Fatalf("unexpected message: %s", ErrFactQueryFailed.Error())
	}
}
