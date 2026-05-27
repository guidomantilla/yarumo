package hashes

import (
	"errors"
	"strings"
	"testing"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("formats error with type and cause", func(t *testing.T) {
		t.Parallel()

		err := ErrAlgorithmNotSupported("MD5")

		got := err.Error()
		if !strings.Contains(got, "hash") {
			t.Fatalf("expected 'hash' in error, got %q", got)
		}

		if !strings.Contains(got, HashNotFound) {
			t.Fatalf("expected type %q in error, got %q", HashNotFound, got)
		}

		if !strings.Contains(got, "MD5") {
			t.Fatalf("expected algorithm name in error, got %q", got)
		}
	})
}

func TestErrAlgorithmNotSupported(t *testing.T) {
	t.Parallel()

	t.Run("returns domain Error", func(t *testing.T) {
		t.Parallel()

		err := ErrAlgorithmNotSupported("UNKNOWN")

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}
	})

	t.Run("includes algorithm name in message", func(t *testing.T) {
		t.Parallel()

		err := ErrAlgorithmNotSupported("RIPEMD160")

		if !strings.Contains(err.Error(), "RIPEMD160") {
			t.Fatalf("expected algorithm name, got %q", err.Error())
		}
	})

	t.Run("sets type to HashNotFound", func(t *testing.T) {
		t.Parallel()

		err := ErrAlgorithmNotSupported("test")

		var domErr *Error

		ok := errors.As(err, &domErr)
		if !ok {
			t.Fatal("expected *Error")
		}

		if domErr.Type != HashNotFound {
			t.Fatalf("expected type %q, got %q", HashNotFound, domErr.Type)
		}
	})
}
