package cache

import (
	"errors"
	"testing"
)

func TestErrCache(t *testing.T) {
	t.Parallel()

	t.Run("wraps causes with ErrCacheFailed", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("boom")
		err := ErrCache(cause)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, ErrCacheFailed) {
			t.Fatal("expected wrapped ErrCacheFailed")
		}
		if !errors.Is(err, cause) {
			t.Fatal("expected wrapped cause")
		}
	})

	t.Run("no causes still wraps ErrCacheFailed", func(t *testing.T) {
		t.Parallel()

		err := ErrCache()
		if !errors.Is(err, ErrCacheFailed) {
			t.Fatal("expected wrapped ErrCacheFailed")
		}
	})
}

func TestErrMiss(t *testing.T) {
	t.Parallel()

	err := ErrMiss()
	if !errors.Is(err, ErrCacheMiss) {
		t.Fatal("expected wrapped ErrCacheMiss")
	}
}

func TestErrSerialize(t *testing.T) {
	t.Parallel()

	err := ErrSerialize(errors.New("bad value"))
	if !errors.Is(err, ErrSerialization) {
		t.Fatal("expected wrapped ErrSerialization")
	}
}

func TestErrBackend(t *testing.T) {
	t.Parallel()

	err := ErrBackend(errors.New("conn refused"))
	if !errors.Is(err, ErrBackendUnavailable) {
		t.Fatal("expected wrapped ErrBackendUnavailable")
	}
}

func TestErrUnsupported(t *testing.T) {
	t.Parallel()

	err := ErrUnsupported()
	if !errors.Is(err, ErrUnsupportedBackend) {
		t.Fatal("expected wrapped ErrUnsupportedBackend")
	}
}

func TestError_Unwrap(t *testing.T) {
	t.Parallel()

	cause := errors.New("inner")
	err := ErrCache(cause)
	unwrapped := errors.Unwrap(err)
	if unwrapped == nil {
		t.Fatal("expected non-nil unwrap result")
	}
}
