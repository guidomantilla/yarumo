package ristretto

import (
	"errors"
	"testing"

	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

func TestErrInit(t *testing.T) {
	t.Parallel()

	t.Run("wraps ErrRistrettoInitFailed without causes", func(t *testing.T) {
		t.Parallel()

		err := ErrInit()
		if !errors.Is(err, ErrRistrettoInitFailed) {
			t.Fatalf("expected wrap of ErrRistrettoInitFailed, got %v", err)
		}
	})

	t.Run("wraps additional causes", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("boom")
		err := ErrInit(cause)

		if !errors.Is(err, cause) {
			t.Fatalf("expected wrap of cause, got %v", err)
		}

		if !errors.Is(err, ErrRistrettoInitFailed) {
			t.Fatalf("expected wrap of ErrRistrettoInitFailed, got %v", err)
		}
	})

	t.Run("returns *Error with CacheRistrettoType tag", func(t *testing.T) {
		t.Parallel()

		err := ErrInit()

		var e *Error
		ok := errors.As(err, &e)
		if !ok {
			t.Fatalf("expected *Error, got %T", err)
		}

		if e.Type != CacheRistrettoType {
			t.Fatalf("Type = %q, want %q", e.Type, CacheRistrettoType)
		}
	})
}

func TestErrSet(t *testing.T) {
	t.Parallel()

	t.Run("wraps ErrRistrettoSetRejected without causes", func(t *testing.T) {
		t.Parallel()

		err := ErrSet()
		if !errors.Is(err, ErrRistrettoSetRejected) {
			t.Fatalf("expected wrap of ErrRistrettoSetRejected, got %v", err)
		}
	})

	t.Run("wraps additional causes", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("admission rejected")
		err := ErrSet(cause)

		if !errors.Is(err, cause) {
			t.Fatalf("expected wrap of cause, got %v", err)
		}
	})

	t.Run("returns *Error with CacheRistrettoType tag", func(t *testing.T) {
		t.Parallel()

		err := ErrSet()

		var e *Error
		ok := errors.As(err, &e)
		if !ok {
			t.Fatalf("expected *Error, got %T", err)
		}

		if e.Type != CacheRistrettoType {
			t.Fatalf("Type = %q, want %q", e.Type, CacheRistrettoType)
		}
	})
}

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("formats type and inner error", func(t *testing.T) {
		t.Parallel()

		e := &Error{
			TypedError: cerrs.TypedError{
				Type: CacheRistrettoType,
				Err:  errors.New("boom"),
			},
		}

		got := e.Error()
		want := CacheRistrettoType + " error: boom"

		if got != want {
			t.Fatalf("Error() = %q, want %q", got, want)
		}
	})
}
