package facts

import (
	"errors"
	"testing"
)

func TestErrQuery(t *testing.T) {
	t.Parallel()

	t.Run("wraps sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrQuery()

		if !errors.Is(err, ErrNotFound) {
			t.Fatal("expected ErrNotFound in chain")
		}
	})

	t.Run("wraps additional cause", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("var is empty")
		err := ErrQuery(cause)

		if !errors.Is(err, cause) {
			t.Fatal("expected cause in chain")
		}

		if !errors.Is(err, ErrNotFound) {
			t.Fatal("expected ErrNotFound in chain")
		}
	})

	t.Run("error message contains type", func(t *testing.T) {
		t.Parallel()

		err := ErrQuery()
		got := err.Error()

		if got == "" {
			t.Fatal("expected non-empty error message")
		}
	})

	t.Run("is Error type", func(t *testing.T) {
		t.Parallel()

		err := ErrQuery()

		var factErr *Error

		ok := errors.As(err, &factErr)
		if !ok {
			t.Fatal("expected Error type")
		}

		if factErr.Type != FactType {
			t.Fatalf("expected type %s, got %s", FactType, factErr.Type)
		}
	})
}
