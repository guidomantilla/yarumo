package cache

import (
	"errors"
	"testing"

	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

func TestErrCache(t *testing.T) {
	t.Parallel()

	t.Run("wraps ErrCacheFailed without causes", func(t *testing.T) {
		t.Parallel()

		err := ErrCache()
		if err == nil {
			t.Fatal("expected non-nil error")
		}

		if !errors.Is(err, ErrCacheFailed) {
			t.Fatalf("expected wrap of ErrCacheFailed, got %v", err)
		}
	})

	t.Run("wraps additional causes", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("boom")
		err := ErrCache(cause)

		if !errors.Is(err, cause) {
			t.Fatalf("expected wrap of cause, got %v", err)
		}

		if !errors.Is(err, ErrCacheFailed) {
			t.Fatalf("expected wrap of ErrCacheFailed, got %v", err)
		}
	})

	t.Run("returns *Error with CacheType tag", func(t *testing.T) {
		t.Parallel()

		err := ErrCache()

		var e *Error
		ok := errors.As(err, &e)
		if !ok {
			t.Fatalf("expected *Error, got %T", err)
		}

		if e.Type != CacheType {
			t.Fatalf("Type = %q, want %q", e.Type, CacheType)
		}
	})
}

func TestErrTypeAssertion(t *testing.T) {
	t.Parallel()

	t.Run("wraps ErrCacheTypeAssertion without causes", func(t *testing.T) {
		t.Parallel()

		err := ErrTypeAssertion()
		if !errors.Is(err, ErrCacheTypeAssertion) {
			t.Fatalf("expected wrap of ErrCacheTypeAssertion, got %v", err)
		}
	})

	t.Run("wraps additional causes", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("cast failed")
		err := ErrTypeAssertion(cause)

		if !errors.Is(err, cause) {
			t.Fatalf("expected wrap of cause, got %v", err)
		}

		if !errors.Is(err, ErrCacheTypeAssertion) {
			t.Fatalf("expected wrap of ErrCacheTypeAssertion, got %v", err)
		}
	})
}

func TestErrMiss(t *testing.T) {
	t.Parallel()

	t.Run("wraps ErrCacheMiss without causes", func(t *testing.T) {
		t.Parallel()

		err := ErrMiss()
		if !errors.Is(err, ErrCacheMiss) {
			t.Fatalf("expected wrap of ErrCacheMiss, got %v", err)
		}
	})

	t.Run("wraps additional causes", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("absent")
		err := ErrMiss(cause)

		if !errors.Is(err, cause) {
			t.Fatalf("expected wrap of cause, got %v", err)
		}

		if !errors.Is(err, ErrCacheMiss) {
			t.Fatalf("expected wrap of ErrCacheMiss, got %v", err)
		}
	})
}

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("formats type and inner error", func(t *testing.T) {
		t.Parallel()

		e := &Error{
			TypedError: cerrs.TypedError{
				Type: CacheType,
				Err:  errors.New("boom"),
			},
		}

		got := e.Error()
		want := "cache error: boom"

		if got != want {
			t.Fatalf("Error() = %q, want %q", got, want)
		}
	})
}
