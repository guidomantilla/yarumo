package repository

import (
	"errors"
	"testing"
)

func TestErrGet(t *testing.T) {
	t.Parallel()

	t.Run("wraps sentinel", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("test cause")
		err := ErrGet(cause)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, ErrGetFailed) {
			t.Fatal("expected error to wrap ErrGetFailed")
		}

		if !errors.Is(err, cause) {
			t.Fatal("expected error to wrap cause")
		}

		var typed *Error
		ok := errors.As(err, &typed)

		if !ok {
			t.Fatal("expected error to be *Error")
		}

		if typed.Type != RepositoryType {
			t.Fatalf("expected type %s, got %s", RepositoryType, typed.Type)
		}
	})
}

func TestErrList(t *testing.T) {
	t.Parallel()

	t.Run("wraps sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrList(errors.New("bad"))

		if !errors.Is(err, ErrListFailed) {
			t.Fatal("expected error to wrap ErrListFailed")
		}
	})
}

func TestErrSave(t *testing.T) {
	t.Parallel()

	t.Run("wraps sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrSave(errors.New("bad"))

		if !errors.Is(err, ErrSaveFailed) {
			t.Fatal("expected error to wrap ErrSaveFailed")
		}
	})
}
