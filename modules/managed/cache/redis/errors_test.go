package redis

import (
	"errors"
	"testing"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

func TestErrCommand(t *testing.T) {
	t.Parallel()

	t.Run("wraps ErrRedisCommandFailed with CacheRedisType tag", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("conn reset")
		err := ErrCommand(cause)

		if !errors.Is(err, ErrRedisCommandFailed) {
			t.Fatalf("expected wrap of ErrRedisCommandFailed, got %v", err)
		}

		if !errors.Is(err, cause) {
			t.Fatalf("expected wrap of cause, got %v", err)
		}

		var e *Error
		ok := errors.As(err, &e)
		if !ok {
			t.Fatalf("expected *Error, got %T", err)
		}

		if e.Type != CacheRedisType {
			t.Fatalf("Type = %q, want %q", e.Type, CacheRedisType)
		}
	})
}

func TestErrEncode(t *testing.T) {
	t.Parallel()

	t.Run("wraps ErrRedisEncodeFailed", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("json marshal failed")
		err := ErrEncode(cause)

		if !errors.Is(err, ErrRedisEncodeFailed) {
			t.Fatalf("expected wrap of ErrRedisEncodeFailed, got %v", err)
		}

		if !errors.Is(err, cause) {
			t.Fatalf("expected wrap of cause, got %v", err)
		}
	})
}

func TestErrDecode(t *testing.T) {
	t.Parallel()

	t.Run("wraps ErrRedisDecodeFailed", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("invalid json")
		err := ErrDecode(cause)

		if !errors.Is(err, ErrRedisDecodeFailed) {
			t.Fatalf("expected wrap of ErrRedisDecodeFailed, got %v", err)
		}

		if !errors.Is(err, cause) {
			t.Fatalf("expected wrap of cause, got %v", err)
		}
	})
}

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("formats type and inner error", func(t *testing.T) {
		t.Parallel()

		e := &Error{
			TypedError: cerrs.TypedError{
				Type: CacheRedisType,
				Err:  errors.New("boom"),
			},
		}

		got := e.Error()
		want := CacheRedisType + " error: boom"

		if got != want {
			t.Fatalf("Error() = %q, want %q", got, want)
		}
	})
}
